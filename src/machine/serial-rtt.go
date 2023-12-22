//go:build baremetal && serial.rtt

// Implement Segger RTT support.
// This is mostly useful for targets that only have a debug connection
// available, and no serial output (or input). It is somewhat like semihosting,
// but not unusably slow.
// It was originally specified by Segger, but support is available in OpenOCD
// for at least the DAPLink debuggers so I assume it works on any SWD debugger.

package machine

import (
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// This symbol name is known by the compiler, see monitor.go.
var rttSerialInstance rttSerial

var Serial = &rttSerialInstance

func InitSerial() {
	Serial.Configure(UARTConfig{})
}

const (
	// Some constants, see:
	// https://github.com/SEGGERMicro/RTT/blob/master/RTT/SEGGER_RTT.h

	rttMaxNumUpBuffers   = 1
	rttMaxNumDownBuffers = 1
	rttBufferSizeUp      = 1024
	rttBufferSizeDown    = 16

	rttModeNoBlockSkip     = 0
	rttModeNoBlockTrim     = 1
	rttModeBlockIfFifoFull = 2
)

// The debugger knows about the layout of this struct, so it must not change.
// This is SEGGER_RTT_CB.
type rttControlBlock struct {
	id                [16]volatile.Register8
	maxNumUpBuffers   int32
	maxNumDownBuffers int32
	buffersUp         [rttMaxNumUpBuffers]rttBuffer
	buffersDown       [rttMaxNumDownBuffers]rttBuffer
}

// Up or down buffer.
// This is SEGGER_RTT_BUFFER_UP and SEGGER_RTT_BUFFER_DOWN.
type rttBuffer struct {
	name        *byte
	buffer      *volatile.Register8
	bufferSize  uint32
	writeOffset volatile.Register32
	readOffset  volatile.Register32
	flags       uint32
}

// Static buffers, for the default up and down buffer.
var (
	rttBufferUpData   [rttBufferSizeUp]volatile.Register8
	rttBufferDownData [rttBufferSizeDown]volatile.Register8
)

type rttSerial struct {
	rttControlBlock
}

func (s *rttSerial) Configure(config UARTConfig) error {
	s.maxNumUpBuffers = rttMaxNumUpBuffers
	s.maxNumDownBuffers = rttMaxNumDownBuffers

	s.buffersUp[0].name = &[]byte("Terminal\x00")[0]
	s.buffersUp[0].buffer = &rttBufferUpData[0]
	s.buffersUp[0].bufferSize = rttBufferSizeUp
	s.buffersUp[0].flags = rttModeNoBlockSkip

	s.buffersDown[0].name = &[]byte("Terminal\x00")[0]
	s.buffersDown[0].buffer = &rttBufferDownData[0]
	s.buffersDown[0].bufferSize = rttBufferSizeDown
	s.buffersDown[0].flags = rttModeNoBlockSkip

	id := "SEGGER RTT"
	for i := 0; i < len(id); i++ {
		s.id[i].Set(id[i])
	}

	return nil
}

func (b *rttBuffer) writeByte(c byte) {
	state := interrupt.Disable()
	readOffset := b.readOffset.Get()
	writeOffset := b.writeOffset.Get()
	newWriteOffset := writeOffset + 1
	if newWriteOffset == b.bufferSize {
		newWriteOffset = 0
	}
	if newWriteOffset != readOffset {
		unsafe.Slice(b.buffer, b.bufferSize)[writeOffset].Set(c)
		b.writeOffset.Set(newWriteOffset)
	}
	interrupt.Restore(state)
}

func (b *rttBuffer) readByte() byte {
	readOffset := b.readOffset.Get()
	writeOffset := b.writeOffset.Get()
	for readOffset == writeOffset {
		readOffset = b.readOffset.Get()
	}
	c := unsafe.Slice(b.buffer, b.bufferSize)[readOffset].Get()
	b.readOffset.Set(readOffset + 1)
	return c
}

func (b *rttBuffer) buffered() int {
	readOffset := b.readOffset.Get()
	writeOffset := b.writeOffset.Get()
	return int((writeOffset - readOffset) % rttBufferSizeDown)
}

func (s *rttSerial) WriteByte(b byte) error {
	s.buffersUp[0].writeByte(b)
	return nil
}

func (s *rttSerial) ReadByte() (byte, error) {
	return s.buffersDown[0].readByte(), errNoByte
}

func (s *rttSerial) Buffered() int {
	return s.buffersDown[0].buffered()
}

func (s *rttSerial) Write(data []byte) (n int, err error) {
	for _, v := range data {
		s.WriteByte(v)
	}
	return len(data), nil
}
