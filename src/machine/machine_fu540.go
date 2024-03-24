//go:build fu540

package machine

import (
	"errors"
	"unsafe"

	"device/riscv"
	"device/sifive"
	"runtime/interrupt"
)

const deviceName = sifive.Device

// non-SVD-stuff.
// On the FE310, these are in the SVD. But not on the other
// parts, which is ... strange.

const (
	MaxCore              = 5
	CLINTMSIP    uintptr = 0x2000000
	CLINTTimeCmp uintptr = 0x204000
	CLINTTime    uintptr = 0x20bff8
)

var (
// Coreplex Local Interrupts
)

// MSIPn returns the n'th MSIP.
// It is allowed to be called with a bad value, and will return
// 0, always.
func MSIPn(n uintptr) uint32 {
	if n > MaxCore {
		return 0
	}
	return *(*uint32)(unsafe.Pointer(CLINTMSIP + n*4))
}

// MSIP returns the MSIP for the current HART.
func MSIP() uint32 {
	return MSIPn(riscv.MHARTID.Get())
}

// MTimeCmpN returns the n'th MTimeCmp.
// It is allowed to be called with a bad value, and will return
// 0, always.
func MTimeCmpN(n uintptr) uint64 {
	if n > MaxCore {
		return 0
	}
	return *(*uint64)(unsafe.Pointer(CLINTTimeCmp + n*8))
}

// MTimeCmp returns the MTimeCmp for the current HART.
func MTimeCmp() uint64 {
	return MTimeCmpN(riscv.MHARTID.Get())
}

// MTimer returns the Mtime
func MTime() uint64 {
	return *(*uint64)(unsafe.Pointer(CLINTTime))
}

// End non-SVD-stuff.
func CPUFrequency() uint32 {
	return 390000000
}

var (
	errUnsupportedSPIController = errors.New("SPI controller not supported. Use SPI0 or SPI1.")
	errI2CTxAbort               = errors.New("I2C transmition has been aborted.")
)

type UART struct {
	Bus    *sifive.UART_Type
	Buffer *RingBuffer
}

var (
	DefaultUART = UART0
	UART0       = &_UART0
	_UART0      = UART{Bus: sifive.UART0, Buffer: NewRingBuffer()}
)

func (uart *UART) Configure(config UARTConfig) {
}

func (uart *UART) handleInterrupt(interrupt.Interrupt) {
}

func (uart *UART) writeByte(c byte) error {
	return nil
}

func (uart *UART) flush() {}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
	panic("Set")
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	panic("Get")
	return false
}

// TODO: SPI
