//go:build stm32l4 || stm32l5

package machine

import (
	"device/stm32"
)

// This variant of the GPIO input interrupt logic is for
// chips with a larger number of interrupt channels (more
// than fits in a single register).

//
// STM32 allows one interrupt source per pin number, with
// the same pin number in different ports sharing a single
// interrupt source (so PA0, PB0, PC0 all share).  Only a
// single physical pin can be connected to each interrupt
// line.
//
// To call interrupt callbacks, we record here for each
// pin number the callback and the actual associated pin.
//

// Callbacks for pin interrupt events
var pinCallbacks [16]func(Pin)

// The pin currently associated with interrupt callback
// for a given slot.
var interruptPins [16]Pin

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
//
// This call will replace a previously set callback on this pin. You can pass a
// nil func to unset the pin change interrupt. If you do so, the change
// parameter is ignored and can be set to any value (such as 0).
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	port := uint32(uint8(p) / 16)
	pin := uint8(p) % 16

	enableEXTIConfigRegisters()

	if callback == nil {
		stm32.EXTI.IMR1.ClearBits(1 << pin)
		pinCallbacks[pin] = nil
		return nil
	}

	if pinCallbacks[pin] != nil {
		// The pin was already configured.
		// To properly re-configure a pin, unset it first and set a new
		// configuration.
		return ErrNoPinChangeChannel
	}

	// Set the callback now (before the interrupt is enabled) to avoid
	// possible race condition
	pinCallbacks[pin] = callback
	interruptPins[pin] = p

	crReg := getEXTIConfigRegister(pin)
	shift := (pin & 0x3) * 4
	crReg.ReplaceBits(port, 0xf, shift)

	if (change & PinRising) != 0 {
		stm32.EXTI.RTSR1.SetBits(1 << pin)
	}
	if (change & PinFalling) != 0 {
		stm32.EXTI.FTSR1.SetBits(1 << pin)
	}
	stm32.EXTI.IMR1.SetBits(1 << pin)

	intr := p.registerInterrupt()
	intr.SetPriority(0)
	intr.Enable()

	return nil
}
