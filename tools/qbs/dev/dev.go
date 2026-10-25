//go:build tinygo

// Package dev is the device side of qbs. A program lists its tests and calls
// Main, which waits for the host, runs the tests and reports over serial.
package dev

import (
	"machine"
	"time"
)

// Test is one named check.
type Test struct {
	Name string
	Run  func(t *T)
}

// T collects the outcome of one test.
type T struct {
	failed  bool
	skipped bool
	detail  string
}

// Fail marks the test as failed. Only the first message is kept.
func (t *T) Fail(msg string) {
	if !t.failed {
		t.failed = true
		t.detail = msg
	}
}

// Skip marks the test as skipped.
func (t *T) Skip(msg string) {
	t.skipped = true
	t.detail = msg
}

// Log sends a message to the host.
func (t *T) Log(msg string) {
	println("@@LOG", msg)
}

// Val sends a named value to the host.
func (t *T) Val(key, value string) {
	println("@@VAL", key, value)
}

// Raw sends bytes to the host unchanged. The host reads exactly len(data)
// bytes after the header line.
func (t *T) Raw(key string, data []byte) {
	println("@@RAW", key, len(data))
	machine.Serial.Write(data)
	println()
}

const hexDigits = "0123456789abcdef"

// Hex returns the lowercase hex encoding of b.
func Hex(b []byte) string {
	out := make([]byte, 0, len(b)*2)
	for _, c := range b {
		out = append(out, hexDigits[c>>4], hexDigits[c&15])
	}
	return string(out)
}

// ReadN reads up to n bytes from the serial port, giving up after timeout.
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

// Main runs the tests of one class and never returns.
func Main(class string, tests []Test) {
	machine.Serial.Configure(machine.UARTConfig{BaudRate: 115200})
	waitGo(class)
	println("@@GO", class)
	for i := range tests {
		println("@@BEGIN", tests[i].Name)
		t := &T{}
		start := time.Now()
		tests[i].Run(t)
		us := time.Since(start).Microseconds()
		status := "PASS"
		switch {
		case t.skipped:
			status = "SKIP"
		case t.failed:
			status = "FAIL"
		}
		println("@@RESULT", tests[i].Name, status, us, t.detail)
	}
	for {
		println("@@END", class)
		time.Sleep(time.Second)
	}
}

func waitGo(class string) {
	for {
		for machine.Serial.Buffered() > 0 {
			b, _ := machine.Serial.ReadByte()
			if b == 'G' {
				return
			}
		}
		println("@@READY qbs", class)
		time.Sleep(300 * time.Millisecond)
	}
}
