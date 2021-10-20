// +build sam,atsamd21,atsamd21g18

// Peripheral abstraction layer for the atsamd21.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/SAMD21-Family-DataSheet-DS40001882D.pdf
//
package machine

import (
	"device/sam"
	"runtime/interrupt"
)

var (
	sercomUSART0 = UART{Buffer: NewRingBuffer(), Bus: sam.SERCOM0_USART, SERCOM: 0}
	sercomUSART1 = UART{Buffer: NewRingBuffer(), Bus: sam.SERCOM1_USART, SERCOM: 1}
	sercomUSART2 = UART{Buffer: NewRingBuffer(), Bus: sam.SERCOM2_USART, SERCOM: 2}
	sercomUSART3 = UART{Buffer: NewRingBuffer(), Bus: sam.SERCOM3_USART, SERCOM: 3}
	sercomUSART4 = UART{Buffer: NewRingBuffer(), Bus: sam.SERCOM4_USART, SERCOM: 4}
	sercomUSART5 = UART{Buffer: NewRingBuffer(), Bus: sam.SERCOM5_USART, SERCOM: 5}

	sercomI2CM0 = &I2C{Bus: sam.SERCOM0_I2CM, SERCOM: 0}
	sercomI2CM1 = &I2C{Bus: sam.SERCOM1_I2CM, SERCOM: 1}
	sercomI2CM2 = &I2C{Bus: sam.SERCOM2_I2CM, SERCOM: 2}
	sercomI2CM3 = &I2C{Bus: sam.SERCOM3_I2CM, SERCOM: 3}
	sercomI2CM4 = &I2C{Bus: sam.SERCOM4_I2CM, SERCOM: 4}
	sercomI2CM5 = &I2C{Bus: sam.SERCOM5_I2CM, SERCOM: 5}

	sercomSPIM0 = SPI{Bus: sam.SERCOM0_SPI, SERCOM: 0}
	sercomSPIM1 = SPI{Bus: sam.SERCOM1_SPI, SERCOM: 1}
	sercomSPIM2 = SPI{Bus: sam.SERCOM2_SPI, SERCOM: 2}
	sercomSPIM3 = SPI{Bus: sam.SERCOM3_SPI, SERCOM: 3}
	sercomSPIM4 = SPI{Bus: sam.SERCOM4_SPI, SERCOM: 4}
	sercomSPIM5 = SPI{Bus: sam.SERCOM5_SPI, SERCOM: 5}
)

func init() {
	sercomUSART0.Interrupt = interrupt.New(sam.IRQ_SERCOM0, sercomUSART0.handleInterrupt)
	sercomUSART1.Interrupt = interrupt.New(sam.IRQ_SERCOM1, sercomUSART1.handleInterrupt)
	sercomUSART2.Interrupt = interrupt.New(sam.IRQ_SERCOM2, sercomUSART2.handleInterrupt)
	sercomUSART3.Interrupt = interrupt.New(sam.IRQ_SERCOM3, sercomUSART3.handleInterrupt)
	sercomUSART4.Interrupt = interrupt.New(sam.IRQ_SERCOM4, sercomUSART4.handleInterrupt)
	sercomUSART5.Interrupt = interrupt.New(sam.IRQ_SERCOM5, sercomUSART5.handleInterrupt)
}

// Return the register and mask to enable a given GPIO pin. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	if p < 32 {
		return &sam.PORT.OUTSET0.Reg, 1 << uint8(p)
	} else {
		return &sam.PORT.OUTSET1.Reg, 1 << uint8(p-32)
	}
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	if p < 32 {
		return &sam.PORT.OUTCLR0.Reg, 1 << uint8(p)
	} else {
		return &sam.PORT.OUTCLR1.Reg, 1 << uint8(p-32)
	}
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	if p < 32 {
		if high {
			sam.PORT.OUTSET0.Set(1 << uint8(p))
		} else {
			sam.PORT.OUTCLR0.Set(1 << uint8(p))
		}
	} else {
		if high {
			sam.PORT.OUTSET1.Set(1 << uint8(p-32))
		} else {
			sam.PORT.OUTCLR1.Set(1 << uint8(p-32))
		}
	}
}

// Get returns the current value of a GPIO pin when configured as an input or as
// an output.
func (p Pin) Get() bool {
	if p < 32 {
		return (sam.PORT.IN0.Get()>>uint8(p))&1 > 0
	} else {
		return (sam.PORT.IN1.Get()>>uint8(p-32))&1 > 0
	}
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	switch config.Mode {
	case PinOutput:
		if p < 32 {
			sam.PORT.DIRSET0.Set(1 << uint8(p))
			// output is also set to input enable so pin can read back its own value
			p.setPinCfg(sam.PORT_PINCFG0_INEN)
		} else {
			sam.PORT.DIRSET1.Set(1 << uint8(p-32))
			// output is also set to input enable so pin can read back its own value
			p.setPinCfg(sam.PORT_PINCFG0_INEN)
		}

	case PinInput:
		if p < 32 {
			sam.PORT.DIRCLR0.Set(1 << uint8(p))
			p.setPinCfg(sam.PORT_PINCFG0_INEN)
		} else {
			sam.PORT.DIRCLR1.Set(1 << uint8(p-32))
			p.setPinCfg(sam.PORT_PINCFG0_INEN)
		}

	case PinInputPulldown:
		if p < 32 {
			sam.PORT.DIRCLR0.Set(1 << uint8(p))
			sam.PORT.OUTCLR0.Set(1 << uint8(p))
			p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)
		} else {
			sam.PORT.DIRCLR1.Set(1 << uint8(p-32))
			sam.PORT.OUTCLR1.Set(1 << uint8(p-32))
			p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)
		}

	case PinInputPullup:
		if p < 32 {
			sam.PORT.DIRCLR0.Set(1 << uint8(p))
			sam.PORT.OUTSET0.Set(1 << uint8(p))
			p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)
		} else {
			sam.PORT.DIRCLR1.Set(1 << uint8(p-32))
			sam.PORT.OUTSET1.Set(1 << uint8(p-32))
			p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)
		}

	case PinSERCOM:
		if p&1 > 0 {
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
		if p&1 > 0 {
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
		if p&1 > 0 {
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
		if p&1 > 0 {
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
	switch uint8(p) >> 1 {
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
	case 16:
		return uint8(sam.PORT.PMUX1_0.Get()>>0) & 0xff
	case 17:
		return uint8(sam.PORT.PMUX1_0.Get()>>8) & 0xff
	case 18:
		return uint8(sam.PORT.PMUX1_0.Get()>>16) & 0xff
	case 19:
		return uint8(sam.PORT.PMUX1_0.Get()>>24) & 0xff
	case 20:
		return uint8(sam.PORT.PMUX1_4.Get()>>0) & 0xff
	case 21:
		return uint8(sam.PORT.PMUX1_4.Get()>>8) & 0xff
	case 22:
		return uint8(sam.PORT.PMUX1_4.Get()>>16) & 0xff
	case 23:
		return uint8(sam.PORT.PMUX1_4.Get()>>24) & 0xff
	case 24:
		return uint8(sam.PORT.PMUX1_8.Get()>>0) & 0xff
	case 25:
		return uint8(sam.PORT.PMUX1_8.Get()>>8) & 0xff
	case 26:
		return uint8(sam.PORT.PMUX1_8.Get()>>16) & 0xff
	case 27:
		return uint8(sam.PORT.PMUX1_8.Get()>>24) & 0xff
	case 28:
		return uint8(sam.PORT.PMUX1_12.Get()>>0) & 0xff
	case 29:
		return uint8(sam.PORT.PMUX1_12.Get()>>8) & 0xff
	case 30:
		return uint8(sam.PORT.PMUX1_12.Get()>>16) & 0xff
	case 31:
		return uint8(sam.PORT.PMUX1_12.Get()>>24) & 0xff
	default:
		return 0
	}
}

// setPMux sets the value for the correct PMUX register for this pin.
func (p Pin) setPMux(val uint8) {
	switch uint8(p) >> 1 {
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
	case 16:
		sam.PORT.PMUX1_0.ReplaceBits(uint32(val), 0xff, 0)
	case 17:
		sam.PORT.PMUX1_0.ReplaceBits(uint32(val), 0xff, 8)
	case 18:
		sam.PORT.PMUX1_0.ReplaceBits(uint32(val), 0xff, 16)
	case 19:
		sam.PORT.PMUX1_0.ReplaceBits(uint32(val), 0xff, 24)
	case 20:
		sam.PORT.PMUX1_4.ReplaceBits(uint32(val), 0xff, 0)
	case 21:
		sam.PORT.PMUX1_4.ReplaceBits(uint32(val), 0xff, 8)
	case 22:
		sam.PORT.PMUX1_4.ReplaceBits(uint32(val), 0xff, 16)
	case 23:
		sam.PORT.PMUX1_4.ReplaceBits(uint32(val), 0xff, 24)
	case 24:
		sam.PORT.PMUX1_8.ReplaceBits(uint32(val), 0xff, 0)
	case 25:
		sam.PORT.PMUX1_8.ReplaceBits(uint32(val), 0xff, 8)
	case 26:
		sam.PORT.PMUX1_8.ReplaceBits(uint32(val), 0xff, 16)
	case 27:
		sam.PORT.PMUX1_8.ReplaceBits(uint32(val), 0xff, 24)
	case 28:
		sam.PORT.PMUX1_12.ReplaceBits(uint32(val), 0xff, 0)
	case 29:
		sam.PORT.PMUX1_12.ReplaceBits(uint32(val), 0xff, 8)
	case 30:
		sam.PORT.PMUX1_12.ReplaceBits(uint32(val), 0xff, 16)
	case 31:
		sam.PORT.PMUX1_12.ReplaceBits(uint32(val), 0xff, 24)
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
	case 32: // PB00
		return uint8(sam.PORT.PINCFG1_0.Get()>>0) & 0xff
	case 33: // PB01
		return uint8(sam.PORT.PINCFG1_0.Get()>>8) & 0xff
	case 34: // PB02
		return uint8(sam.PORT.PINCFG1_0.Get()>>16) & 0xff
	case 35: // PB03
		return uint8(sam.PORT.PINCFG1_0.Get()>>24) & 0xff
	case 37: // PB04
		return uint8(sam.PORT.PINCFG1_4.Get()>>0) & 0xff
	case 38: // PB05
		return uint8(sam.PORT.PINCFG1_4.Get()>>8) & 0xff
	case 39: // PB06
		return uint8(sam.PORT.PINCFG1_4.Get()>>16) & 0xff
	case 40: // PB07
		return uint8(sam.PORT.PINCFG1_4.Get()>>24) & 0xff
	case 41: // PB08
		return uint8(sam.PORT.PINCFG1_8.Get()>>0) & 0xff
	case 42: // PB09
		return uint8(sam.PORT.PINCFG1_8.Get()>>8) & 0xff
	case 43: // PB10
		return uint8(sam.PORT.PINCFG1_8.Get()>>16) & 0xff
	case 44: // PB11
		return uint8(sam.PORT.PINCFG1_8.Get()>>24) & 0xff
	case 45: // PB12
		return uint8(sam.PORT.PINCFG1_12.Get()>>0) & 0xff
	case 46: // PB13
		return uint8(sam.PORT.PINCFG1_12.Get()>>8) & 0xff
	case 47: // PB14
		return uint8(sam.PORT.PINCFG1_12.Get()>>16) & 0xff
	case 48: // PB15
		return uint8(sam.PORT.PINCFG1_12.Get()>>24) & 0xff
	case 49: // PB16
		return uint8(sam.PORT.PINCFG1_16.Get()>>0) & 0xff
	case 50: // PB17
		return uint8(sam.PORT.PINCFG1_16.Get()>>8) & 0xff
	case 51: // PB18
		return uint8(sam.PORT.PINCFG1_16.Get()>>16) & 0xff
	case 52: // PB19
		return uint8(sam.PORT.PINCFG1_16.Get()>>24) & 0xff
	case 53: // PB20
		return uint8(sam.PORT.PINCFG1_20.Get()>>0) & 0xff
	case 54: // PB21
		return uint8(sam.PORT.PINCFG1_20.Get()>>8) & 0xff
	case 55: // PB22
		return uint8(sam.PORT.PINCFG1_20.Get()>>16) & 0xff
	case 56: // PB23
		return uint8(sam.PORT.PINCFG1_20.Get()>>24) & 0xff
	case 57: // PB24
		return uint8(sam.PORT.PINCFG1_24.Get()>>0) & 0xff
	case 58: // PB25
		return uint8(sam.PORT.PINCFG1_24.Get()>>8) & 0xff
	case 59: // PB26
		return uint8(sam.PORT.PINCFG1_24.Get()>>16) & 0xff
	case 60: // PB27
		return uint8(sam.PORT.PINCFG1_24.Get()>>24) & 0xff
	case 61: // PB28
		return uint8(sam.PORT.PINCFG1_28.Get()>>0) & 0xff
	case 62: // PB29
		return uint8(sam.PORT.PINCFG1_28.Get()>>8) & 0xff
	case 63: // PB30
		return uint8(sam.PORT.PINCFG1_28.Get()>>16) & 0xff
	case 64: // PB31
		return uint8(sam.PORT.PINCFG1_28.Get()>>24) & 0xff
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
	case 32: // PB00
		sam.PORT.PINCFG1_0.ReplaceBits(uint32(val), 0xff, 0)
	case 33: // PB01
		sam.PORT.PINCFG1_0.ReplaceBits(uint32(val), 0xff, 8)
	case 34: // PB02
		sam.PORT.PINCFG1_0.ReplaceBits(uint32(val), 0xff, 16)
	case 35: // PB03
		sam.PORT.PINCFG1_0.ReplaceBits(uint32(val), 0xff, 24)
	case 36: // PB04
		sam.PORT.PINCFG1_4.ReplaceBits(uint32(val), 0xff, 0)
	case 37: // PB05
		sam.PORT.PINCFG1_4.ReplaceBits(uint32(val), 0xff, 8)
	case 38: // PB06
		sam.PORT.PINCFG1_4.ReplaceBits(uint32(val), 0xff, 16)
	case 39: // PB07
		sam.PORT.PINCFG1_4.ReplaceBits(uint32(val), 0xff, 24)
	case 40: // PB08
		sam.PORT.PINCFG1_8.ReplaceBits(uint32(val), 0xff, 0)
	case 41: // PB09
		sam.PORT.PINCFG1_8.ReplaceBits(uint32(val), 0xff, 8)
	case 42: // PB10
		sam.PORT.PINCFG1_8.ReplaceBits(uint32(val), 0xff, 16)
	case 43: // PB11
		sam.PORT.PINCFG1_8.ReplaceBits(uint32(val), 0xff, 24)
	case 44: // PB12
		sam.PORT.PINCFG1_12.ReplaceBits(uint32(val), 0xff, 0)
	case 45: // PB13
		sam.PORT.PINCFG1_12.ReplaceBits(uint32(val), 0xff, 8)
	case 46: // PB14
		sam.PORT.PINCFG1_12.ReplaceBits(uint32(val), 0xff, 16)
	case 47: // PB15
		sam.PORT.PINCFG1_12.ReplaceBits(uint32(val), 0xff, 24)
	case 48: // PB16
		sam.PORT.PINCFG1_16.ReplaceBits(uint32(val), 0xff, 0)
	case 49: // PB17
		sam.PORT.PINCFG1_16.ReplaceBits(uint32(val), 0xff, 8)
	case 50: // PB18
		sam.PORT.PINCFG1_16.ReplaceBits(uint32(val), 0xff, 16)
	case 51: // PB19
		sam.PORT.PINCFG1_16.ReplaceBits(uint32(val), 0xff, 24)
	case 52: // PB20
		sam.PORT.PINCFG1_20.ReplaceBits(uint32(val), 0xff, 0)
	case 53: // PB21
		sam.PORT.PINCFG1_20.ReplaceBits(uint32(val), 0xff, 8)
	case 54: // PB22
		sam.PORT.PINCFG1_20.ReplaceBits(uint32(val), 0xff, 16)
	case 55: // PB23
		sam.PORT.PINCFG1_20.ReplaceBits(uint32(val), 0xff, 24)
	case 56: // PB24
		sam.PORT.PINCFG1_24.ReplaceBits(uint32(val), 0xff, 0)
	case 57: // PB25
		sam.PORT.PINCFG1_24.ReplaceBits(uint32(val), 0xff, 8)
	case 58: // PB26
		sam.PORT.PINCFG1_24.ReplaceBits(uint32(val), 0xff, 16)
	case 59: // PB27
		sam.PORT.PINCFG1_24.ReplaceBits(uint32(val), 0xff, 24)
	case 60: // PB28
		sam.PORT.PINCFG1_28.ReplaceBits(uint32(val), 0xff, 0)
	case 61: // PB29
		sam.PORT.PINCFG1_28.ReplaceBits(uint32(val), 0xff, 8)
	case 62: // PB30
		sam.PORT.PINCFG1_28.ReplaceBits(uint32(val), 0xff, 16)
	case 63: // PB31
		sam.PORT.PINCFG1_28.ReplaceBits(uint32(val), 0xff, 24)
	}
}
