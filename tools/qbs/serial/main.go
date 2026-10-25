//go:build tinygo

package main

import (
	"machine"
	"strconv"
	"time"

	"github.com/tinygo-org/tinygo/tools/qbs/dev"
)

// extraTests holds tests that not every target can build.
var extraTests []dev.Test

func main() {
	tests := []dev.Test{
		{Name: "echo", Run: echo},
		{Name: "burst", Run: burst},
		{Name: "seq", Run: seq},
		{Name: "bytes", Run: bytes},
	}
	dev.Main("serial", append(tests, extraTests...))
}

// echoLen is the number of bytes the host sends to the echo test.
const echoLen = 666

// echo sends back every byte it receives.
func echo(t *dev.T) {
	n := 0
	last := time.Now()
	for n < echoLen && time.Since(last) < 3*time.Second {
		if machine.Serial.Buffered() == 0 {
			time.Sleep(time.Millisecond)
			continue
		}
		b, _ := machine.Serial.ReadByte()
		machine.Serial.WriteByte(b)
		n++
		last = time.Now()
	}
	t.Val("echoed", strconv.Itoa(n))
}

// burst does not read while the host sends more than the RX buffer holds.
func burst(t *dev.T) {
	time.Sleep(1500 * time.Millisecond)
	var got []byte
	for machine.Serial.Buffered() > 0 {
		b, _ := machine.Serial.ReadByte()
		got = append(got, b)
	}
	t.Val("buf", strconv.Itoa(len(got)))
	t.Val("data", dev.Hex(got))

	// The port must keep receiving after the overrun.
	t.Val("rec", dev.Hex(dev.ReadN(16, 5*time.Second)))
}

// seq prints numbered lines as fast as it can.
func seq(t *dev.T) {
	for i := 0; i < 2000; i++ {
		println("S", i, i*7919%10007)
	}
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

// bytes sends every byte value in four patterns.
func bytes(t *dev.T) {
	for k := 0; k < 4; k++ {
		t.Raw("p"+strconv.Itoa(k), bytePattern(k))
	}
}
