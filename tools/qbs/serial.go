package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var serialDrivers = map[string]driver{
	"echo":  driveEcho,
	"burst": driveBurst,
	"conc":  driveConc,
}

// driveEcho sends 666 bytes in three phases and checks that they come back.
func driveEcho(l *link, t *testResult) {
	rng := rand.New(rand.NewSource(1))
	random := func(n int) []byte {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte(rng.Intn(256))
		}
		return b
	}
	phase := func(label string, chunks [][]byte, gap time.Duration) string {
		var sent []byte
		for _, c := range chunks {
			l.port.Write(c)
			sent = append(sent, c...)
			time.Sleep(gap)
		}
		got := l.readExact(len(sent), 5*time.Second)
		if !bytes.Equal(got, sent) {
			n := 0
			for n < len(got) && n < len(sent) && got[n] == sent[n] {
				n++
			}
			return fmt.Sprintf("%s: sent %d got %d bytes, first difference at %d", label, len(sent), len(got), n)
		}
		return ""
	}

	var lines [][]byte
	for i := 0; i < 10; i++ {
		lines = append(lines, []byte(fmt.Sprintf("hello %d\r\n", i)))
	}
	var paced [][]byte
	for i := 0; i < 16; i++ {
		paced = append(paced, random(32))
	}
	for _, p := range []struct {
		label  string
		chunks [][]byte
		gap    time.Duration
	}{
		{"small", lines, 20 * time.Millisecond},
		{"64B block", [][]byte{random(64)}, 50 * time.Millisecond},
		{"512B paced", paced, 10 * time.Millisecond},
	} {
		if msg := phase(p.label, p.chunks, p.gap); msg != "" {
			t.hostFail = msg
			return
		}
	}
}

const burstRecovery = "RECOVERY-OK-0123"

// driveBurst sends more than the RX buffer holds while the device is not
// reading, then sends a short message after the device has read the buffer.
func driveBurst(l *link, t *testResult) {
	pat := make([]byte, 4096)
	for i := range pat {
		pat[i] = byte(i % 251)
	}
	l.port.Write(pat)
	time.Sleep(2500 * time.Millisecond)
	l.port.Write([]byte(burstRecovery))
}

// driveConc sends 512 bytes while the device prints from goroutines.
func driveConc(l *link, t *testResult) {
	payload := make([]byte, 512)
	for i := range payload {
		payload[i] = byte(i)
	}
	for i := 0; i < len(payload); i += 16 {
		l.port.Write(payload[i : i+16])
		time.Sleep(10 * time.Millisecond)
	}
}

func evalSerial(r *classResult) {
	for _, t := range r.tests {
		if t.status != "PASS" {
			continue
		}
		var msg, ok string
		switch t.name {
		case "echo":
			msg, ok = checkEcho(t)
		case "burst":
			msg, ok = checkBurst(t)
		case "seq":
			msg, ok = checkSeq(t)
		case "bytes":
			msg, ok = checkBytes(t)
		case "conc":
			msg, ok = checkConc(t)
		default:
			continue
		}
		if msg != "" {
			t.fail("%s", msg)
		} else {
			t.detail = ok
		}
	}
}

func checkEcho(t *testResult) (string, string) {
	if t.hostFail != "" {
		return t.hostFail, ""
	}
	if v, _ := t.val("echoed"); v != "666" {
		return "device echoed " + v + " of 666 bytes", ""
	}
	return "", "echoed 10 lines, 64 bytes and 512 paced random bytes"
}

func checkBurst(t *testResult) (string, string) {
	nstr, _ := t.val("buf")
	n, err := strconv.Atoi(nstr)
	if err != nil {
		return "no buffer size from the device", ""
	}
	data, _ := t.val("data")
	got, err := hex.DecodeString(data)
	if err != nil || len(got) != n {
		return "device data does not match its buffer size", ""
	}
	if n == 0 {
		return "no bytes received at all", ""
	}
	note := fmt.Sprintf("RX buffer held %d of 4096 bytes", n)
	for i, b := range got {
		if b != byte(i%251) {
			return fmt.Sprintf("%s, but the data is not a prefix of the burst (first bad index %d)", note, i), ""
		}
	}
	rec, _ := t.val("rec")
	recBytes, _ := hex.DecodeString(rec)
	if string(recBytes) != burstRecovery {
		return fmt.Sprintf("receiving stopped after the overrun, got %q", recBytes), ""
	}
	return "", note + ", receive works after the overrun"
}

func checkSeq(t *testResult) (string, string) {
	var lines []string
	for _, s := range t.stray {
		if strings.HasPrefix(s, "S ") {
			lines = append(lines, s)
		}
	}
	if len(lines) != 2000 {
		return fmt.Sprintf("got %d of 2000 lines", len(lines)), ""
	}
	for i, s := range lines {
		want := fmt.Sprintf("S %d %d", i, i*7919%10007)
		if s != want {
			return fmt.Sprintf("line %d: got %q want %q", i, s, want), ""
		}
	}
	return "", "2000 lines in order"
}

func checkBytes(t *testResult) (string, string) {
	for k := 0; k < 4; k++ {
		want := make([]byte, 256)
		for i := range want {
			switch k {
			case 0:
				want[i] = byte(i)
			case 1:
				want[i] = byte(255 - i)
			case 2:
				want[i] = 0
			default:
				if i%2 == 0 {
					want[i] = 0xAA
				} else {
					want[i] = 0x55
				}
			}
		}
		got := t.raw["p"+strconv.Itoa(k)]
		if !bytes.Equal(got, want) {
			return fmt.Sprintf("pattern %d differs (got %d bytes)", k, len(got)), ""
		}
	}
	return "", "4 patterns of 256 bytes, all byte values intact"
}

var concLine = regexp.MustCompile(`^G (\d) (\d+)$`)

func checkConc(t *testResult) (string, string) {
	last := [3]int{-1, -1, -1}
	count := 0
	for _, s := range t.stray {
		if s == "" {
			continue
		}
		m := concLine.FindStringSubmatch(s)
		if m == nil {
			return fmt.Sprintf("torn or garbled line %q", s), ""
		}
		gid, _ := strconv.Atoi(m[1])
		k, _ := strconv.Atoi(m[2])
		if gid > 2 || k != last[gid]+1 {
			return fmt.Sprintf("goroutine %d out of order: %d after %d", gid, k, last[gid]), ""
		}
		last[gid] = k
		count++
	}
	if count != 600 || last != [3]int{199, 199, 199} {
		return fmt.Sprintf("saw %d goroutine lines, last %v", count, last), ""
	}
	if v, _ := t.val("rx"); v != "512" {
		return "device counted " + v + " of 512 RX bytes", ""
	}
	return "", "600 interleaved lines intact, 512/512 RX bytes counted"
}
