//go:build avr && atmega328p
// +build avr,atmega328p

package machine

import (
	"device/avr"
	"errors"
	"runtime/interrupt"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART_RX

// getPortMask returns the PORTx register and mask for the pin.
func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	switch {
	case p >= PB0 && p <= PB7: // port B
		return avr.PORTB, 1 << uint8(p-portB)
	case p >= PC0 && p <= PC7: // port C
		return avr.PORTC, 1 << uint8(p-portC)
	default: // port D
		return avr.PORTD, 1 << uint8(p-portD)
	}
}

// PWM is one PWM peripheral, which consists of a counter and two output
// channels (that can be connected to two fixed pins). You can set the frequency
// using SetPeriod, but only for all the channels in this PWM peripheral at
// once.
type PWM struct {
	num uint8
}

var (
	Timer0 = PWM{0} // 8 bit timer for PD5 and PD6
	Timer1 = PWM{1} // 16 bit timer for PB1 and PB2
	Timer2 = PWM{2} // 8 bit timer for PB3 and PD3
)

// Configure enables and configures this PWM.
//
// For the two 8 bit timers, there is only a limited number of periods
// available, namely the CPU frequency divided by 256 and again divided by 1, 8,
// 64, 256, or 1024. For a MCU running at 16MHz, this would be a period of 16µs,
// 128µs, 1024µs, 4096µs, or 16384µs.
func (pwm PWM) Configure(config PWMConfig) error {
	switch pwm.num {
	case 0, 2: // 8-bit timers (Timer/counter 0 and Timer/counter 2)
		// Calculate the timer prescaler.
		// While we could configure a flexible top, that would sacrifice one of
		// the PWM output compare registers and thus a PWM channel. I've chosen
		// to instead limit this timer to a fixed number of frequencies.
		var prescaler uint8
		switch config.Period {
		case 0, (uint64(1e9) * 256 * 1) / uint64(CPUFrequency()):
			prescaler = 1
		case (uint64(1e9) * 256 * 8) / uint64(CPUFrequency()):
			prescaler = 2
		case (uint64(1e9) * 256 * 64) / uint64(CPUFrequency()):
			prescaler = 3
		case (uint64(1e9) * 256 * 256) / uint64(CPUFrequency()):
			prescaler = 4
		case (uint64(1e9) * 256 * 1024) / uint64(CPUFrequency()):
			prescaler = 5
		default:
			return ErrPWMPeriodTooLong
		}

		if pwm.num == 0 {
			avr.TCCR0B.Set(prescaler)
			// Set the PWM mode to fast PWM (mode = 3).
			avr.TCCR0A.Set(avr.TCCR0A_WGM00 | avr.TCCR0A_WGM01)
			// monotonic timer is using the same time as PWM:0
			// we must adust internal settings of monotonic timer when PWM:0 settings changed
			adjustMonotonicTimer()
		} else {
			avr.TCCR2B.Set(prescaler)
			// Set the PWM mode to fast PWM (mode = 3).
			avr.TCCR2A.Set(avr.TCCR2A_WGM20 | avr.TCCR2A_WGM21)
		}
	case 1: // Timer/counter 1
		// The top value is the number of PWM ticks a PWM period takes. It is
		// initially picked assuming an unlimited counter top and no PWM
		// prescaler.
		var top uint64
		if config.Period == 0 {
			// Use a top appropriate for LEDs. Picking a relatively low period
			// here (0xff) for consistency with the other timers.
			top = 0xff
		} else {
			// The formula below calculates the following formula, optimized:
			//     top = period * (CPUFrequency() / 1e9)
			// By dividing the CPU frequency first (an operation that is easily
			// optimized away) the period has less chance of overflowing.
			top = config.Period * (uint64(CPUFrequency()) / 1000000) / 1000
		}

		avr.TCCR1A.Set(avr.TCCR1A_WGM11)

		// The ideal PWM period may be larger than would fit in the PWM counter,
		// which is 16 bits (see maxTop). Therefore, try to make the PWM clock
		// speed lower with a prescaler to make the top value fit the maximum
		// top value.
		const maxTop = 0x10000
		switch {
		case top <= maxTop:
			avr.TCCR1B.Set(3<<3 | 1) // no prescaling
		case top/8 <= maxTop:
			avr.TCCR1B.Set(3<<3 | 2) // divide by 8
			top /= 8
		case top/64 <= maxTop:
			avr.TCCR1B.Set(3<<3 | 3) // divide by 64
			top /= 64
		case top/256 <= maxTop:
			avr.TCCR1B.Set(3<<3 | 4) // divide by 256
			top /= 256
		case top/1024 <= maxTop:
			avr.TCCR1B.Set(3<<3 | 5) // divide by 1024
			top /= 1024
		default:
			return ErrPWMPeriodTooLong
		}

		// A top of 0x10000 is at 100% duty cycle. Subtract one because the
		// counter counts from 0, not 1 (avoiding an off-by-one).
		top -= 1

		avr.ICR1H.Set(uint8(top >> 8))
		avr.ICR1L.Set(uint8(top))
	}
	return nil
}

// SetPeriod updates the period of this PWM peripheral.
// To set a particular frequency, use the following formula:
//
//	period = 1e9 / frequency
//
// If you use a period of 0, a period that works well for LEDs will be picked.
//
// SetPeriod will not change the prescaler, but also won't change the current
// value in any of the channels. This means that you may need to update the
// value for the particular channel.
//
// Note that you cannot pick any arbitrary period after the PWM peripheral has
// been configured. If you want to switch between frequencies, pick the lowest
// frequency (longest period) once when calling Configure and adjust the
// frequency here as needed.
func (pwm PWM) SetPeriod(period uint64) error {
	if pwm.num != 1 {
		return ErrPWMPeriodTooLong // TODO better error message
	}

	// The top value is the number of PWM ticks a PWM period takes. It is
	// initially picked assuming an unlimited counter top and no PWM
	// prescaler.
	var top uint64
	if period == 0 {
		// Use a top appropriate for LEDs. Picking a relatively low period
		// here (0xff) for consistency with the other timers.
		top = 0xff
	} else {
		// The formula below calculates the following formula, optimized:
		//     top = period * (CPUFrequency() / 1e9)
		// By dividing the CPU frequency first (an operation that is easily
		// optimized away) the period has less chance of overflowing.
		top = period * (uint64(CPUFrequency()) / 1000000) / 1000
	}

	prescaler := avr.TCCR1B.Get() & 0x7
	switch prescaler {
	case 1:
		top /= 1
	case 2:
		top /= 8
	case 3:
		top /= 64
	case 4:
		top /= 256
	case 5:
		top /= 1024
	}

	// A top of 0x10000 is at 100% duty cycle. Subtract one because the counter
	// counts from 0, not 1 (avoiding an off-by-one).
	top -= 1

	if top > 0xffff {
		return ErrPWMPeriodTooLong
	}

	// Warning: this change is not atomic!
	avr.ICR1H.Set(uint8(top >> 8))
	avr.ICR1L.Set(uint8(top))

	// ... and because of that, set the counter back to zero to avoid most of
	// the effects of this non-atomicity.
	avr.TCNT1H.Set(0)
	avr.TCNT1L.Set(0)

	return nil
}

// Top returns the current counter top, for use in duty cycle calculation. It
// will only change with a call to Configure or SetPeriod, otherwise it is
// constant.
//
// The value returned here is hardware dependent. In general, it's best to treat
// it as an opaque value that can be divided by some number and passed to Set
// (see Set documentation for more information).
func (pwm PWM) Top() uint32 {
	if pwm.num == 1 {
		// Timer 1 has a configurable top value.
		low := avr.ICR1L.Get()
		high := avr.ICR1H.Get()
		return uint32(high)<<8 | uint32(low) + 1
	}
	// Other timers go from 0 to 0xff (0x100 or 256 in total).
	return 256
}

// Counter returns the current counter value of the timer in this PWM
// peripheral. It may be useful for debugging.
func (pwm PWM) Counter() uint32 {
	switch pwm.num {
	case 0:
		return uint32(avr.TCNT0.Get())
	case 1:
		mask := interrupt.Disable()
		low := avr.TCNT1L.Get()
		high := avr.TCNT1H.Get()
		interrupt.Restore(mask)
		return uint32(high)<<8 | uint32(low)
	case 2:
		return uint32(avr.TCNT2.Get())
	}
	// Unknown PWM.
	return 0
}

// Period returns the used PWM period in nanoseconds. It might deviate slightly
// from the configured period due to rounding.
func (pwm PWM) Period() uint64 {
	var prescaler uint8
	switch pwm.num {
	case 0:
		prescaler = avr.TCCR0B.Get() & 0x7
	case 1:
		prescaler = avr.TCCR1B.Get() & 0x7
	case 2:
		prescaler = avr.TCCR2B.Get() & 0x7
	}
	top := uint64(pwm.Top())
	switch prescaler {
	case 1: // prescaler 1
		return 1 * top * 1000 / uint64(CPUFrequency()/1e6)
	case 2: // prescaler 8
		return 8 * top * 1000 / uint64(CPUFrequency()/1e6)
	case 3: // prescaler 64
		return 64 * top * 1000 / uint64(CPUFrequency()/1e6)
	case 4: // prescaler 256
		return 256 * top * 1000 / uint64(CPUFrequency()/1e6)
	case 5: // prescaler 1024
		return 1024 * top * 1000 / uint64(CPUFrequency()/1e6)
	default: // unknown clock source
		return 0
	}
}

// Channel returns a PWM channel for the given pin.
func (pwm PWM) Channel(pin Pin) (uint8, error) {
	pin.Configure(PinConfig{Mode: PinOutput})
	pin.Low()
	switch pwm.num {
	case 0:
		switch pin {
		case PD6: // channel A
			avr.TCCR0A.SetBits(avr.TCCR0A_COM0A1)
			return 0, nil
		case PD5: // channel B
			avr.TCCR0A.SetBits(avr.TCCR0A_COM0B1)
			return 1, nil
		}
	case 1:
		switch pin {
		case PB1: // channel A
			avr.TCCR1A.SetBits(avr.TCCR1A_COM1A1)
			return 0, nil
		case PB2: // channel B
			avr.TCCR1A.SetBits(avr.TCCR1A_COM1B1)
			return 1, nil
		}
	case 2:
		switch pin {
		case PB3: // channel A
			avr.TCCR2A.SetBits(avr.TCCR2A_COM2A1)
			return 0, nil
		case PD3: // channel B
			avr.TCCR2A.SetBits(avr.TCCR2A_COM2B1)
			return 1, nil
		}
	}
	return 0, ErrInvalidOutputPin
}

// SetInverting sets whether to invert the output of this channel.
// Without inverting, a 25% duty cycle would mean the output is high for 25% of
// the time and low for the rest. Inverting flips the output as if a NOT gate
// was placed at the output, meaning that the output would be 25% low and 75%
// high with a duty cycle of 25%.
//
// Note: the invert state may not be applied on the AVR until the next call to
// ch.Set().
func (pwm PWM) SetInverting(channel uint8, inverting bool) {
	switch pwm.num {
	case 0:
		switch channel {
		case 0: // channel A
			if inverting {
				avr.PORTB.SetBits(1 << 6) // PB6 high
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0A0)
			} else {
				avr.PORTB.ClearBits(1 << 6) // PB6 low
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0A0)
			}
		case 1: // channel B
			if inverting {
				avr.PORTB.SetBits(1 << 5) // PB5 high
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0B0)
			} else {
				avr.PORTB.ClearBits(1 << 5) // PB5 low
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0B0)
			}
		}
	case 1:
		// Note: the COM1A0/COM1B0 bit is not set with the configuration below.
		// It will be set the following call to Set(), however.
		switch channel {
		case 0: // channel A, PB1
			if inverting {
				avr.PORTB.SetBits(1 << 1) // PB1 high
			} else {
				avr.PORTB.ClearBits(1 << 1) // PB1 low
			}
		case 1: // channel B, PB2
			if inverting {
				avr.PORTB.SetBits(1 << 2) // PB2 high
			} else {
				avr.PORTB.ClearBits(1 << 2) // PB2 low
			}
		}
	case 2:
		switch channel {
		case 0: // channel A
			if inverting {
				avr.PORTB.SetBits(1 << 3) // PB3 high
				avr.TCCR2A.SetBits(avr.TCCR2A_COM2A0)
			} else {
				avr.PORTB.ClearBits(1 << 3) // PB3 low
				avr.TCCR2A.ClearBits(avr.TCCR2A_COM2A0)
			}
		case 1: // channel B
			if inverting {
				avr.PORTD.SetBits(1 << 3) // PD3 high
				avr.TCCR2A.SetBits(avr.TCCR2A_COM2B0)
			} else {
				avr.PORTD.ClearBits(1 << 3) // PD3 low
				avr.TCCR2A.ClearBits(avr.TCCR2A_COM2B0)
			}
		}
	}
}

// Set updates the channel value. This is used to control the channel duty
// cycle, in other words the fraction of time the channel output is high (or low
// when inverted). For example, to set it to a 25% duty cycle, use:
//
//	pwm.Set(channel, pwm.Top() / 4)
//
// pwm.Set(channel, 0) will set the output to low and pwm.Set(channel,
// pwm.Top()) will set the output to high, assuming the output isn't inverted.
func (pwm PWM) Set(channel uint8, value uint32) {
	switch pwm.num {
	case 0:
		value := uint16(value)
		switch channel {
		case 0: // channel A
			if value == 0 {
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0A1)
			} else {
				avr.OCR0A.Set(uint8(value - 1))
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0A1)
			}
		case 1: // channel B
			if value == 0 {
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0B1)
			} else {
				avr.OCR0B.Set(uint8(value) - 1)
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0B1)
			}
		}
		// monotonic timer is using the same time as PWM:0
		// we must adust internal settings of monotonic timer when PWM:0 settings changed
		adjustMonotonicTimer()
	case 1:
		mask := interrupt.Disable()
		switch channel {
		case 0: // channel A, PB1
			if value == 0 {
				avr.TCCR1A.ClearBits(avr.TCCR1A_COM1A1 | avr.TCCR1A_COM1A0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR1AH.Set(uint8(value >> 8))
				avr.OCR1AL.Set(uint8(value))
				if avr.PORTB.HasBits(1 << 1) { // is PB1 high?
					// Yes, set the inverting bit.
					avr.TCCR1A.SetBits(avr.TCCR1A_COM1A1 | avr.TCCR1A_COM1A0)
				} else {
					// No, output is non-inverting.
					avr.TCCR1A.SetBits(avr.TCCR1A_COM1A1)
				}
			}
		case 1: // channel B, PB2
			if value == 0 {
				avr.TCCR1A.ClearBits(avr.TCCR1A_COM1B1 | avr.TCCR1A_COM1B0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR1BH.Set(uint8(value >> 8))
				avr.OCR1BL.Set(uint8(value))
				if avr.PORTB.HasBits(1 << 2) { // is PB2 high?
					// Yes, set the inverting bit.
					avr.TCCR1A.SetBits(avr.TCCR1A_COM1B1 | avr.TCCR1A_COM1B0)
				} else {
					// No, output is non-inverting.
					avr.TCCR1A.SetBits(avr.TCCR1A_COM1B1)
				}
			}
		}
		interrupt.Restore(mask)
	case 2:
		value := uint16(value)
		switch channel {
		case 0: // channel A
			if value == 0 {
				avr.TCCR2A.ClearBits(avr.TCCR2A_COM2A1)
			} else {
				avr.OCR2A.Set(uint8(value - 1))
				avr.TCCR2A.SetBits(avr.TCCR2A_COM2A1)
			}
		case 1: // channel B
			if value == 0 {
				avr.TCCR2A.ClearBits(avr.TCCR2A_COM2B1)
			} else {
				avr.OCR2B.Set(uint8(value - 1))
				avr.TCCR2A.SetBits(avr.TCCR2A_COM2B1)
			}
		}
	}
}

// SPI configuration
var SPI0 = SPI{
	spcr: avr.SPCR,
	spdr: avr.SPDR,
	spsr: avr.SPSR,
	sck:  PB5,
	sdo:  PB3,
	sdi:  PB4,
	cs:   PB2}

// Pin Change Interrupts
var (
	pinCallbacks [3][8]func(Pin)
	interrupts   [3]*interrupt.Interrupt
	previous     [3]uint8
)

func handlePCINTInterrupts(intr uint8, port *volatile.Register8) {
	current := port.Get()
	for i := uint8(0); i < 8; i++ {
		if (current>>i)&0x01 == 0x01 && pinCallbacks[intr][i] != nil {
			pinCallbacks[intr][i](Pin(i * (intr + 1)))
		}
	}
	previous[intr] = current
}

func handlePCINT0Interrupts(intr interrupt.Interrupt) {
	handlePCINTInterrupts(avr.IRQ_PCINT0-avr.IRQ_PCINT0, avr.PINB)
}

func handlePCINT1Interrupts(intr interrupt.Interrupt) {
	handlePCINTInterrupts(avr.IRQ_PCINT1-avr.IRQ_PCINT0, avr.PINC)
}

func handlePCINT2Interrupts(intr interrupt.Interrupt) {
	handlePCINTInterrupts(avr.IRQ_PCINT2-avr.IRQ_PCINT0, avr.PIND)
}

func SetInterrupt(pin Pin, callback func(Pin)) (err error) {
	if callback == nil {
		// Disable this pin interrupt
		return errors.New("Not implemented")
	} else {
		var PCMSK_REG *volatile.Register8
		var PCIE uint8
		var PCINT uint8
		var IRQ uint8
		var PIN0 Pin
		//var PIN *volatile.Register8
		switch {
		case pin >= PB0 && pin <= PB7:
			// PCMSK0 - PCINT0-7
			PCMSK_REG = avr.PCMSK0
			PCIE = avr.PCICR_PCIE0
			PIN0 = PD0
			PCINT = 1 << (pin - PIN0)
			IRQ = avr.IRQ_PCINT0
			if interrupts[IRQ-avr.IRQ_PCINT0] == nil {
				intr := interrupt.New(avr.IRQ_PCINT0, handlePCINT0Interrupts)
				interrupts[IRQ-avr.IRQ_PCINT0] = &intr
			} else {
			}
		case pin >= PC0 && pin <= PC7:
			// PCMSK1 - PCINT8-14
			PCMSK_REG = avr.PCMSK1
			PCIE = avr.PCICR_PCIE1
			PIN0 = PC0
			PCINT = 1 << (pin - PIN0)
			IRQ = avr.IRQ_PCINT1
			if interrupts[IRQ-avr.IRQ_PCINT0] == nil {
				intr := interrupt.New(avr.IRQ_PCINT1, handlePCINT1Interrupts)
				interrupts[IRQ-avr.IRQ_PCINT0] = &intr
			} else {
			}
		case pin >= PD0 && pin <= PD7:
			// PCMSK2 - PCINT16-23
			PCMSK_REG = avr.PCMSK2
			PCIE = avr.PCICR_PCIE2
			PIN0 = PD0
			PCINT = 1 << (pin - PIN0)
			IRQ = avr.IRQ_PCINT2
			if interrupts[IRQ-avr.IRQ_PCINT0] == nil {
				intr := interrupt.New(avr.IRQ_PCINT2, handlePCINT2Interrupts)
				interrupts[IRQ-avr.IRQ_PCINT0] = &intr
			}
		default:
			return errors.New("Unsupported pin")
		}

		if pinCallbacks[IRQ-avr.IRQ_PCINT0][pin-PIN0] == nil {
			pinCallbacks[IRQ-avr.IRQ_PCINT0][pin-PIN0] = callback
		} else {
			return errors.New("callback already defined")
		}

		PCMSK_REG.SetBits(PCINT)
		avr.PCICR.SetBits(PCIE)

		return nil
	}
}
