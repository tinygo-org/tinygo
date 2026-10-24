//go:build tinygo

// Package common has helpers shared by the UART test programs.
package common

import (
	"machine"
	"time"
)

const hexDigits = "0123456789abcdef"

// WaitGo announces the test and blocks until the host sends 'G'.
func WaitGo(name string) {
	for {
		for machine.Serial.Buffered() > 0 {
			b, _ := machine.Serial.ReadByte()
			if b == 'G' {
				return
			}
		}
		println("@@READY", name)
		time.Sleep(300 * time.Millisecond)
	}
}

// Hex returns the lowercase hex encoding of b.
func Hex(b []byte) string {
	out := make([]byte, 0, len(b)*2)
	for _, c := range b {
		out = append(out, hexDigits[c>>4], hexDigits[c&15])
	}
	return string(out)
}

// ReadN reads up to n bytes from the default serial, giving up after timeout.
func ReadN(n int, timeout time.Duration) []byte {
	var got []byte
	deadline := time.Now().Add(timeout)
	for len(got) < n && time.Now().Before(deadline) {
		if machine.Serial.Buffered() == 0 {
			time.Sleep(time.Millisecond)
			continue
		}
		b, _ := machine.Serial.ReadByte()
		got = append(got, b)
	}
	return got
}
