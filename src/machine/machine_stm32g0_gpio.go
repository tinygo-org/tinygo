//go:build stm32g0

package machine

import (
	"device/stm32"
)

// This variant of the GPIO input interrupt logic is for
// STM32G0 chips which use a different EXTI register structure
// with IMR1, RTSR1, FTSR1, and separate RPR1/FPR1 pending registers.

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

func handlePinInterrupt(pin uint8) {
	// STM32G0 has separate rising and falling pending registers
	// Check both and clear the appropriate one
	mask := uint32(1 << pin)
	if stm32.EXTI.RPR1.HasBits(mask) {
		// Writing 1 to the pending register clears the pending flag
		stm32.EXTI.RPR1.Set(mask)

		callback := pinCallbacks[pin]
		if callback != nil {
			callback(interruptPins[pin])
		}
	}
	if stm32.EXTI.FPR1.HasBits(mask) {
		// Writing 1 to the pending register clears the pending flag
		stm32.EXTI.FPR1.Set(mask)

		callback := pinCallbacks[pin]
		if callback != nil {
			callback(interruptPins[pin])
		}
	}
}
