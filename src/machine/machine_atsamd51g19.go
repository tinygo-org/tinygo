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
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	return &sam.PORT.GROUP[group].OUTSET.Reg, 1 << pin_in_group
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	return &sam.PORT.GROUP[group].OUTCLR.Reg, 1 << pin_in_group
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	if high {
		sam.PORT.GROUP[group].OUTSET.Set(1 << pin_in_group)
	} else {
		sam.PORT.GROUP[group].OUTCLR.Set(1 << pin_in_group)
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	return (sam.PORT.GROUP[group].IN.Get()>>pin_in_group)&1 > 0
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	switch config.Mode {
	case PinOutput:
		sam.PORT.GROUP[group].DIRSET.Set(1 << pin_in_group)
		// output is also set to input enable so pin can read back its own value
		p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN)


	case PinInput:
		sam.PORT.GROUP[group].DIRCLR.Set(1 << pin_in_group)
		p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN)

	case PinInputPulldown:
		sam.PORT.GROUP[group].DIRCLR.Set(1 << pin_in_group)
		sam.PORT.GROUP[group].OUTCLR.Set(1 << pin_in_group)
		p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN | sam.PORT_GROUP_PINCFG_PULLEN)

	case PinInputPullup:
		sam.PORT.GROUP[group].DIRCLR.Set(1 << pin_in_group)
		sam.PORT.GROUP[group].OUTSET.Set(1 << pin_in_group)
		p.setPinCfg(sam.PORT_GROUP_PINCFG_INEN | sam.PORT_GROUP_PINCFG_PULLEN)

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
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	return sam.PORT.GROUP[group].PMUX[pin_in_group>>1].Get()
}

// setPMux sets the value for the correct PMUX register for this pin.
func (p Pin) setPMux(val uint8) {
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	sam.PORT.GROUP[group].PMUX[pin_in_group>>1].Set(val)
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func (p Pin) getPinCfg() uint8 {
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	return sam.PORT.GROUP[group].PINCFG[pin_in_group].Get()
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func (p Pin) setPinCfg(val uint8) {
	group := uint8(p) >> 5
	pin_in_group := uint8(p) & 0x1f
	sam.PORT.GROUP[group].PINCFG[pin_in_group].Set(val)
}
