//go:build avrtiny

package machine

import (
	"device/avr"
	"runtime/volatile"
	"unsafe"
)

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

		// Note: the output state (high or low) is as it was before.
	} else {
		// Configure the pin as an input.
		// First set up the configuration that will be used when it is an input.
		pinctrl := uint8(0)
		if config.Mode == PinInputPullup {
			pinctrl |= avr.PORT_PIN0CTRL_PULLUPEN
		}
		// Find the PINxCTRL register for this pin.
		ctrlAddress := (*volatile.Register8)(unsafe.Add(unsafe.Pointer(&port.PIN0CTRL), p%8))
		ctrlAddress.Set(pinctrl)

		// Configure the pin as input (if it wasn't an input pin before).
		port.DIRCLR.Set(mask)
	}
}

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	port, mask := p.getPortMask()
	// As noted above, the PINx register is always two registers below the PORTx
	// register, so we can find it simply by subtracting two from the PORTx
	// register address.
	return (port.IN.Get() & mask) > 0
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
