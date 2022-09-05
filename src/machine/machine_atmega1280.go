//go:build avr && atmega1280
// +build avr,atmega1280

package machine

import (
	"device/avr"
	"runtime/interrupt"
	"runtime/volatile"
)

const irq_USART0_RX = avr.IRQ_USART0_RX

const (
	portA Pin = iota * 8
	portB
	portC
	portD
	portE
	portF
	portG
	portH
	portJ
	portK
	portL
)

const (
	PA0 = portA + 0
	PA1 = portA + 1
	PA2 = portA + 2
	PA3 = portA + 3
	PA4 = portA + 4
	PA5 = portA + 5
	PA6 = portA + 6
	PA7 = portA + 7
	PB0 = portB + 0
	PB1 = portB + 1
	PB2 = portB + 2
	PB3 = portB + 3
	PB4 = portB + 4 // peripherals: Timer2 channel A
	PB5 = portB + 5 // peripherals: Timer1 channel A
	PB6 = portB + 6 // peripherals: Timer1 channel B
	PB7 = portB + 7 // peripherals: Timer0 channel A
	PC0 = portC + 0
	PC1 = portC + 1
	PC2 = portC + 2
	PC3 = portC + 3
	PC4 = portC + 4
	PC5 = portC + 5
	PC6 = portC + 6
	PC7 = portC + 7
	PD0 = portD + 0
	PD1 = portD + 1
	PD2 = portD + 2
	PD3 = portD + 3
	PD7 = portD + 7
	PE0 = portE + 0
	PE1 = portE + 1
	PE3 = portE + 3 // peripherals: Timer3 channel A
	PE4 = portE + 4 // peripherals: Timer3 channel B
	PE5 = portE + 5 // peripherals: Timer3 channel C
	PE6 = portE + 6
	PF0 = portF + 0
	PF1 = portF + 1
	PF2 = portF + 2
	PF3 = portF + 3
	PF4 = portF + 4
	PF5 = portF + 5
	PF6 = portF + 6
	PF7 = portF + 7
	PG0 = portG + 0
	PG1 = portG + 1
	PG2 = portG + 2
	PG5 = portG + 5 // peripherals: Timer0 channel B
	PH0 = portH + 0
	PH1 = portH + 1
	PH3 = portH + 3 // peripherals: Timer4 channel A
	PH4 = portH + 4 // peripherals: Timer4 channel B
	PH5 = portH + 5 // peripherals: Timer4 channel C
	PH6 = portH + 6 // peripherals: Timer0 channel B
	PJ0 = portJ + 0
	PJ1 = portJ + 1
	PK0 = portK + 0
	PK1 = portK + 1
	PK2 = portK + 2
	PK3 = portK + 3
	PK4 = portK + 4
	PK5 = portK + 5
	PK6 = portK + 6
	PK7 = portK + 7
	PL0 = portL + 0
	PL1 = portL + 1
	PL2 = portL + 2
	PL3 = portL + 3 // peripherals: Timer5 channel A
	PL4 = portL + 4 // peripherals: Timer5 channel B
	PL5 = portL + 5 // peripherals: Timer5 channel C
	PL6 = portL + 6
	PL7 = portL + 7
)

// getPortMask returns the PORTx register and mask for the pin.
func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	switch {
	case p >= PA0 && p <= PA7:
		return avr.PORTA, 1 << uint8(p-portA)
	case p >= PB0 && p <= PB7:
		return avr.PORTB, 1 << uint8(p-portB)
	case p >= PC0 && p <= PC7:
		return avr.PORTC, 1 << uint8(p-portC)
	case p >= PD0 && p <= PD7:
		return avr.PORTD, 1 << uint8(p-portD)
	case p >= PE0 && p <= PE6:
		return avr.PORTE, 1 << uint8(p-portE)
	case p >= PF0 && p <= PF7:
		return avr.PORTF, 1 << uint8(p-portF)
	case p >= PG0 && p <= PG5:
		return avr.PORTG, 1 << uint8(p-portG)
	case p >= PH0 && p <= PH6:
		return avr.PORTH, 1 << uint8(p-portH)
	case p >= PJ0 && p <= PJ1:
		return avr.PORTJ, 1 << uint8(p-portJ)
	case p >= PK0 && p <= PK7:
		return avr.PORTK, 1 << uint8(p-portK)
	case p >= PL0 && p <= PL7:
		return avr.PORTL, 1 << uint8(p-portL)
	default:
		return avr.PORTA, 255
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
	Timer0 = PWM{0} // 8 bit timer for PB7 and PG5
	Timer1 = PWM{1} // 16 bit timer for PB5 and PB6
	Timer2 = PWM{2} // 8 bit timer for PB4 and PH6
	Timer3 = PWM{3} // 16 bit timer for PE3, PE4 and PE5
	Timer4 = PWM{4} // 16 bit timer for PH3, PH4 and PH5
	Timer5 = PWM{5} // 16 bit timer for PL3, PL4 and PL5
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
	case 1, 3, 4, 5:
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

		// The ideal PWM period may be larger than would fit in the PWM counter,
		// which is 16 bits (see maxTop). Therefore, try to make the PWM clock
		// speed lower with a prescaler to make the top value fit the maximum
		// top value.

		const maxTop = 0x10000
		var prescalingTop uint8
		switch {
		case top <= maxTop:
			prescalingTop = 3<<3 | 1 // no prescaling
		case top/8 <= maxTop:
			prescalingTop = 3<<3 | 2 // divide by 8
			top /= 8
		case top/64 <= maxTop:
			prescalingTop = 3<<3 | 3 // divide by 64
			top /= 64
		case top/256 <= maxTop:
			prescalingTop = 3<<3 | 4 // divide by 256
			top /= 256
		case top/1024 <= maxTop:
			prescalingTop = 3<<3 | 5 // divide by 1024
			top /= 1024
		default:
			return ErrPWMPeriodTooLong
		}

		// A top of 0x10000 is at 100% duty cycle. Subtract one because the
		// counter counts from 0, not 1 (avoiding an off-by-one).
		top -= 1

		switch pwm.num {
		case 1:
			avr.TCCR1A.Set(avr.TCCR1A_WGM11)
			avr.TCCR1B.Set(prescalingTop)
			avr.ICR1H.Set(uint8(top >> 8))
			avr.ICR1L.Set(uint8(top))
		case 3:
			avr.TCCR3A.Set(avr.TCCR3A_WGM31)
			avr.TCCR3B.Set(prescalingTop)
			avr.ICR3H.Set(uint8(top >> 8))
			avr.ICR3L.Set(uint8(top))
		case 4:
			avr.TCCR4A.Set(avr.TCCR4A_WGM41)
			avr.TCCR4B.Set(prescalingTop)
			avr.ICR4H.Set(uint8(top >> 8))
			avr.ICR4L.Set(uint8(top))
		case 5:
			avr.TCCR5A.Set(avr.TCCR5A_WGM51)
			avr.TCCR5B.Set(prescalingTop)
			avr.ICR5H.Set(uint8(top >> 8))
			avr.ICR5L.Set(uint8(top))
		}
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
	if pwm.num == 0 || pwm.num == 2 {
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

	var prescaler uint8

	switch pwm.num {
	case 1:
		prescaler = avr.TCCR1B.Get() & 0x7
	case 3:
		prescaler = avr.TCCR3B.Get() & 0x7
	case 4:
		prescaler = avr.TCCR4B.Get() & 0x7
	case 5:
		prescaler = avr.TCCR5B.Get() & 0x7
	}

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

	switch pwm.num {
	case 1:
		// Warning: this change is not atomic!
		avr.ICR1H.Set(uint8(top >> 8))
		avr.ICR1L.Set(uint8(top))

		// ... and because of that, set the counter back to zero to avoid most of
		// the effects of this non-atomicity.
		avr.TCNT1H.Set(0)
		avr.TCNT1L.Set(0)
	case 3:
		// Warning: this change is not atomic!
		avr.ICR3H.Set(uint8(top >> 8))
		avr.ICR3L.Set(uint8(top))

		// ... and because of that, set the counter back to zero to avoid most of
		// the effects of this non-atomicity.
		avr.TCNT3H.Set(0)
		avr.TCNT3L.Set(0)
	case 4:
		// Warning: this change is not atomic!
		avr.ICR4H.Set(uint8(top >> 8))
		avr.ICR4L.Set(uint8(top))

		// ... and because of that, set the counter back to zero to avoid most of
		// the effects of this non-atomicity.
		avr.TCNT4H.Set(0)
		avr.TCNT4L.Set(0)
	case 5:
		// Warning: this change is not atomic!
		avr.ICR5H.Set(uint8(top >> 8))
		avr.ICR5L.Set(uint8(top))

		// ... and because of that, set the counter back to zero to avoid most of
		// the effects of this non-atomicity.
		avr.TCNT5H.Set(0)
		avr.TCNT5L.Set(0)
	}

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
	switch pwm.num {
	case 1:
		// Timer 1 has a configurable top value.
		low := avr.ICR1L.Get()
		high := avr.ICR1H.Get()
		return uint32(high)<<8 | uint32(low) + 1
	case 3:
		// Timer 3 has a configurable top value.
		low := avr.ICR3L.Get()
		high := avr.ICR3H.Get()
		return uint32(high)<<8 | uint32(low) + 1
	case 4:
		// Timer 4 has a configurable top value.
		low := avr.ICR4L.Get()
		high := avr.ICR4H.Get()
		return uint32(high)<<8 | uint32(low) + 1
	case 5:
		// Timer 5 has a configurable top value.
		low := avr.ICR5L.Get()
		high := avr.ICR5H.Get()
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
	case 3:
		mask := interrupt.Disable()
		low := avr.TCNT3L.Get()
		high := avr.TCNT3H.Get()
		interrupt.Restore(mask)
		return uint32(high)<<8 | uint32(low)
	case 4:
		mask := interrupt.Disable()
		low := avr.TCNT4L.Get()
		high := avr.TCNT4H.Get()
		interrupt.Restore(mask)
		return uint32(high)<<8 | uint32(low)
	case 5:
		mask := interrupt.Disable()
		low := avr.TCNT5L.Get()
		high := avr.TCNT5H.Get()
		interrupt.Restore(mask)
		return uint32(high)<<8 | uint32(low)
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
	case 3:
		prescaler = avr.TCCR3B.Get() & 0x7
	case 4:
		prescaler = avr.TCCR4B.Get() & 0x7
	case 5:
		prescaler = avr.TCCR5B.Get() & 0x7
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
		case PB7: // channel A
			avr.TCCR0A.SetBits(avr.TCCR0A_COM0A1)
			return 0, nil
		case PG5: // channel B
			avr.TCCR0A.SetBits(avr.TCCR0A_COM0B1)
			return 1, nil
		}
	case 1:
		switch pin {
		case PB5: // channel A
			avr.TCCR1A.SetBits(avr.TCCR1A_COM1A1)
			return 0, nil
		case PB6: // channel B
			avr.TCCR1A.SetBits(avr.TCCR1A_COM1B1)
			return 1, nil
		}
	case 2:
		switch pin {
		case PB4: // channel A
			avr.TCCR2A.SetBits(avr.TCCR2A_COM2A1)
			return 0, nil
		case PH6: // channel B
			avr.TCCR2A.SetBits(avr.TCCR2A_COM2B1)
			return 1, nil
		}
	case 3:
		switch pin {
		case PE3: // channel A
			avr.TCCR3A.SetBits(avr.TCCR3A_COM3A1)
			return 0, nil
		case PE4: //channel B
			avr.TCCR3A.SetBits(avr.TCCR3A_COM3B1)
			return 1, nil
		case PE5: //channel C
			avr.TCCR3A.SetBits(avr.TCCR3A_COM3C1)
			return 2, nil
		}
	case 4:
		switch pin {
		case PH3: // channel A
			avr.TCCR4A.SetBits(avr.TCCR4A_COM4A1)
			return 0, nil
		case PH4: //channel B
			avr.TCCR4A.SetBits(avr.TCCR4A_COM4B1)
			return 1, nil
		case PH5: //channel C
			avr.TCCR4A.SetBits(avr.TCCR4A_COM4C1)
			return 2, nil
		}
	case 5:
		switch pin {
		case PL3: // channel A
			avr.TCCR5A.SetBits(avr.TCCR5A_COM5A1)
			return 0, nil
		case PL4: //channel B
			avr.TCCR5A.SetBits(avr.TCCR5A_COM5B1)
			return 1, nil
		case PL5: //channel C
			avr.TCCR5A.SetBits(avr.TCCR5A_COM5C1)
			return 2, nil
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
		case 0: // channel A, PB7
			if inverting {
				avr.PORTB.SetBits(1 << 7) // PB7 high
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0A0)
			} else {
				avr.PORTB.ClearBits(1 << 7) // PB7 low
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0A0)
			}
		case 1: // channel B, PG5
			if inverting {
				avr.PORTG.SetBits(1 << 5) // PG5 high
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0B0)
			} else {
				avr.PORTG.ClearBits(1 << 5) // PG5 low
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0B0)
			}
		}
	case 1:
		// Note: the COM1A0/COM1B0 bit is not set with the configuration below.
		// It will be set the following call to Set(), however.
		switch channel {
		case 0: // channel A, PB5
			if inverting {
				avr.PORTB.SetBits(1 << 5) // PB5 high
			} else {
				avr.PORTB.ClearBits(1 << 5) // PB5 low
			}
		case 1: // channel B, PB6
			if inverting {
				avr.PORTB.SetBits(1 << 6) // PB6 high
			} else {
				avr.PORTB.ClearBits(1 << 6) // PB6 low
			}
		}
	case 2:
		switch channel {
		case 0: // channel A, PB4
			if inverting {
				avr.PORTB.SetBits(1 << 4) // PB4 high
				avr.TCCR2A.SetBits(avr.TCCR2A_COM2A0)
			} else {
				avr.PORTB.ClearBits(1 << 4) // PB4 low
				avr.TCCR2A.ClearBits(avr.TCCR2A_COM2A0)
			}
		case 1: // channel B, PH6
			if inverting {
				avr.PORTH.SetBits(1 << 6) // PH6 high
				avr.TCCR2A.SetBits(avr.TCCR2A_COM2B0)
			} else {
				avr.PORTH.ClearBits(1 << 6) // PH6 low
				avr.TCCR2A.ClearBits(avr.TCCR2A_COM2B0)
			}
		}
	case 3:
		// Note: the COM3A0/COM3B0 bit is not set with the configuration below.
		// It will be set the following call to Set(), however.
		switch channel {
		case 0: // channel A, PE3
			if inverting {
				avr.PORTE.SetBits(1 << 3) // PE3 high
			} else {
				avr.PORTE.ClearBits(1 << 3) // PE3 low
			}
		case 1: // channel B, PE4
			if inverting {
				avr.PORTE.SetBits(1 << 4) // PE4 high
			} else {
				avr.PORTE.ClearBits(1 << 4) // PE4 low
			}
		case 2: // channel C, PE5
			if inverting {
				avr.PORTE.SetBits(1 << 5) // PE4 high
			} else {
				avr.PORTE.ClearBits(1 << 5) // PE4 low
			}
		}
	case 4:
		// Note: the COM3A0/COM3B0 bit is not set with the configuration below.
		// It will be set the following call to Set(), however.
		switch channel {
		case 0: // channel A, PH3
			if inverting {
				avr.PORTH.SetBits(1 << 3) // PH3 high
			} else {
				avr.PORTH.ClearBits(1 << 3) // PH3 low
			}
		case 1: // channel B, PH4
			if inverting {
				avr.PORTH.SetBits(1 << 4) // PH4 high
			} else {
				avr.PORTH.ClearBits(1 << 4) // PH4 low
			}
		case 2: // channel C, PH5
			if inverting {
				avr.PORTH.SetBits(1 << 5) // PH4 high
			} else {
				avr.PORTH.ClearBits(1 << 5) // PH4 low
			}
		}
	case 5:
		// Note: the COM3A0/COM3B0 bit is not set with the configuration below.
		// It will be set the following call to Set(), however.
		switch channel {
		case 0: // channel A, PL3
			if inverting {
				avr.PORTL.SetBits(1 << 3) // PL3 high
			} else {
				avr.PORTL.ClearBits(1 << 3) // PL3 low
			}
		case 1: // channel B, PL4
			if inverting {
				avr.PORTL.SetBits(1 << 4) // PL4 high
			} else {
				avr.PORTL.ClearBits(1 << 4) // PL4 low
			}
		case 2: // channel C, PH5
			if inverting {
				avr.PORTL.SetBits(1 << 5) // PL4 high
			} else {
				avr.PORTL.ClearBits(1 << 5) // PL4 low
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
		case 0: // channel A, PB5
			if value == 0 {
				avr.TCCR1A.ClearBits(avr.TCCR1A_COM1A1 | avr.TCCR1A_COM1A0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR1AH.Set(uint8(value >> 8))
				avr.OCR1AL.Set(uint8(value))
				if avr.PORTB.HasBits(1 << 5) { // is PB1 high?
					// Yes, set the inverting bit.
					avr.TCCR1A.SetBits(avr.TCCR1A_COM1A1 | avr.TCCR1A_COM1A0)
				} else {
					// No, output is non-inverting.
					avr.TCCR1A.SetBits(avr.TCCR1A_COM1A1)
				}
			}
		case 1: // channel B, PB6
			if value == 0 {
				avr.TCCR1A.ClearBits(avr.TCCR1A_COM1B1 | avr.TCCR1A_COM1B0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR1BH.Set(uint8(value >> 8))
				avr.OCR1BL.Set(uint8(value))
				if avr.PORTB.HasBits(1 << 6) { // is PB6 high?
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
	case 3:
		mask := interrupt.Disable()
		switch channel {
		case 0: // channel A, PE3
			if value == 0 {
				avr.TCCR3A.ClearBits(avr.TCCR3A_COM3A1 | avr.TCCR3A_COM3A0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR3AH.Set(uint8(value >> 8))
				avr.OCR3AL.Set(uint8(value))
				if avr.PORTE.HasBits(1 << 3) { // is PE3 high?
					// Yes, set the inverting bit.
					avr.TCCR3A.SetBits(avr.TCCR3A_COM3A1 | avr.TCCR3A_COM3A0)
				} else {
					// No, output is non-inverting.
					avr.TCCR3A.SetBits(avr.TCCR3A_COM3A1)
				}
			}
		case 1: // channel B, PE4
			if value == 0 {
				avr.TCCR3A.ClearBits(avr.TCCR3A_COM3B1 | avr.TCCR3A_COM3B0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR3BH.Set(uint8(value >> 8))
				avr.OCR3BL.Set(uint8(value))
				if avr.PORTE.HasBits(1 << 4) { // is PE4 high?
					// Yes, set the inverting bit.
					avr.TCCR3A.SetBits(avr.TCCR3A_COM3B1 | avr.TCCR3A_COM3B0)
				} else {
					// No, output is non-inverting.
					avr.TCCR3A.SetBits(avr.TCCR3A_COM3B1)
				}
			}
		case 2: // channel C, PE5
			if value == 0 {
				avr.TCCR3A.ClearBits(avr.TCCR3A_COM3C1 | avr.TCCR3A_COM3C0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR3CH.Set(uint8(value >> 8))
				avr.OCR3CL.Set(uint8(value))
				if avr.PORTE.HasBits(1 << 5) { // is PE5 high?
					// Yes, set the inverting bit.
					avr.TCCR3A.SetBits(avr.TCCR3A_COM3C1 | avr.TCCR3A_COM3C0)
				} else {
					// No, output is non-inverting.
					avr.TCCR3A.SetBits(avr.TCCR3A_COM3C1)
				}
			}
		}
		interrupt.Restore(mask)
	case 4:
		mask := interrupt.Disable()
		switch channel {
		case 0: // channel A, PH3
			if value == 0 {
				avr.TCCR4A.ClearBits(avr.TCCR4A_COM4A1 | avr.TCCR4A_COM4A0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR4AH.Set(uint8(value >> 8))
				avr.OCR4AL.Set(uint8(value))
				if avr.PORTH.HasBits(1 << 3) { // is PH3 high?
					// Yes, set the inverting bit.
					avr.TCCR4A.SetBits(avr.TCCR4A_COM4A1 | avr.TCCR4A_COM4A0)
				} else {
					// No, output is non-inverting.
					avr.TCCR4A.SetBits(avr.TCCR4A_COM4A1)
				}
			}
		case 1: // channel B, PH4
			if value == 0 {
				avr.TCCR4A.ClearBits(avr.TCCR4A_COM4B1 | avr.TCCR4A_COM4B0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR4BH.Set(uint8(value >> 8))
				avr.OCR4BL.Set(uint8(value))
				if avr.PORTH.HasBits(1 << 4) { // is PH4 high?
					// Yes, set the inverting bit.
					avr.TCCR4A.SetBits(avr.TCCR4A_COM4B1 | avr.TCCR4A_COM4B0)
				} else {
					// No, output is non-inverting.
					avr.TCCR4A.SetBits(avr.TCCR4A_COM4B1)
				}
			}
		case 2: // channel C, PH5
			if value == 0 {
				avr.TCCR4A.ClearBits(avr.TCCR4A_COM4C1 | avr.TCCR4A_COM4C0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR4CH.Set(uint8(value >> 8))
				avr.OCR4CL.Set(uint8(value))
				if avr.PORTH.HasBits(1 << 5) { // is PH5 high?
					// Yes, set the inverting bit.
					avr.TCCR4A.SetBits(avr.TCCR4A_COM4C1 | avr.TCCR4A_COM4C0)
				} else {
					// No, output is non-inverting.
					avr.TCCR4A.SetBits(avr.TCCR4A_COM4C1)
				}
			}
		}
		interrupt.Restore(mask)
	case 5:
		mask := interrupt.Disable()
		switch channel {
		case 0: // channel A, PL3
			if value == 0 {
				avr.TCCR5A.ClearBits(avr.TCCR5A_COM5A1 | avr.TCCR5A_COM5A0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR5AH.Set(uint8(value >> 8))
				avr.OCR5AL.Set(uint8(value))
				if avr.PORTL.HasBits(1 << 3) { // is PL3 high?
					// Yes, set the inverting bit.
					avr.TCCR5A.SetBits(avr.TCCR5A_COM5A1 | avr.TCCR5A_COM5A0)
				} else {
					// No, output is non-inverting.
					avr.TCCR5A.SetBits(avr.TCCR5A_COM5A1)
				}
			}
		case 1: // channel B, PL4
			if value == 0 {
				avr.TCCR5A.ClearBits(avr.TCCR5A_COM5B1 | avr.TCCR5A_COM5B0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR5BH.Set(uint8(value >> 8))
				avr.OCR5BL.Set(uint8(value))
				if avr.PORTL.HasBits(1 << 4) { // is PL4 high?
					// Yes, set the inverting bit.
					avr.TCCR5A.SetBits(avr.TCCR5A_COM5B1 | avr.TCCR5A_COM5B0)
				} else {
					// No, output is non-inverting.
					avr.TCCR5A.SetBits(avr.TCCR5A_COM5B1)
				}
			}
		case 2: // channel C, PL5
			if value == 0 {
				avr.TCCR5A.ClearBits(avr.TCCR5A_COM5C1 | avr.TCCR5A_COM5C0)
			} else {
				value := uint16(value) - 1 // yes, this is safe (it relies on underflow)
				avr.OCR5CH.Set(uint8(value >> 8))
				avr.OCR5CL.Set(uint8(value))
				if avr.PORTL.HasBits(1 << 5) { // is PL5 high?
					// Yes, set the inverting bit.
					avr.TCCR5A.SetBits(avr.TCCR5A_COM5C1 | avr.TCCR5A_COM5C0)
				} else {
					// No, output is non-inverting.
					avr.TCCR5A.SetBits(avr.TCCR5A_COM5C1)
				}
			}
		}
		interrupt.Restore(mask)
	}
}

// SPI configuration
var SPI0 = SPI{
	spcr: avr.SPCR,
	spdr: avr.SPDR,
	spsr: avr.SPSR,
	sck:  PB1,
	sdo:  PB2,
	sdi:  PB3,
	cs:   PB0}
