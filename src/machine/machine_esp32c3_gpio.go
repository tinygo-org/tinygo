// +build esp32c3

package machine

import (
	"device/esp"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const (
	PinOutput PinMode = iota
	PinInput
	PinInputPullup
	PinInputPulldown
)

type PinChange uint8

const (
	PinInterruptDisabled PinChange = 0
	PinInterruptRising   PinChange = 1
	PinInterruptFalling  PinChange = 2
	PinInterruptToggle   PinChange = 3
)

var ErrInvalidPinNumber = errors.New("invlaid pin number")

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	if p == NoPin {
		// This simplifies pin configuration in peripherals such as SPI.
		return
	}

	var muxConfig uint32

	// Configure this pin as a GPIO pin.
	const function = 1 // function 1 is GPIO for every pin
	muxConfig |= function << esp.IO_MUX_GPIO_MCU_SEL_Pos

	// Make this pin an input pin (always).
	muxConfig |= esp.IO_MUX_GPIO_FUN_IE

	// Set drive strength: 0 is lowest, 3 is highest.
	muxConfig |= 2 << esp.IO_MUX_GPIO_FUN_DRV_Pos

	// Select pull mode.
	if config.Mode == PinInputPullup {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPU
	} else if config.Mode == PinInputPulldown {
		muxConfig |= esp.IO_MUX_GPIO_FUN_WPD
	}

	// Configure the pad with the given IO mux configuration.
	p.mux().Set(muxConfig)

	// Set the output signal to the simple GPIO output.
	p.outFunc().Set(0x80)

	switch config.Mode {
	case PinOutput:
		// Set the 'output enable' bit.
		esp.GPIO.ENABLE_W1TS.Set(1 << p)
	case PinInput, PinInputPullup, PinInputPulldown:
		// Clear the 'output enable' bit.
		esp.GPIO.ENABLE_W1TC.Set(1 << p)
	}
}

// outFunc returns the FUNCx_OUT_SEL_CFG register used for configuring the
// output function selection.
func (p Pin) outFunc() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_OUT_SEL_CFG)) + uintptr(p)*4)))
}

// inFunc returns the FUNCy_IN_SEL_CFG register used for configuring the input
// function selection.
func inFunc(signal uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.FUNC0_IN_SEL_CFG)) + uintptr(signal)*4)))
}

// mux returns the I/O mux configuration register corresponding to the given
// GPIO pin.
func (p Pin) mux() *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.IO_MUX.GPIO0)) + uintptr(p)*4)))
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	return (esp.GPIO.IN.Get() & (1 << p)) == (1 << p)
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(value bool) {
	if value {
		reg, mask := p.portMaskSet()
		reg.Set(mask)
	} else {
		reg, mask := p.portMaskClear()
		reg.Set(mask)
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskSet() (*uint32, uint32) {
	reg, mask := p.portMaskSet()
	return &reg.Reg, mask
}

// Return the register and mask to disable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: only use this on an output pin!
func (p Pin) PortMaskClear() (*uint32, uint32) {
	reg, mask := p.portMaskClear()
	return &reg.Reg, mask
}

func (p Pin) portMaskSet() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TS, 1 << p
}

func (p Pin) portMaskClear() (*volatile.Register32, uint32) {
	return &esp.GPIO.OUT_W1TC, 1 << p
}

//go:extern void gpio_matrix_in(uint32_t gpio, uint32_t signal_idx, bool inv);
func gpio_matrix_in(gpio, signla uint32, inv byte)

var (
	pinCallbacks [26]func(Pin)
)

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// You can pass a nil func to unset the pin change interrupt. If you do so,
// the change parameter is ignored and can be set to any value (such as 0).
// If the pin is already configured with a callback, you must first unset
// this pins interrupt before you can set a new callback.
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	if p >= 26 {
		return ErrInvalidPinNumber
	}

	if callback == nil || change == PinInterruptDisabled {
		// Disable this pin interrupt (if it was enabled).
		p.disablePinInterrupt()
		if pinCallbacks[p] != nil {
			pinCallbacks[p] = nil
		}
		return nil
	}

	if pinCallbacks[p] != nil {
		// The pin was already configured.
		// To properly re-configure a pin, unset it first and set a new
		// configuration.
		return ErrNoPinChangeChannel
	}
	pinCallbacks[p] = callback

	intr := interrupt.New(MODULE_GPIO, func(i interrupt.Interrupt) {
		status := esp.GPIO.STATUS.Get()
		for i := 0; i < 26; i++ {
			if (status & (1 << i)) != 0 {
				pinCallbacks[i](Pin(i))
			}
		}

		// clear interrupt bit
		esp.GPIO.STATUS_W1TC.SetBits(status)
	})
	intr.Enable(InterruptForModule(MODULE_GPIO), MapRegisterForModule(MODULE_GPIO))

	p.enablePinInterrupt(change)

	return nil
}

func (p Pin) enablePinInterrupt(change PinChange) {
	reg := (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.PIN0)) + uintptr(p)*4)))
	reg.SetBits((uint32(change) << esp.GPIO_PIN_PIN_INT_TYPE_Pos) | (uint32(1) << esp.GPIO_PIN_PIN_INT_ENA_Pos))
}

func (p Pin) disablePinInterrupt() {
	reg := (*volatile.Register32)(unsafe.Pointer((uintptr(unsafe.Pointer(&esp.GPIO.PIN0)) + uintptr(p)*4)))
	reg.ClearBits(esp.GPIO_PIN_PIN_INT_TYPE_Msk | esp.GPIO_PIN_PIN_INT_ENA_Msk)
}
