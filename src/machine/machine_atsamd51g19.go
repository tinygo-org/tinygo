// +build sam,atsamd51,atsamd51g19

// Peripheral abstraction layer for the atsamd51.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/60001507C.pdf
//
package machine

import (
	"device/sam"
)

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	if p < 32 {
		return &sam.PORT.GROUP[0].OUTSET.Reg, 1 << uint8(p)
	} else {
		return &sam.PORT.GROUP[1].OUTSET.Reg, 1 << uint8(p-32)
	}
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	if p < 32 {
		return &sam.PORT.GROUP[0].OUTCLR.Reg, 1 << uint8(p)
	} else {
		return &sam.PORT.GROUP[1].OUTCLR.Reg, 1 << uint8(p-32)
	}
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	if p < 32 {
		if high {
			sam.PORT.GROUP[0].OUTSET.Set(1 << uint8(p))
		} else {
			sam.PORT.GROUP[0].OUTCLR.Set(1 << uint8(p))
		}
	} else {
		if high {
			sam.PORT.GROUP[1].OUTSET.Set(1 << uint8(p-32))
		} else {
			sam.PORT.GROUP[1].OUTCLR.Set(1 << uint8(p-32))
		}
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	if p < 32 {
		return (sam.PORT.GROUP[0].IN.Get()>>uint8(p))&1 > 0
	} else {
		return (sam.PORT.GROUP[1].IN.Get()>>(uint8(p)-32))&1 > 0
	}
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	switch config.Mode {
	case PinOutput:
		if p < 32 {
			sam.PORT.GROUP[0].DIRSET.Set(1 << uint8(p))
			// output is also set to input enable so pin can read back its own value
			p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN)
		} else {
			sam.PORT.GROUP[1].DIRSET.Set(1 << uint8(p-32))
			// output is also set to input enable so pin can read back its own value
			p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN)
		}

	case PinInput:
		if p < 32 {
			sam.PORT.GROUP[0].DIRCLR.Set(1 << uint8(p))
			p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN)
		} else {
			sam.PORT.GROUP[1].DIRCLR.Set(1 << uint8(p-32))
			p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN)
		}

	case PinInputPulldown:
		if p < 32 {
			sam.PORT.GROUP[0].DIRCLR.Set(1 << uint8(p))
			sam.PORT.GROUP[0].OUTCLR.Set(1 << uint8(p))
			p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN | sam.PORT_GROUP_PINCFG_PULLEN)
		} else {
			sam.PORT.GROUP[1].DIRCLR.Set(1 << uint8(p-32))
			sam.PORT.GROUP[1].OUTCLR.Set(1 << uint8(p-32))
			p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN | sam.PORT_GROUP_PINCFG_PULLEN)
		}

	case PinInputPullup:
		if p < 32 {
			sam.PORT.GROUP[0].DIRCLR.Set(1 << uint8(p))
			sam.PORT.GROUP[0].OUTSET.Set(1 << uint8(p))
			p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN | sam.PORT_GROUP_PINCFG_PULLEN)
		} else {
			sam.PORT.GROUP[1].DIRCLR.Set(1 << uint8(p-32))
			sam.PORT.GROUP[1].OUTSET.Set(1 << uint8(p-32))
			p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN | sam.PORT_GROUP_PINCFG_PULLEN)
		}

	case PinSERCOM:
		if p&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
			p.setPMux(val | (uint8(PinSERCOM) << sam.PORT_GROUP_PMUX_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
			p.setPMux(val | (uint8(PinSERCOM) << sam.PORT_GROUP_PMUX_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN | sam.PORT_GROUP_PINCFG_DRVSTR | sam.PORT_GROUP_PINCFG_INEN)

	case PinSERCOMAlt:
		if p&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
			p.setPMux(val | (uint8(PinSERCOMAlt) << sam.PORT_GROUP_PMUX_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
			p.setPMux(val | (uint8(PinSERCOMAlt) << sam.PORT_GROUP_PMUX_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN | sam.PORT_GROUP_PINCFG_DRVSTR)

	case PinCom:
		if p&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
			p.setPMux(val | (uint8(PinCom) << sam.PORT_GROUP_PMUX_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
			p.setPMux(val | (uint8(PinCom) << sam.PORT_GROUP_PMUX_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN)
	case PinAnalog:
		if p&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXE_Msk
			p.setPMux(val | (uint8(PinAnalog) << sam.PORT_GROUP_PMUX_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_GROUP_PMUX_PMUXO_Msk
			p.setPMux(val | (uint8(PinAnalog) << sam.PORT_GROUP_PMUX_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_GROUP_PINCFG_PMUXEN | sam.PORT_GROUP_PINCFG_DRVSTR)
	}
}

// getPMux returns the value for the correct PMUX register for this pin.
func (p Pin) getPMux() uint8 {
	switch {
	case p < 32:
		return sam.PORT.GROUP[0].PMUX[uint8(p)>>1].Get()
	case p >= 32 && p < 64:
		return sam.PORT.GROUP[1].PMUX[uint8(p-32)>>1].Get()
	default:
		return 0
	}
}

// setPMux sets the value for the correct PMUX register for this pin.
func (p Pin) setPMux(val uint8) {
	switch {
	case p < 32:
		sam.PORT.GROUP[0].PMUX[uint8(p)>>1].Set(val)
	case p >= 32 && p < 64:
		sam.PORT.GROUP[1].PMUX[uint8(p-32)>>1].Set(val)
	}
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func (p Pin) getPinCfg() uint8 {
	switch {
	case p < 32:
		return sam.PORT.GROUP[0].PINCFG[p].Get()
	case p >= 32 && p <= 64:
		return sam.PORT.GROUP[1].PINCFG[p-32].Get()
	default:
		return 0
	}
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func (p Pin) setPinCfg(val uint8) {
	switch {
	case p < 32:
		sam.PORT.GROUP[0].PINCFG[p].Set(val)
	case p >= 32 && p <= 64:
		sam.PORT.GROUP[1].PINCFG[p-32].Set(val)
	}
}
