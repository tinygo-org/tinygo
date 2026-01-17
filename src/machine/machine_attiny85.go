//go:build attiny85

package machine

import (
	"device/avr"
	"runtime/volatile"
)

const (
	PB0 Pin = iota
	PB1
	PB2
	PB3
	PB4
	PB5
)

// getPortMask returns the PORTx register and mask for the pin.
func (p Pin) getPortMask() (*volatile.Register8, uint8) {
	// Very simple for the attiny85, which only has a single port.
	return avr.PORTB, 1 << uint8(p)
}

// PWM is one PWM peripheral, which consists of a counter and two output
// channels (that can be connected to two fixed pins). You can set the frequency
// using SetPeriod, but only for all the channels in this PWM peripheral at
// once.
type PWM struct {
	num uint8
}

var (
	Timer0 = PWM{0} // 8 bit timer for PB0 and PB1
	Timer1 = PWM{1} // 8 bit high-speed timer for PB1 and PB4
)

// GTCCR bits for Timer1 that are not defined in the device file
const (
	gtccrPWM1B  = 0x40 // Pulse Width Modulator B Enable
	gtccrCOM1B0 = 0x10 // Comparator B Output Mode bit 0
	gtccrCOM1B1 = 0x20 // Comparator B Output Mode bit 1
)

// Configure enables and configures this PWM.
//
// For Timer0, there is only a limited number of periods available, namely the
// CPU frequency divided by 256 and again divided by 1, 8, 64, 256, or 1024.
// For a MCU running at 8MHz, this would be a period of 32µs, 256µs, 2048µs,
// 8192µs, or 32768µs.
//
// For Timer1, the period is more flexible as it uses OCR1C as the top value.
// Timer1 also supports more prescaler values (1 to 16384).
func (pwm PWM) Configure(config PWMConfig) error {
	switch pwm.num {
	case 0: // Timer/Counter 0 (8-bit)
		// Calculate the timer prescaler.
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

		avr.TCCR0B.Set(prescaler)
		// Set the PWM mode to fast PWM (mode = 3).
		avr.TCCR0A.Set(avr.TCCR0A_WGM00 | avr.TCCR0A_WGM01)

	case 1: // Timer/Counter 1 (8-bit high-speed)
		// Timer1 on ATtiny85 is different from ATmega328:
		// - It's 8-bit with configurable top (OCR1C)
		// - Has more prescaler options (1-16384)
		// - PWM mode is enabled per-channel via PWM1A/PWM1B bits
		var top uint64
		if config.Period == 0 {
			// Use a top appropriate for LEDs.
			top = 0xff
		} else {
			// Calculate top value: top = period * (CPUFrequency / 1e9)
			top = config.Period * (uint64(CPUFrequency()) / 1000000) / 1000
		}

		// Timer1 prescaler values: 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384
		const maxTop = 256
		var prescaler uint8
		switch {
		case top <= maxTop:
			prescaler = 1 // prescaler 1
		case top/2 <= maxTop:
			prescaler = 2 // prescaler 2
			top /= 2
		case top/4 <= maxTop:
			prescaler = 3 // prescaler 4
			top /= 4
		case top/8 <= maxTop:
			prescaler = 4 // prescaler 8
			top /= 8
		case top/16 <= maxTop:
			prescaler = 5 // prescaler 16
			top /= 16
		case top/32 <= maxTop:
			prescaler = 6 // prescaler 32
			top /= 32
		case top/64 <= maxTop:
			prescaler = 7 // prescaler 64
			top /= 64
		case top/128 <= maxTop:
			prescaler = 8 // prescaler 128
			top /= 128
		case top/256 <= maxTop:
			prescaler = 9 // prescaler 256
			top /= 256
		case top/512 <= maxTop:
			prescaler = 10 // prescaler 512
			top /= 512
		case top/1024 <= maxTop:
			prescaler = 11 // prescaler 1024
			top /= 1024
		case top/2048 <= maxTop:
			prescaler = 12 // prescaler 2048
			top /= 2048
		case top/4096 <= maxTop:
			prescaler = 13 // prescaler 4096
			top /= 4096
		case top/8192 <= maxTop:
			prescaler = 14 // prescaler 8192
			top /= 8192
		case top/16384 <= maxTop:
			prescaler = 15 // prescaler 16384
			top /= 16384
		default:
			return ErrPWMPeriodTooLong
		}

		// Set prescaler (CS1[3:0] bits)
		avr.TCCR1.Set(prescaler)
		// Set top value
		avr.OCR1C.Set(uint8(top - 1))
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
	if pwm.num == 0 {
		return ErrPWMPeriodTooLong // Timer0 doesn't support dynamic period
	}

	// Timer1 can adjust period via OCR1C
	var top uint64
	if period == 0 {
		top = 0xff
	} else {
		top = period * (uint64(CPUFrequency()) / 1000000) / 1000
	}

	// Get current prescaler
	prescaler := avr.TCCR1.Get() & 0x0f
	// Timer1 prescaler values follow a power-of-2 pattern:
	// prescaler n maps to divisor 2^(n-1), so we can use a simple shift
	if prescaler > 0 && prescaler <= 15 {
		top >>= (prescaler - 1)
	}

	if top > 256 {
		return ErrPWMPeriodTooLong
	}

	avr.OCR1C.Set(uint8(top - 1))
	avr.TCNT1.Set(0)

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
		// Timer1 has configurable top via OCR1C
		return uint32(avr.OCR1C.Get()) + 1
	}
	// Timer0 goes from 0 to 0xff (256 in total)
	return 256
}

// Counter returns the current counter value of the timer in this PWM
// peripheral. It may be useful for debugging.
func (pwm PWM) Counter() uint32 {
	switch pwm.num {
	case 0:
		return uint32(avr.TCNT0.Get())
	case 1:
		return uint32(avr.TCNT1.Get())
	}
	return 0
}

// Prescaler lookup tables using uint16 (more efficient than uint64 on AVR)
// Timer0 prescaler lookup table (index 0-7 maps to prescaler bits)
var timer0Prescalers = [8]uint16{0, 1, 8, 64, 256, 1024, 0, 0}

// Timer1 prescaler lookup table (index 0-15 maps to prescaler bits)
var timer1Prescalers = [16]uint16{0, 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384}

// Period returns the used PWM period in nanoseconds. It might deviate slightly
// from the configured period due to rounding.
func (pwm PWM) Period() uint64 {
	var prescaler uint64
	switch pwm.num {
	case 0:
		prescalerBits := avr.TCCR0B.Get() & 0x7
		prescaler = uint64(timer0Prescalers[prescalerBits])
		if prescaler == 0 {
			return 0
		}
	case 1:
		prescalerBits := avr.TCCR1.Get() & 0x0f
		prescaler = uint64(timer1Prescalers[prescalerBits])
		if prescaler == 0 {
			return 0
		}
	}
	top := uint64(pwm.Top())
	return prescaler * top * 1000 / uint64(CPUFrequency()/1e6)
}

// Channel returns a PWM channel for the given pin.
func (pwm PWM) Channel(pin Pin) (uint8, error) {
	pin.Configure(PinConfig{Mode: PinOutput})
	pin.Low()
	switch pwm.num {
	case 0:
		switch pin {
		case PB0: // OC0A
			avr.TCCR0A.SetBits(avr.TCCR0A_COM0A1)
			return 0, nil
		case PB1: // OC0B
			avr.TCCR0A.SetBits(avr.TCCR0A_COM0B1)
			return 1, nil
		}
	case 1:
		switch pin {
		case PB1: // OC1A
			// Enable PWM on channel A
			avr.TCCR1.SetBits(avr.TCCR1_PWM1A | avr.TCCR1_COM1A1)
			return 0, nil
		case PB4: // OC1B
			// Enable PWM on channel B (controlled via GTCCR)
			avr.GTCCR.SetBits(gtccrPWM1B | gtccrCOM1B1)
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
func (pwm PWM) SetInverting(channel uint8, inverting bool) {
	switch pwm.num {
	case 0:
		switch channel {
		case 0: // channel A, PB0
			if inverting {
				avr.PORTB.SetBits(1 << 0)
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0A0)
			} else {
				avr.PORTB.ClearBits(1 << 0)
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0A0)
			}
		case 1: // channel B, PB1
			if inverting {
				avr.PORTB.SetBits(1 << 1)
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0B0)
			} else {
				avr.PORTB.ClearBits(1 << 1)
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0B0)
			}
		}
	case 1:
		switch channel {
		case 0: // channel A, PB1
			if inverting {
				avr.PORTB.SetBits(1 << 1)
				avr.TCCR1.SetBits(avr.TCCR1_COM1A0)
			} else {
				avr.PORTB.ClearBits(1 << 1)
				avr.TCCR1.ClearBits(avr.TCCR1_COM1A0)
			}
		case 1: // channel B, PB4
			if inverting {
				avr.PORTB.SetBits(1 << 4)
				avr.GTCCR.SetBits(gtccrCOM1B0)
			} else {
				avr.PORTB.ClearBits(1 << 4)
				avr.GTCCR.ClearBits(gtccrCOM1B0)
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
		switch channel {
		case 0: // channel A, PB0
			if value == 0 {
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0A1)
			} else {
				avr.OCR0A.Set(uint8(value - 1))
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0A1)
			}
		case 1: // channel B, PB1
			if value == 0 {
				avr.TCCR0A.ClearBits(avr.TCCR0A_COM0B1)
			} else {
				avr.OCR0B.Set(uint8(value - 1))
				avr.TCCR0A.SetBits(avr.TCCR0A_COM0B1)
			}
		}
	case 1:
		switch channel {
		case 0: // channel A, PB1
			if value == 0 {
				avr.TCCR1.ClearBits(avr.TCCR1_COM1A1)
			} else {
				avr.OCR1A.Set(uint8(value - 1))
				avr.TCCR1.SetBits(avr.TCCR1_COM1A1)
			}
		case 1: // channel B, PB4
			if value == 0 {
				avr.GTCCR.ClearBits(gtccrCOM1B1)
			} else {
				avr.OCR1B.Set(uint8(value - 1))
				avr.GTCCR.SetBits(gtccrCOM1B1)
			}
		}
	}
}

// SPIConfig is used to store config info for SPI.
type SPIConfig struct {
	Frequency uint32
	LSBFirst  bool
	Mode      uint8
}

// SPI is the USI-based SPI implementation for ATTiny85.
// The ATTiny85 doesn't have dedicated SPI hardware, but uses the USI
// (Universal Serial Interface) in three-wire mode.
//
// Fixed pin mapping (directly controlled by USI hardware):
//   - PB2: SCK (clock)
//   - PB1: DO/MOSI (data out)
//   - PB0: DI/MISO (data in)
//
// Note: CS pin must be managed by the user.
type SPI struct {
	// Delay cycles for frequency control (0 = max speed)
	delayCycles uint16

	// USICR value configured for the selected SPI mode
	usicrValue uint8

	// LSB-first mode (requires software bit reversal)
	lsbFirst bool
}

// SPI0 is the USI-based SPI interface on the ATTiny85
var SPI0 = SPI{}

// Configure sets up the USI for SPI communication.
// Note: The user must configure and control the CS pin separately.
func (s *SPI) Configure(config SPIConfig) error {
	// Configure USI pins (fixed by hardware)
	// PB1 (DO/MOSI) -> OUTPUT
	// PB2 (USCK/SCK) -> OUTPUT
	// PB0 (DI/MISO) -> INPUT
	PB1.Configure(PinConfig{Mode: PinOutput})
	PB2.Configure(PinConfig{Mode: PinOutput})
	PB0.Configure(PinConfig{Mode: PinInput})

	// Reset USI registers
	avr.USIDR.Set(0)
	avr.USISR.Set(0)

	// Configure USI for SPI mode:
	// - USIWM0: Three-wire mode (SPI)
	// - USICS1: External clock source (software controlled via USITC)
	// - USICLK: Clock strobe - enables counter increment on USITC toggle
	// - USICS0: Controls clock phase (CPHA)
	//
	// SPI Modes:
	//   Mode 0 (CPOL=0, CPHA=0): Clock idle low, sample on rising edge
	//   Mode 1 (CPOL=0, CPHA=1): Clock idle low, sample on falling edge
	//   Mode 2 (CPOL=1, CPHA=0): Clock idle high, sample on falling edge
	//   Mode 3 (CPOL=1, CPHA=1): Clock idle high, sample on rising edge
	//
	// For USI, USICS0 controls the sampling edge when USICS1=1:
	//   USICS0=0: Positive edge (rising)
	//   USICS0=1: Negative edge (falling)
	switch config.Mode {
	case Mode0: // CPOL=0, CPHA=0: idle low, sample rising
		PB2.Low()
		s.usicrValue = avr.USICR_USIWM0 | avr.USICR_USICS1 | avr.USICR_USICLK
	case Mode1: // CPOL=0, CPHA=1: idle low, sample falling
		PB2.Low()
		s.usicrValue = avr.USICR_USIWM0 | avr.USICR_USICS1 | avr.USICR_USICS0 | avr.USICR_USICLK
	case Mode2: // CPOL=1, CPHA=0: idle high, sample falling
		PB2.High()
		s.usicrValue = avr.USICR_USIWM0 | avr.USICR_USICS1 | avr.USICR_USICS0 | avr.USICR_USICLK
	case Mode3: // CPOL=1, CPHA=1: idle high, sample rising
		PB2.High()
		s.usicrValue = avr.USICR_USIWM0 | avr.USICR_USICS1 | avr.USICR_USICLK
	default: // Default to Mode 0
		PB2.Low()
		s.usicrValue = avr.USICR_USIWM0 | avr.USICR_USICS1 | avr.USICR_USICLK
	}
	avr.USICR.Set(s.usicrValue)

	// Calculate delay cycles for frequency control
	// Each bit transfer requires 2 clock toggles (rising + falling edge)
	// The loop overhead is approximately 10-15 cycles per toggle on AVR
	// We calculate additional delay cycles needed to achieve the target frequency
	if config.Frequency > 0 && config.Frequency < CPUFrequency()/2 {
		// Cycles per half-period = CPUFrequency / (2 * Frequency)
		// Subtract loop overhead (~15 cycles) to get delay cycles
		cyclesPerHalfPeriod := CPUFrequency() / (2 * config.Frequency)
		const loopOverhead = 15
		if cyclesPerHalfPeriod > loopOverhead {
			s.delayCycles = uint16(cyclesPerHalfPeriod - loopOverhead)
		} else {
			s.delayCycles = 0
		}
	} else {
		// Max speed - no delay
		s.delayCycles = 0
	}

	// Store LSBFirst setting for use in Transfer
	s.lsbFirst = config.LSBFirst

	return nil
}

// reverseByte reverses the bit order of a byte (MSB <-> LSB)
// Used for LSB-first SPI mode since USI hardware only supports MSB-first
func reverseByte(b byte) byte {
	b = (b&0xF0)>>4 | (b&0x0F)<<4
	b = (b&0xCC)>>2 | (b&0x33)<<2
	b = (b&0xAA)>>1 | (b&0x55)<<1
	return b
}

// Transfer performs a single byte SPI transfer (send and receive simultaneously)
// This implements the USI-based SPI transfer using the "clock strobing" technique
func (s *SPI) Transfer(b byte) (byte, error) {
	// For LSB-first mode, reverse the bits before sending
	// USI hardware only supports MSB-first, so we do it in software
	if s.lsbFirst {
		b = reverseByte(b)
	}

	// Load the byte to transmit into the USI Data Register
	avr.USIDR.Set(b)

	// Clear the counter overflow flag by writing 1 to it (AVR quirk)
	// This also resets the 4-bit counter to 0
	avr.USISR.Set(avr.USISR_USIOIF)

	// Clock the data out/in
	// We need 16 clock toggles (8 bits × 2 edges per bit)
	// The USI counter counts each clock edge, so it overflows at 16
	// After 16 toggles, the clock returns to its idle state (set by CPOL in Configure)
	//
	// IMPORTANT: Only toggle USITC here!
	// - USITC toggles the clock pin
	// - The USICR mode bits (USIWM0, USICS1, USICS0, USICLK) were set in Configure()
	// - SetBits preserves those bits and only sets USITC
	if s.delayCycles == 0 {
		// Fast path: no delay, run at maximum speed
		for !avr.USISR.HasBits(avr.USISR_USIOIF) {
			avr.USICR.SetBits(avr.USICR_USITC)
		}
	} else {
		// Frequency-controlled path: add delay between clock toggles
		for !avr.USISR.HasBits(avr.USISR_USIOIF) {
			avr.USICR.SetBits(avr.USICR_USITC)
			// Delay loop for frequency control
			// Each iteration is approximately 3 cycles on AVR (dec, brne)
			for i := s.delayCycles; i > 0; i-- {
				avr.Asm("nop")
			}
		}
	}

	// Get the received byte
	result := avr.USIDR.Get()

	// For LSB-first mode, reverse the received bits
	if s.lsbFirst {
		result = reverseByte(result)
	}

	return result, nil
}
