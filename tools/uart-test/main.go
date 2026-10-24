// Flashes the serial test programs in the subdirectories onto a connected
// board and checks their output over the board's serial port.
//
// Each program prints "@@READY <name>" until the host sends 'G', then runs
// and prints "@@" marker lines that this tool validates.
//
//	go run ./tools/uart-test -target arduino-uno-r4-wifi
//	go run ./tools/uart-test -target pico -port /dev/ttyACM0 echo seq
package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

type link struct {
	port serial.Port
	buf  []byte
	log  []byte
}

func (l *link) fill() {
	var tmp [256]byte
	n, _ := l.port.Read(tmp[:])
	if n > 0 {
		l.buf = append(l.buf, tmp[:n]...)
		l.log = append(l.log, tmp[:n]...)
	}
}

// readUntil returns the bytes up to and including marker.
func (l *link) readUntil(marker []byte, timeout time.Duration) ([]byte, bool) {
	deadline := time.Now().Add(timeout)
	for {
		if i := bytes.Index(l.buf, marker); i >= 0 {
			out := append([]byte(nil), l.buf[:i+len(marker)]...)
			l.buf = l.buf[i+len(marker):]
			return out, true
		}
		if time.Now().After(deadline) {
			return nil, false
		}
		l.fill()
	}
}

// readExact returns n bytes, or what arrived before the timeout.
func (l *link) readExact(n int, timeout time.Duration) []byte {
	deadline := time.Now().Add(timeout)
	for len(l.buf) < n && time.Now().Before(deadline) {
		l.fill()
	}
	if len(l.buf) < n {
		n = len(l.buf)
	}
	out := append([]byte(nil), l.buf[:n]...)
	l.buf = l.buf[n:]
	return out
}

// waitLine returns the text after prefix on the next line that starts with it.
func (l *link) waitLine(prefix string, timeout time.Duration) (string, bool) {
	deadline := time.Now().Add(timeout)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return "", false
		}
		got, ok := l.readUntil([]byte("\n"), left)
		if !ok {
			return "", false
		}
		line := strings.TrimRight(string(got), "\r\n")
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(line[len(prefix):]), true
		}
	}
}

func (l *link) write(b []byte) error {
	_, err := l.port.Write(b)
	return err
}

// sync waits for the READY banner, sends 'G' and waits for the GO line.
func sync(l *link, name string) error {
	if _, ok := l.waitLine("@@READY "+name, 25*time.Second); !ok {
		return fmt.Errorf("no '@@READY %s' banner, flash worked but no serial output", name)
	}
	if err := l.write([]byte("G")); err != nil {
		return err
	}
	if _, ok := l.waitLine("@@GO "+name, 5*time.Second); !ok {
		return fmt.Errorf("device did not answer 'G', RX path not working")
	}
	return nil
}

func testEcho(l *link) (string, error) {
	if err := sync(l, "echo"); err != nil {
		return "", err
	}
	rng := rand.New(rand.NewSource(1))
	random := func(n int) []byte {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte(rng.Intn(256))
		}
		return b
	}
	phase := func(label string, chunks [][]byte, gap time.Duration) error {
		var sent []byte
		for _, c := range chunks {
			if err := l.write(c); err != nil {
				return err
			}
			sent = append(sent, c...)
			time.Sleep(gap)
		}
		got := l.readExact(len(sent), 5*time.Second)
		if !bytes.Equal(got, sent) {
			n := 0
			for n < len(got) && n < len(sent) && got[n] == sent[n] {
				n++
			}
			return fmt.Errorf("%s: sent %d got %d bytes, first difference at %d", label, len(sent), len(got), n)
		}
		return nil
	}

	var lines [][]byte
	for i := 0; i < 10; i++ {
		lines = append(lines, []byte(fmt.Sprintf("hello %d\r\n", i)))
	}
	if err := phase("small", lines, 20*time.Millisecond); err != nil {
		return "", err
	}
	if err := phase("64B block", [][]byte{random(64)}, 50*time.Millisecond); err != nil {
		return "", err
	}
	var paced [][]byte
	for i := 0; i < 16; i++ {
		paced = append(paced, random(32))
	}
	if err := phase("512B paced", paced, 10*time.Millisecond); err != nil {
		return "", err
	}
	return "echoed 10 lines, 64 bytes and 512 paced random bytes", nil
}

func testBurst(l *link) (string, error) {
	if err := sync(l, "burst"); err != nil {
		return "", err
	}
	pat := make([]byte, 4096)
	for i := range pat {
		pat[i] = byte(i % 251)
	}
	if err := l.write(pat); err != nil {
		return "", err
	}
	nstr, ok := l.waitLine("@@BUF", 10*time.Second)
	if !ok {
		return "", fmt.Errorf("no @@BUF line after sending 4096 bytes")
	}
	n, err := strconv.Atoi(nstr)
	if err != nil {
		return "", fmt.Errorf("bad @@BUF line %q", nstr)
	}
	data, ok := l.waitLine("@@DATA", 5*time.Second)
	if !ok {
		return "", fmt.Errorf("no @@DATA line")
	}
	got, err := hex.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("bad @@DATA line: %v", err)
	}
	if len(got) != n {
		return "", fmt.Errorf("device says %d bytes but sent %d", n, len(got))
	}
	if n == 0 {
		return "", fmt.Errorf("no bytes received at all")
	}
	note := fmt.Sprintf("RX buffer held %d of 4096 bytes", n)
	if !bytes.Equal(got, pat[:n]) {
		bad := 0
		for got[bad] == pat[bad] {
			bad++
		}
		return "", fmt.Errorf("%s, but data is not a prefix of the burst (first bad index %d)", note, bad)
	}
	if _, ok := l.waitLine("@@RECOVER", 10*time.Second); !ok {
		return "", fmt.Errorf("no @@RECOVER line")
	}
	const recovery = "RECOVERY-OK-0123"
	if err := l.write([]byte(recovery)); err != nil {
		return "", err
	}
	rec, ok := l.waitLine("@@REC", 10*time.Second)
	if !ok {
		return "", fmt.Errorf("no @@REC line, RX stopped after the overrun")
	}
	recBytes, _ := hex.DecodeString(rec)
	if string(recBytes) != recovery {
		return "", fmt.Errorf("after overrun got %q", recBytes)
	}
	return note + ", receive works after the overrun", nil
}

func testSeq(l *link) (string, error) {
	if err := sync(l, "seq"); err != nil {
		return "", err
	}
	start := time.Now()
	for i := 0; i < 2000; i++ {
		line, ok := l.waitLine("S ", 10*time.Second)
		if !ok {
			return "", fmt.Errorf("timeout waiting for line %d", i)
		}
		want := fmt.Sprintf("%d %d", i, i*7919%10007)
		if line != want {
			return "", fmt.Errorf("line %d: got %q want %q", i, line, want)
		}
	}
	if _, ok := l.waitLine("@@END seq", 5*time.Second); !ok {
		return "", fmt.Errorf("no @@END")
	}
	return fmt.Sprintf("2000 lines in order in %.1fs", time.Since(start).Seconds()), nil
}

func bytePattern(k int) []byte {
	b := make([]byte, 256)
	for i := range b {
		switch k {
		case 0:
			b[i] = byte(i)
		case 1:
			b[i] = byte(255 - i)
		case 2:
			b[i] = 0x00
		default:
			if i%2 == 0 {
				b[i] = 0xAA
			} else {
				b[i] = 0x55
			}
		}
	}
	return b
}

func testBytes(l *link) (string, error) {
	if err := sync(l, "bytes"); err != nil {
		return "", err
	}
	for k := 0; k < 4; k++ {
		if _, ok := l.readUntil([]byte(fmt.Sprintf("@@BIN %d\r\n", k)), 10*time.Second); !ok {
			return "", fmt.Errorf("no header for pattern %d", k)
		}
		got := l.readExact(256, 5*time.Second)
		if !bytes.Equal(got, bytePattern(k)) {
			return "", fmt.Errorf("pattern %d differs (got %d bytes)", k, len(got))
		}
	}
	if _, ok := l.waitLine("@@END bytes", 5*time.Second); !ok {
		return "", fmt.Errorf("no @@END")
	}
	return "4 patterns of 256 bytes, all byte values intact", nil
}

var concLine = regexp.MustCompile(`^G (\d) (\d+)$`)

func testConc(l *link) (string, error) {
	if err := sync(l, "conc"); err != nil {
		return "", err
	}
	payload := make([]byte, 512)
	for i := range payload {
		payload[i] = byte(i)
	}
	for i := 0; i < len(payload); i += 16 {
		if err := l.write(payload[i : i+16]); err != nil {
			return "", err
		}
		time.Sleep(10 * time.Millisecond)
	}
	last := [3]int{-1, -1, -1}
	count := 0
	rx := -1
	for rx < 0 {
		line, ok := l.waitLine("", 15*time.Second)
		if !ok {
			return "", fmt.Errorf("timeout, saw %d goroutine lines", count)
		}
		if strings.HasPrefix(line, "@@RXCOUNT") {
			rx, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "@@RXCOUNT")))
			break
		}
		m := concLine.FindStringSubmatch(line)
		if m == nil {
			return "", fmt.Errorf("torn or garbled line %q", line)
		}
		gid, _ := strconv.Atoi(m[1])
		k, _ := strconv.Atoi(m[2])
		if gid > 2 || k != last[gid]+1 {
			return "", fmt.Errorf("goroutine %d out of order: %d after %d", gid, k, last[gid])
		}
		last[gid] = k
		count++
	}
	if count != 600 || last != [3]int{199, 199, 199} {
		return "", fmt.Errorf("saw %d goroutine lines, last %v", count, last)
	}
	if rx != 512 {
		return "", fmt.Errorf("device counted %d of 512 RX bytes", rx)
	}
	return "600 interleaved lines intact, 512/512 RX bytes counted", nil
}

var tests = []struct {
	name string
	run  func(*link) (string, error)
}{
	{"echo", testEcho},
	{"burst", testBurst},
	{"seq", testSeq},
	{"bytes", testBytes},
	{"conc", testConc},
}

func findPort() string {
	for _, pattern := range []string{"/dev/cu.usbmodem*", "/dev/ttyACM*", "/dev/cu.usbserial*", "/dev/ttyUSB*"} {
		if ports, _ := filepath.Glob(pattern); len(ports) > 0 {
			return ports[0]
		}
	}
	return ""
}

func flash(tinygo, target, port, pathPrepend, dir string) error {
	cmd := exec.Command(tinygo, "flash", "-target", target, "-port", port, dir)
	cmd.Dir = dir
	if pathPrepend != "" {
		cmd.Env = append(os.Environ(), "PATH="+pathPrepend+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 1500 {
			out = out[len(out)-1500:]
		}
		return fmt.Errorf("flash failed: %v\n%s", err, out)
	}
	return nil
}

func openPort(name string) (*link, error) {
	deadline := time.Now().Add(15 * time.Second)
	var err error
	for time.Now().Before(deadline) {
		var p serial.Port
		p, err = serial.Open(name, &serial.Mode{BaudRate: 115200})
		if err == nil {
			p.SetReadTimeout(50 * time.Millisecond)
			return &link{port: p}, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("cannot open %s: %v", name, err)
}

func main() {
	target := flag.String("target", "", "tinygo target (required)")
	tinygo := flag.String("tinygo", "tinygo", "tinygo binary to use")
	port := flag.String("port", "", "serial port (default: first USB serial port found)")
	pathPrepend := flag.String("path-prepend", "", "directory to put first on PATH when flashing")
	dump := flag.Bool("dump", false, "print raw serial data of failed tests")
	flag.Parse()

	if *target == "" {
		fmt.Fprintln(os.Stderr, "missing -target")
		os.Exit(2)
	}
	if *port == "" {
		*port = findPort()
	}
	if *port == "" {
		fmt.Fprintln(os.Stderr, "no serial port found, use -port")
		os.Exit(2)
	}
	want := map[string]bool{}
	for _, name := range flag.Args() {
		want[name] = true
	}

	_, file, _, _ := runtime.Caller(0)
	here := filepath.Dir(file)

	fmt.Printf("port %s, target %s\n", *port, *target)
	type row struct{ name, status, msg string }
	var rows []row
	failed := false
	ran := 0
	for _, t := range tests {
		if len(want) > 0 && !want[t.name] {
			continue
		}
		ran++
		fmt.Printf("\n== %s\n", t.name)
		var l *link
		var msg string
		err := flash(*tinygo, *target, *port, *pathPrepend, filepath.Join(here, t.name))
		if err == nil {
			time.Sleep(time.Second)
			l, err = openPort(*port)
		}
		if err == nil {
			msg, err = t.run(l)
		}
		status := "PASS"
		if err != nil {
			status, msg = "FAIL", err.Error()
			failed = true
			if *dump && l != nil {
				tail := l.log
				if len(tail) > 600 {
					tail = tail[len(tail)-600:]
				}
				fmt.Printf("raw serial data: %q\n", tail)
			}
		}
		if l != nil {
			l.port.Close()
		}
		fmt.Printf("%s: %s\n", status, msg)
		rows = append(rows, row{t.name, status, msg})
	}
	if ran == 0 {
		fmt.Fprintln(os.Stderr, "no matching tests")
		os.Exit(2)
	}

	fmt.Println("\n== summary")
	for _, r := range rows {
		fmt.Printf("%-8s %-5s %s\n", r.name, r.status, r.msg)
	}
	if failed {
		os.Exit(1)
	}
}
