// +build avr

package machine

import (
	"device/avr"
	"runtime/volatile"
	"unsafe"
)

const (
	PinInput PinMode = iota
	PinInputPullup
	PinOutput
)

// In all the AVRs I've looked at, the PIN/DDR/PORT registers followed a regular
// pattern: PINx, DDRx, PORTx in this order without registers in between.
// Therefore, if you know any of them, you can calculate the other two.
//
// For now, I've chosen to let the PORTx register be the one that is returned
// for each specific chip and to calculate the others from that one. Setting an
// output port (done using PORTx) is likely the most common operation and the
// one that is the most time critical. For others, the PINx and DDRx register
// can trivially be calculated using a subtraction.

// Configure sets the pin to input or output.
func (p Pin) Configure(config PinConfig) {
	port, mask := p.getPortMask()
	// The DDRx register can be found by subtracting one from the PORTx
	// register, as this appears to be the case for many (most? all?) AVR chips.
	ddr := (*volatile.Register8)(unsafe.Pointer(uintptr(unsafe.Pointer(port)) - 1))
	if config.Mode == PinOutput {
		// set output bit
		ddr.SetBits(mask)

		// Note: if the pin was PinInputPullup before, it'll now be high.
		// Otherwise it will be low.
	} else {
		// configure input: clear output bit
		ddr.ClearBits(mask)

		if config.Mode == PinInput {
			// No pullup (floating).
			// The transition may be one of the following:
			//   output high -> input pullup -> input (safe: output high and input pullup are similar)
			//   output low  -> input        -> input (safe: no extra transition)
			port.ClearBits(mask)
		} else {
			// Pullup.
			// The transition may be one of the following:
			//   output high -> input pullup -> input pullup (safe: no extra transition)
			//   output low  -> input        -> input pullup (possibly problematic)
			// For the last transition (output low -> input -> input pullup),
			// the transition may be problematic in some cases because there is
			// an intermediate floating state (which may cause irratic
			// interrupts, for example). If this is a problem, the application
			// should set the pin high before configuring it as PinInputPullup.
			// We can't do that here because setting it to high as an
			// intermediate state may have other problems.
			port.SetBits(mask)
		}
	}
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	port, mask := p.getPortMask()
	// As noted above, the PINx register is always two registers below the PORTx
	// register, so we can find it simply by subtracting two from the PORTx
	// register address.
	pin := (*volatile.Register8)(unsafe.Pointer(uintptr(unsafe.Pointer(port)) - 2)) // PINA, PINB, etc
	return (pin.Get() & mask) > 0
}

// Set changes the value of the GPIO pin. The pin must be configured as output.
func (p Pin) Set(value bool) {
	if value { // set bits
		port, mask := p.PortMaskSet()
		port.Set(mask)
	} else { // clear bits
		port, mask := p.PortMaskClear()
		port.Set(mask)
	}
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
//
// Warning: there are no separate pin set/clear registers on the AVR. The
// returned mask is only valid as long as no other pin in the same port has been
// changed.
func (p Pin) PortMaskSet() (*volatile.Register8, uint8) {
	port, mask := p.getPortMask()
	return port, port.Get() | mask
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
//
// Warning: there are no separate pin set/clear registers on the AVR. The
// returned mask is only valid as long as no other pin in the same port has been
// changed.
func (p Pin) PortMaskClear() (*volatile.Register8, uint8) {
	port, mask := p.getPortMask()
	return port, port.Get() &^ mask
}

// InitADC initializes the registers needed for ADC.
func InitADC() {
	// set a2d prescaler so we are inside the desired 50-200 KHz range at 16MHz.
	avr.ADCSRA.SetBits(avr.ADCSRA_ADPS2 | avr.ADCSRA_ADPS1 | avr.ADCSRA_ADPS0)

	// enable a2d conversions
	avr.ADCSRA.SetBits(avr.ADCSRA_ADEN)
}

// Configure configures a ADCPin to be able to be used to read data.
func (a ADC) Configure(ADCConfig) {
	return // no pin specific setup on AVR machine.
}

// Get returns the current value of a ADC pin, in the range 0..0xffff. The AVR
// has an ADC of 10 bits precision so the lower 6 bits will be zero.
func (a ADC) Get() uint16 {
	// set the analog reference (high two bits of ADMUX) and select the
	// channel (low 4 bits), masked to only turn on one ADC at a time.
	// set the ADLAR bit (left-adjusted result) to get a value scaled to 16
	// bits. This has the same effect as shifting the return value left by 6
	// bits.
	avr.ADMUX.Set(avr.ADMUX_REFS0 | avr.ADMUX_ADLAR | (uint8(a.Pin) & 0x07))

	// start the conversion
	avr.ADCSRA.SetBits(avr.ADCSRA_ADSC)

	// ADSC is cleared when the conversion finishes
	for ok := true; ok; ok = avr.ADCSRA.HasBits(avr.ADCSRA_ADSC) {
	}

	return uint16(avr.ADCL.Get()) | uint16(avr.ADCH.Get())<<8
}
