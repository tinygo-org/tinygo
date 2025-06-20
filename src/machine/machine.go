package machine

import (
	"errors"
	"unsafe"
)

var (
	ErrTimeoutRNG         = errors.New("machine: RNG Timeout")
	ErrClockRNG           = errors.New("machine: RNG Clock Error")
	ErrSeedRNG            = errors.New("machine: RNG Seed Error")
	ErrInvalidInputPin    = errors.New("machine: invalid input pin")
	ErrInvalidOutputPin   = errors.New("machine: invalid output pin")
	ErrInvalidClockPin    = errors.New("machine: invalid clock pin")
	ErrInvalidDataPin     = errors.New("machine: invalid data pin")
	ErrNoPinChangeChannel = errors.New("machine: no channel available for pin interrupt")
)

// Device is the running program's chip name, such as "ATSAMD51J19A" or
// "nrf52840". It is not the same as the CPU name.
//
// The constant is some hardcoded default value if the program does not target a
// particular chip but instead runs in WebAssembly for example.
const Device = deviceName

// Generic constants.
const (
	KHz = 1000
	MHz = 1000_000
	GHz = 1000_000_000
)

// PinMode sets the direction and pull mode of the pin. For example, PinOutput
// sets the pin as an output and PinInputPullup sets the pin as an input with a
// pull-up.
type PinMode uint8

type PinConfig struct {
	Mode PinMode
}

// Pin is a single pin on a chip, which may be connected to other hardware
// devices. It can either be used directly as GPIO pin or it can be used in
// other peripherals like ADC, I2C, etc.
type Pin uint8

// NoPin explicitly indicates "not a pin". Use this pin if you want to leave one
// of the pins in a peripheral unconfigured (if supported by the hardware).
const NoPin = Pin(0xff)

// High sets this GPIO pin to high, assuming it has been configured as an output
// pin. It is hardware dependent (and often undefined) what happens if you set a
// pin to high that is not configured as an output pin.
func (p Pin) High() {
	p.Set(true)
}

// Low sets this GPIO pin to low, assuming it has been configured as an output
// pin. It is hardware dependent (and often undefined) what happens if you set a
// pin to low that is not configured as an output pin.
func (p Pin) Low() {
	p.Set(false)
}

type ADC struct {
	Pin Pin
}

// Convert the pointer to a uintptr, to be used for memory I/O (DMA for
// example). It also means the pointer is "gone" as far as the compiler is
// concerned, and a GC cycle might deallocate the object. To prevent this from
// happening, also call keepAliveNoEscape at a point after the address isn't
// accessed anymore by the hardware.
// The only exception is if the pointer is accessed later in a volatile way
// (volatile read/write), which also forces the value to stay alive until that
// point.
//
// This function is treated specially by the compiler to mark the 'ptr'
// parameter as not escaping.
//
// TODO: this function should eventually be replaced with the proposed ptrtoaddr
// instruction in LLVM. See:
// https://discourse.llvm.org/t/clarifiying-the-semantics-of-ptrtoint/83987/10
// https://github.com/llvm/llvm-project/pull/139357
func unsafeNoEscape(ptr unsafe.Pointer) uintptr {
	return uintptr(ptr)
}

// Make sure the given pointer stays alive until this point. This is similar to
// runtime.KeepAlive, with the difference that it won't let the pointer escape.
// This is typically used together with unsafeNoEscape.
//
// This is a compiler intrinsic.
func keepAliveNoEscape(ptr unsafe.Pointer)
