//go:build tinygo && wasip1 && !qbs_noinput

package dev

import (
	"os"
	"unsafe"
)

// Serial is the standard input and output of the program. The runner talks to
// the emulated program through them.
var Serial serial

type serial struct{}

// rx holds bytes read from standard input that have not been returned yet.
var rx []byte

// Configure does nothing, there is no baud rate.
func (serial) Configure(baud uint32) {}

// Buffered returns the number of bytes that can be read without blocking.
func (serial) Buffered() int {
	if len(rx) == 0 && stdinReady() {
		var buf [64]byte
		n, _ := os.Stdin.Read(buf[:])
		rx = append(rx, buf[:n]...)
	}
	return len(rx)
}

func (s serial) ReadByte() (byte, error) {
	if s.Buffered() == 0 {
		return 0, os.ErrDeadlineExceeded
	}
	b := rx[0]
	rx = rx[1:]
	return b, nil
}

func (serial) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

func (s serial) WriteByte(c byte) error {
	_, err := s.Write([]byte{c})
	return err
}

// The layouts follow the WASI preview 1 subscription and event structs.
type subscription struct {
	userdata  uint64
	tag       uint8
	_         [7]byte
	fdOrClock uint32
	_         uint32
	timeout   uint64
	precision uint64
	flags     uint16
	_         [6]byte
}

type event struct {
	userdata uint64
	errno    uint16
	typ      uint8
	_        [5]byte
	nbytes   uint64
	flags    uint16
	_        [6]byte
}

const (
	eventClock  = 0
	eventFDRead = 1
	clockMono   = 1
)

//go:wasmimport wasi_snapshot_preview1 poll_oneoff
func pollOneoff(in unsafe.Pointer, out unsafe.Pointer, nsubscriptions uint32, nevents unsafe.Pointer) uint32

var (
	pollSubs   [2]subscription
	pollEvents [2]event
)

// stdinReady reports whether reading standard input would not block.
func stdinReady() bool {
	pollSubs[0] = subscription{tag: eventFDRead, fdOrClock: 0}
	// With a zero timeout the clock is already due and poll_oneoff may return
	// only that event, so give input a short time to be reported.
	pollSubs[1] = subscription{tag: eventClock, fdOrClock: clockMono, timeout: 200_000}
	var n uint32
	if pollOneoff(unsafe.Pointer(&pollSubs[0]), unsafe.Pointer(&pollEvents[0]), 2, unsafe.Pointer(&n)) != 0 {
		return false
	}
	for i := uint32(0); i < n; i++ {
		if pollEvents[i].typ == eventFDRead && pollEvents[i].errno == 0 {
			return true
		}
	}
	return false
}
