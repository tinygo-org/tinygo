// +build sam,atsamd21,atsamd21e18

// Peripheral abstraction layer for the atsamd21.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/SAMD21-Family-DataSheet-DS40001882D.pdf
//
package machine

import (
	"device/sam"
)

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	return &sam.PORT.OUTSET0.Reg, 1 << uint8(p)
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	return &sam.PORT.OUTCLR0.Reg, 1 << uint8(p)
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	if high {
		sam.PORT.OUTSET0.Set(1 << uint8(p))
	} else {
		sam.PORT.OUTCLR0.Set(1 << uint8(p))
	}
}

// Get returns the current value of a GPIO pin when configured as an input or as
// an output.
func (p Pin) Get() bool {
	return (sam.PORT.IN0.Get()>>uint8(p))&1 > 0
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	switch config.Mode {
	case PinOutput:
		sam.PORT.DIRSET0.Set(1 << uint8(p))
		// output is also set to input enable so pin can read back its own value
		p.setPinCfg(sam.PORT_PINCFG0_INEN)

	case PinInput:
		sam.PORT.DIRCLR0.Set(1 << uint8(p))
		p.setPinCfg(sam.PORT_PINCFG0_INEN)

	case PinInputPulldown:
		sam.PORT.DIRCLR0.Set(1 << uint8(p))
		sam.PORT.OUTCLR0.Set(1 << uint8(p))
		p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)

	case PinInputPullup:
		sam.PORT.DIRCLR0.Set(1 << uint8(p))
		sam.PORT.OUTSET0.Set(1 << uint8(p))
		p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)

	case PinSERCOM:
		if uint8(p)&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (uint8(PinSERCOM) << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (uint8(PinSERCOM) << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN | sam.PORT_PINCFG0_DRVSTR | sam.PORT_PINCFG0_INEN)

	case PinSERCOMAlt:
		if uint8(p)&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (uint8(PinSERCOMAlt) << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (uint8(PinSERCOMAlt) << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN | sam.PORT_PINCFG0_DRVSTR)

	case PinCom:
		if uint8(p)&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (uint8(PinCom) << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (uint8(PinCom) << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN)
	case PinAnalog:
		if uint8(p)&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (uint8(PinAnalog) << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (uint8(PinAnalog) << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN | sam.PORT_PINCFG0_DRVSTR)
	}
}

// getPMux returns the value for the correct PMUX register for this pin.
func (p Pin) getPMux() uint8 {
	switch p >> 1 {
	case 0:
		return sam.PORT.PMUX0_0.Get()
	case 1:
		return sam.PORT.PMUX0_1.Get()
	case 2:
		return sam.PORT.PMUX0_2.Get()
	case 3:
		return sam.PORT.PMUX0_3.Get()
	case 4:
		return sam.PORT.PMUX0_4.Get()
	case 5:
		return sam.PORT.PMUX0_5.Get()
	case 6:
		return sam.PORT.PMUX0_6.Get()
	case 7:
		return sam.PORT.PMUX0_7.Get()
	case 8:
		return sam.PORT.PMUX0_8.Get()
	case 9:
		return sam.PORT.PMUX0_9.Get()
	case 10:
		return sam.PORT.PMUX0_10.Get()
	case 11:
		return sam.PORT.PMUX0_11.Get()
	case 12:
		return sam.PORT.PMUX0_12.Get()
	case 13:
		return sam.PORT.PMUX0_13.Get()
	case 14:
		return sam.PORT.PMUX0_14.Get()
	case 15:
		return sam.PORT.PMUX0_15.Get()
	default:
		return 0
	}
}

// setPMux sets the value for the correct PMUX register for this pin.
func (p Pin) setPMux(val uint8) {
	switch p >> 1 {
	case 0:
		sam.PORT.PMUX0_0.Set(val)
	case 1:
		sam.PORT.PMUX0_1.Set(val)
	case 2:
		sam.PORT.PMUX0_2.Set(val)
	case 3:
		sam.PORT.PMUX0_3.Set(val)
	case 4:
		sam.PORT.PMUX0_4.Set(val)
	case 5:
		sam.PORT.PMUX0_5.Set(val)
	case 6:
		sam.PORT.PMUX0_6.Set(val)
	case 7:
		sam.PORT.PMUX0_7.Set(val)
	case 8:
		sam.PORT.PMUX0_8.Set(val)
	case 9:
		sam.PORT.PMUX0_9.Set(val)
	case 10:
		sam.PORT.PMUX0_10.Set(val)
	case 11:
		sam.PORT.PMUX0_11.Set(val)
	case 12:
		sam.PORT.PMUX0_12.Set(val)
	case 13:
		sam.PORT.PMUX0_13.Set(val)
	case 14:
		sam.PORT.PMUX0_14.Set(val)
	case 15:
		sam.PORT.PMUX0_15.Set(val)
	}
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func (p Pin) getPinCfg() uint8 {
	switch p {
	case 0:
		return sam.PORT.PINCFG0_0.Get()
	case 1:
		return sam.PORT.PINCFG0_1.Get()
	case 2:
		return sam.PORT.PINCFG0_2.Get()
	case 3:
		return sam.PORT.PINCFG0_3.Get()
	case 4:
		return sam.PORT.PINCFG0_4.Get()
	case 5:
		return sam.PORT.PINCFG0_5.Get()
	case 6:
		return sam.PORT.PINCFG0_6.Get()
	case 7:
		return sam.PORT.PINCFG0_7.Get()
	case 8:
		return sam.PORT.PINCFG0_8.Get()
	case 9:
		return sam.PORT.PINCFG0_9.Get()
	case 10:
		return sam.PORT.PINCFG0_10.Get()
	case 11:
		return sam.PORT.PINCFG0_11.Get()
	case 12:
		return sam.PORT.PINCFG0_12.Get()
	case 13:
		return sam.PORT.PINCFG0_13.Get()
	case 14:
		return sam.PORT.PINCFG0_14.Get()
	case 15:
		return sam.PORT.PINCFG0_15.Get()
	case 16:
		return sam.PORT.PINCFG0_16.Get()
	case 17:
		return sam.PORT.PINCFG0_17.Get()
	case 18:
		return sam.PORT.PINCFG0_18.Get()
	case 19:
		return sam.PORT.PINCFG0_19.Get()
	case 20:
		return sam.PORT.PINCFG0_20.Get()
	case 21:
		return sam.PORT.PINCFG0_21.Get()
	case 22:
		return sam.PORT.PINCFG0_22.Get()
	case 23:
		return sam.PORT.PINCFG0_23.Get()
	case 24:
		return sam.PORT.PINCFG0_24.Get()
	case 25:
		return sam.PORT.PINCFG0_25.Get()
	case 26:
		return sam.PORT.PINCFG0_26.Get()
	case 27:
		return sam.PORT.PINCFG0_27.Get()
	case 28:
		return sam.PORT.PINCFG0_28.Get()
	case 29:
		return sam.PORT.PINCFG0_29.Get()
	case 30:
		return sam.PORT.PINCFG0_30.Get()
	case 31:
		return sam.PORT.PINCFG0_31.Get()
	default:
		return 0
	}
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func (p Pin) setPinCfg(val uint8) {
	switch p {
	case 0:
		sam.PORT.PINCFG0_0.Set(val)
	case 1:
		sam.PORT.PINCFG0_1.Set(val)
	case 2:
		sam.PORT.PINCFG0_2.Set(val)
	case 3:
		sam.PORT.PINCFG0_3.Set(val)
	case 4:
		sam.PORT.PINCFG0_4.Set(val)
	case 5:
		sam.PORT.PINCFG0_5.Set(val)
	case 6:
		sam.PORT.PINCFG0_6.Set(val)
	case 7:
		sam.PORT.PINCFG0_7.Set(val)
	case 8:
		sam.PORT.PINCFG0_8.Set(val)
	case 9:
		sam.PORT.PINCFG0_9.Set(val)
	case 10:
		sam.PORT.PINCFG0_10.Set(val)
	case 11:
		sam.PORT.PINCFG0_11.Set(val)
	case 12:
		sam.PORT.PINCFG0_12.Set(val)
	case 13:
		sam.PORT.PINCFG0_13.Set(val)
	case 14:
		sam.PORT.PINCFG0_14.Set(val)
	case 15:
		sam.PORT.PINCFG0_15.Set(val)
	case 16:
		sam.PORT.PINCFG0_16.Set(val)
	case 17:
		sam.PORT.PINCFG0_17.Set(val)
	case 18:
		sam.PORT.PINCFG0_18.Set(val)
	case 19:
		sam.PORT.PINCFG0_19.Set(val)
	case 20:
		sam.PORT.PINCFG0_20.Set(val)
	case 21:
		sam.PORT.PINCFG0_21.Set(val)
	case 22:
		sam.PORT.PINCFG0_22.Set(val)
	case 23:
		sam.PORT.PINCFG0_23.Set(val)
	case 24:
		sam.PORT.PINCFG0_24.Set(val)
	case 25:
		sam.PORT.PINCFG0_25.Set(val)
	case 26:
		sam.PORT.PINCFG0_26.Set(val)
	case 27:
		sam.PORT.PINCFG0_27.Set(val)
	case 28:
		sam.PORT.PINCFG0_28.Set(val)
	case 29:
		sam.PORT.PINCFG0_29.Set(val)
	case 30:
		sam.PORT.PINCFG0_30.Set(val)
	case 31:
		sam.PORT.PINCFG0_31.Set(val)
	}
}
