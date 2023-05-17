//go:build avrtiny

package machine

import "device/avr"

const deviceName = avr.DEVICE

const (
	PinInput PinMode = iota
	PinInputPullup
	PinOutput
)

// Configure sets the pin to input or output.
func (p Pin) Configure(config PinConfig) {
	port, mask := p.getPortMask()

	if config.Mode == PinOutput {
		// set output bit
		port.DIRSET.Set(mask)

		// Note: if the pin was PinInputPullup before, it'll now be high.
		// Otherwise it will be low.
	} else {
		// configure input: clear output bit
		port.DIRCLR.Set(mask)

		if config.Mode == PinInput {
			// No pullup (floating).
			// The transition may be one of the following:
			//   output high -> input pullup -> input (safe: output high and input pullup are similar)
			//   output low  -> input        -> input (safe: no extra transition)
			port.OUTCLR.Set(mask)
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
			port.OUTSET.Set(mask)
		}
	}
}

// Set changes the value of the GPIO pin. The pin must be configured as output.
func (p Pin) Set(high bool) {
	port, mask := p.getPortMask()
	if high {
		port.OUTSET.Set(mask)
	} else {
		port.OUTCLR.Set(mask)
	}
}
