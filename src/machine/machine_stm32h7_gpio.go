//go:build stm32 && stm32h7

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

// Callbacks for pin interrupt events
var pinCallbacks [16]func(Pin)

// The pin currently associated with interrupt callback
// for a given slot.
var interruptPins [16]Pin

// SetInterrupt sets an interrupt to be executed when a particular pin changes
// state. The pin should already be configured as an input, including a pull up
// or down if no external pull is provided.
func (p Pin) SetInterrupt(change PinChange, callback func(Pin)) error {
	port := uint32(uint8(p) / 16)
	pin := uint8(p) % 16

	enableEXTIConfigRegisters()

	if callback == nil {
		stm32.EXTI.CPUIMR1.ClearBits(1 << pin)
		pinCallbacks[pin] = nil
		return nil
	}

	if pinCallbacks[pin] != nil {
		return ErrNoPinChangeChannel
	}

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
	stm32.EXTI.CPUIMR1.SetBits(1 << pin)

	intr := p.registerInterrupt()
	intr.SetPriority(0)
	intr.Enable()

	return nil
}

func (p Pin) registerInterrupt() interrupt.Interrupt {
	pin := uint8(p) % 16
	switch pin {
	case 0:
		return interrupt.New(stm32.IRQ_EXTI0, handlePinInterrupt0)
	case 1:
		return interrupt.New(stm32.IRQ_EXTI1, handlePinInterrupt1)
	case 2:
		return interrupt.New(stm32.IRQ_EXTI2, handlePinInterrupt2)
	case 3:
		return interrupt.New(stm32.IRQ_EXTI3, handlePinInterrupt3)
	case 4:
		return interrupt.New(stm32.IRQ_EXTI4, handlePinInterrupt4)
	case 5, 6, 7, 8, 9:
		return interrupt.New(stm32.IRQ_EXTI9_5, handlePinInterrupt9_5)
	case 10, 11, 12, 13, 14, 15:
		return interrupt.New(stm32.IRQ_EXTI15_10, handlePinInterrupt15_10)
	}
	return interrupt.Interrupt{}
}

func handlePinInterrupt0(interrupt.Interrupt) { handlePinInterrupt(0) }
func handlePinInterrupt1(interrupt.Interrupt) { handlePinInterrupt(1) }
func handlePinInterrupt2(interrupt.Interrupt) { handlePinInterrupt(2) }
func handlePinInterrupt3(interrupt.Interrupt) { handlePinInterrupt(3) }
func handlePinInterrupt4(interrupt.Interrupt) { handlePinInterrupt(4) }
func handlePinInterrupt9_5(interrupt.Interrupt) {
	handlePinInterrupt(5)
	handlePinInterrupt(6)
	handlePinInterrupt(7)
	handlePinInterrupt(8)
	handlePinInterrupt(9)
}
func handlePinInterrupt15_10(interrupt.Interrupt) {
	handlePinInterrupt(10)
	handlePinInterrupt(11)
	handlePinInterrupt(12)
	handlePinInterrupt(13)
	handlePinInterrupt(14)
	handlePinInterrupt(15)
}

func handlePinInterrupt(pin uint8) {
	if stm32.EXTI.CPUPR1.HasBits(1 << pin) {
		// Writing 1 to the pending register clears the
		// pending flag for that bit
		stm32.EXTI.CPUPR1.Set(1 << pin)

		callback := pinCallbacks[pin]
		if callback != nil {
			callback(interruptPins[pin])
		}
	}
}
