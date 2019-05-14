// +build sam,atsamd21,atsamd21g18

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
func (p GPIO) PortMaskSet() (*uint32, uint32) {
	if p.Pin < 32 {
		return (*uint32)(&sam.PORT.OUTSET0), 1 << p.Pin
	} else {
		return (*uint32)(&sam.PORT.OUTSET1), 1 << (p.Pin - 32)
	}
}

// Return the register and mask to disable a given port. This can be used to
// implement bit-banged drivers.
func (p GPIO) PortMaskClear() (*uint32, uint32) {
	if p.Pin < 32 {
		return (*uint32)(&sam.PORT.OUTCLR0), 1 << p.Pin
	} else {
		return (*uint32)(&sam.PORT.OUTCLR1), 1 << (p.Pin - 32)
	}
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
	if p.Pin < 32 {
		if high {
			sam.PORT.OUTSET0 = (1 << p.Pin)
		} else {
			sam.PORT.OUTCLR0 = (1 << p.Pin)
		}
	} else {
		if high {
			sam.PORT.OUTSET1 = (1 << (p.Pin - 32))
		} else {
			sam.PORT.OUTCLR1 = (1 << (p.Pin - 32))
		}
	}
}

// Get returns the current value of a GPIO pin.
func (p GPIO) Get() bool {
	if p.Pin < 32 {
		return (sam.PORT.IN0>>p.Pin)&1 > 0
	} else {
		return (sam.PORT.IN1>>(p.Pin-32))&1 > 0
	}
}

// Configure this pin with the given configuration.
func (p GPIO) Configure(config GPIOConfig) {
	switch config.Mode {
	case GPIO_OUTPUT:
		if p.Pin < 32 {
			sam.PORT.DIRSET0 = (1 << p.Pin)
			// output is also set to input enable so pin can read back its own value
			p.setPinCfg(sam.PORT_PINCFG0_INEN)
		} else {
			sam.PORT.DIRSET1 = (1 << (p.Pin - 32))
			// output is also set to input enable so pin can read back its own value
			p.setPinCfg(sam.PORT_PINCFG0_INEN)
		}

	case GPIO_INPUT:
		if p.Pin < 32 {
			sam.PORT.DIRCLR0 = (1 << p.Pin)
			p.setPinCfg(sam.PORT_PINCFG0_INEN)
		} else {
			sam.PORT.DIRCLR1 = (1<<p.Pin - 32)
			p.setPinCfg(sam.PORT_PINCFG0_INEN)
		}

	case GPIO_INPUT_PULLDOWN:
		if p.Pin < 32 {
			sam.PORT.DIRCLR0 = (1 << p.Pin)
			sam.PORT.OUTCLR0 = (1 << p.Pin)
			p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)
		} else {
			sam.PORT.DIRCLR1 = (1<<p.Pin - 32)
			sam.PORT.OUTCLR1 = (1<<p.Pin - 32)
			p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)
		}

	case GPIO_INPUT_PULLUP:
		if p.Pin < 32 {
			sam.PORT.DIRCLR0 = (1 << p.Pin)
			sam.PORT.OUTSET0 = (1 << p.Pin)
			p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)
		} else {
			sam.PORT.DIRCLR1 = (1<<p.Pin - 32)
			sam.PORT.OUTSET1 = (1<<p.Pin - 32)
			p.setPinCfg(sam.PORT_PINCFG0_INEN | sam.PORT_PINCFG0_PULLEN)
		}

	case GPIO_SERCOM:
		if p.Pin&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (GPIO_SERCOM << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (GPIO_SERCOM << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN | sam.PORT_PINCFG0_DRVSTR | sam.PORT_PINCFG0_INEN)

	case GPIO_SERCOM_ALT:
		if p.Pin&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (GPIO_SERCOM_ALT << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (GPIO_SERCOM_ALT << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN | sam.PORT_PINCFG0_DRVSTR)

	case GPIO_COM:
		if p.Pin&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (GPIO_COM << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (GPIO_COM << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN)
	case GPIO_ANALOG:
		if p.Pin&1 > 0 {
			// odd pin, so save the even pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXE_Msk
			p.setPMux(val | (GPIO_ANALOG << sam.PORT_PMUX0_PMUXO_Pos))
		} else {
			// even pin, so save the odd pins
			val := p.getPMux() & sam.PORT_PMUX0_PMUXO_Msk
			p.setPMux(val | (GPIO_COM << sam.PORT_PMUX0_PMUXE_Pos))
		}
		// enable port config
		p.setPinCfg(sam.PORT_PINCFG0_PMUXEN | sam.PORT_PINCFG0_DRVSTR)
	}
}

// getPMux returns the value for the correct PMUX register for this pin.
func getPMux(p uint8) sam.RegValue8 {
	pin := p >> 1
	switch pin {
	case 0:
		return sam.PORT.PMUX0_0
	case 1:
		return sam.PORT.PMUX0_1
	case 2:
		return sam.PORT.PMUX0_2
	case 3:
		return sam.PORT.PMUX0_3
	case 4:
		return sam.PORT.PMUX0_4
	case 5:
		return sam.PORT.PMUX0_5
	case 6:
		return sam.PORT.PMUX0_6
	case 7:
		return sam.PORT.PMUX0_7
	case 8:
		return sam.PORT.PMUX0_8
	case 9:
		return sam.PORT.PMUX0_9
	case 10:
		return sam.PORT.PMUX0_10
	case 11:
		return sam.PORT.PMUX0_11
	case 12:
		return sam.PORT.PMUX0_12
	case 13:
		return sam.PORT.PMUX0_13
	case 14:
		return sam.PORT.PMUX0_14
	case 15:
		return sam.PORT.PMUX0_15
	case 16:
		return sam.RegValue8(sam.PORT.PMUX1_0>>0) & 0xff
	case 17:
		return sam.RegValue8(sam.PORT.PMUX1_0>>8) & 0xff
	case 18:
		return sam.RegValue8(sam.PORT.PMUX1_0>>16) & 0xff
	case 19:
		return sam.RegValue8(sam.PORT.PMUX1_0>>24) & 0xff
	case 20:
		return sam.RegValue8(sam.PORT.PMUX1_4>>0) & 0xff
	case 21:
		return sam.RegValue8(sam.PORT.PMUX1_4>>8) & 0xff
	case 22:
		return sam.RegValue8(sam.PORT.PMUX1_4>>16) & 0xff
	case 23:
		return sam.RegValue8(sam.PORT.PMUX1_4>>24) & 0xff
	case 24:
		return sam.RegValue8(sam.PORT.PMUX1_8>>0) & 0xff
	case 25:
		return sam.RegValue8(sam.PORT.PMUX1_8>>8) & 0xff
	case 26:
		return sam.RegValue8(sam.PORT.PMUX1_8>>16) & 0xff
	case 27:
		return sam.RegValue8(sam.PORT.PMUX1_8>>24) & 0xff
	case 28:
		return sam.RegValue8(sam.PORT.PMUX1_12>>0) & 0xff
	case 29:
		return sam.RegValue8(sam.PORT.PMUX1_12>>8) & 0xff
	case 30:
		return sam.RegValue8(sam.PORT.PMUX1_12>>16) & 0xff
	case 31:
		return sam.RegValue8(sam.PORT.PMUX1_12>>24) & 0xff
	default:
		return 0
	}
}

// setPMux sets the value for the correct PMUX register for this pin.
func setPMux(p uint8, val sam.RegValue8) {
	pin := p >> 1
	switch pin {
	case 0:
		sam.PORT.PMUX0_0 = val
	case 1:
		sam.PORT.PMUX0_1 = val
	case 2:
		sam.PORT.PMUX0_2 = val
	case 3:
		sam.PORT.PMUX0_3 = val
	case 4:
		sam.PORT.PMUX0_4 = val
	case 5:
		sam.PORT.PMUX0_5 = val
	case 6:
		sam.PORT.PMUX0_6 = val
	case 7:
		sam.PORT.PMUX0_7 = val
	case 8:
		sam.PORT.PMUX0_8 = val
	case 9:
		sam.PORT.PMUX0_9 = val
	case 10:
		sam.PORT.PMUX0_10 = val
	case 11:
		sam.PORT.PMUX0_11 = val
	case 12:
		sam.PORT.PMUX0_12 = val
	case 13:
		sam.PORT.PMUX0_13 = val
	case 14:
		sam.PORT.PMUX0_14 = val
	case 15:
		sam.PORT.PMUX0_15 = val
	case 16:
		sam.PORT.PMUX1_0 = (sam.PORT.PMUX1_0 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 17:
		sam.PORT.PMUX1_0 = (sam.PORT.PMUX1_0 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 18:
		sam.PORT.PMUX1_0 = (sam.PORT.PMUX1_0 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 19:
		sam.PORT.PMUX1_0 = (sam.PORT.PMUX1_0 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 20:
		sam.PORT.PMUX1_4 = (sam.PORT.PMUX1_4 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 21:
		sam.PORT.PMUX1_4 = (sam.PORT.PMUX1_4 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 22:
		sam.PORT.PMUX1_4 = (sam.PORT.PMUX1_4 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 23:
		sam.PORT.PMUX1_4 = (sam.PORT.PMUX1_4 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 24:
		sam.PORT.PMUX1_8 = (sam.PORT.PMUX1_8 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 25:
		sam.PORT.PMUX1_8 = (sam.PORT.PMUX1_8 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 26:
		sam.PORT.PMUX1_8 = (sam.PORT.PMUX1_8 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 27:
		sam.PORT.PMUX1_8 = (sam.PORT.PMUX1_8 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 28:
		sam.PORT.PMUX1_12 = (sam.PORT.PMUX1_12 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 29:
		sam.PORT.PMUX1_12 = (sam.PORT.PMUX1_12 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 30:
		sam.PORT.PMUX1_12 = (sam.PORT.PMUX1_12 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 31:
		sam.PORT.PMUX1_12 = (sam.PORT.PMUX1_12 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	}
}

// getPinCfg returns the value for the correct PINCFG register for this pin.
func getPinCfg(p uint8) sam.RegValue8 {
	switch p {
	case 0:
		return sam.PORT.PINCFG0_0
	case 1:
		return sam.PORT.PINCFG0_1
	case 2:
		return sam.PORT.PINCFG0_2
	case 3:
		return sam.PORT.PINCFG0_3
	case 4:
		return sam.PORT.PINCFG0_4
	case 5:
		return sam.PORT.PINCFG0_5
	case 6:
		return sam.PORT.PINCFG0_6
	case 7:
		return sam.PORT.PINCFG0_7
	case 8:
		return sam.PORT.PINCFG0_8
	case 9:
		return sam.PORT.PINCFG0_9
	case 10:
		return sam.PORT.PINCFG0_10
	case 11:
		return sam.PORT.PINCFG0_11
	case 12:
		return sam.PORT.PINCFG0_12
	case 13:
		return sam.PORT.PINCFG0_13
	case 14:
		return sam.PORT.PINCFG0_14
	case 15:
		return sam.PORT.PINCFG0_15
	case 16:
		return sam.PORT.PINCFG0_16
	case 17:
		return sam.PORT.PINCFG0_17
	case 18:
		return sam.PORT.PINCFG0_18
	case 19:
		return sam.PORT.PINCFG0_19
	case 20:
		return sam.PORT.PINCFG0_20
	case 21:
		return sam.PORT.PINCFG0_21
	case 22:
		return sam.PORT.PINCFG0_22
	case 23:
		return sam.PORT.PINCFG0_23
	case 24:
		return sam.PORT.PINCFG0_24
	case 25:
		return sam.PORT.PINCFG0_25
	case 26:
		return sam.PORT.PINCFG0_26
	case 27:
		return sam.PORT.PINCFG0_27
	case 28:
		return sam.PORT.PINCFG0_28
	case 29:
		return sam.PORT.PINCFG0_29
	case 30:
		return sam.PORT.PINCFG0_30
	case 31:
		return sam.PORT.PINCFG0_31
	case 32: // PB00
		return sam.RegValue8(sam.PORT.PINCFG1_0>>0) & 0xff
	case 33: // PB01
		return sam.RegValue8(sam.PORT.PINCFG1_0>>8) & 0xff
	case 34: // PB02
		return sam.RegValue8(sam.PORT.PINCFG1_0>>16) & 0xff
	case 35: // PB03
		return sam.RegValue8(sam.PORT.PINCFG1_0>>24) & 0xff
	case 37: // PB04
		return sam.RegValue8(sam.PORT.PINCFG1_4>>0) & 0xff
	case 38: // PB05
		return sam.RegValue8(sam.PORT.PINCFG1_4>>8) & 0xff
	case 39: // PB06
		return sam.RegValue8(sam.PORT.PINCFG1_4>>16) & 0xff
	case 40: // PB07
		return sam.RegValue8(sam.PORT.PINCFG1_4>>24) & 0xff
	case 41: // PB08
		return sam.RegValue8(sam.PORT.PINCFG1_8>>0) & 0xff
	case 42: // PB09
		return sam.RegValue8(sam.PORT.PINCFG1_8>>8) & 0xff
	case 43: // PB10
		return sam.RegValue8(sam.PORT.PINCFG1_8>>16) & 0xff
	case 44: // PB11
		return sam.RegValue8(sam.PORT.PINCFG1_8>>24) & 0xff
	case 45: // PB12
		return sam.RegValue8(sam.PORT.PINCFG1_12>>0) & 0xff
	case 46: // PB13
		return sam.RegValue8(sam.PORT.PINCFG1_12>>8) & 0xff
	case 47: // PB14
		return sam.RegValue8(sam.PORT.PINCFG1_12>>16) & 0xff
	case 48: // PB15
		return sam.RegValue8(sam.PORT.PINCFG1_12>>24) & 0xff
	case 49: // PB16
		return sam.RegValue8(sam.PORT.PINCFG1_16>>0) & 0xff
	case 50: // PB17
		return sam.RegValue8(sam.PORT.PINCFG1_16>>8) & 0xff
	case 51: // PB18
		return sam.RegValue8(sam.PORT.PINCFG1_16>>16) & 0xff
	case 52: // PB19
		return sam.RegValue8(sam.PORT.PINCFG1_16>>24) & 0xff
	case 53: // PB20
		return sam.RegValue8(sam.PORT.PINCFG1_20>>0) & 0xff
	case 54: // PB21
		return sam.RegValue8(sam.PORT.PINCFG1_20>>8) & 0xff
	case 55: // PB22
		return sam.RegValue8(sam.PORT.PINCFG1_20>>16) & 0xff
	case 56: // PB23
		return sam.RegValue8(sam.PORT.PINCFG1_20>>24) & 0xff
	case 57: // PB24
		return sam.RegValue8(sam.PORT.PINCFG1_24>>0) & 0xff
	case 58: // PB25
		return sam.RegValue8(sam.PORT.PINCFG1_24>>8) & 0xff
	case 59: // PB26
		return sam.RegValue8(sam.PORT.PINCFG1_24>>16) & 0xff
	case 60: // PB27
		return sam.RegValue8(sam.PORT.PINCFG1_24>>24) & 0xff
	case 61: // PB28
		return sam.RegValue8(sam.PORT.PINCFG1_28>>0) & 0xff
	case 62: // PB29
		return sam.RegValue8(sam.PORT.PINCFG1_28>>8) & 0xff
	case 63: // PB30
		return sam.RegValue8(sam.PORT.PINCFG1_28>>16) & 0xff
	case 64: // PB31
		return sam.RegValue8(sam.PORT.PINCFG1_28>>24) & 0xff
	default:
		return 0
	}
}

// setPinCfg sets the value for the correct PINCFG register for this pin.
func setPinCfg(p uint8, val sam.RegValue8) {
	switch p {
	case 0:
		sam.PORT.PINCFG0_0 = val
	case 1:
		sam.PORT.PINCFG0_1 = val
	case 2:
		sam.PORT.PINCFG0_2 = val
	case 3:
		sam.PORT.PINCFG0_3 = val
	case 4:
		sam.PORT.PINCFG0_4 = val
	case 5:
		sam.PORT.PINCFG0_5 = val
	case 6:
		sam.PORT.PINCFG0_6 = val
	case 7:
		sam.PORT.PINCFG0_7 = val
	case 8:
		sam.PORT.PINCFG0_8 = val
	case 9:
		sam.PORT.PINCFG0_9 = val
	case 10:
		sam.PORT.PINCFG0_10 = val
	case 11:
		sam.PORT.PINCFG0_11 = val
	case 12:
		sam.PORT.PINCFG0_12 = val
	case 13:
		sam.PORT.PINCFG0_13 = val
	case 14:
		sam.PORT.PINCFG0_14 = val
	case 15:
		sam.PORT.PINCFG0_15 = val
	case 16:
		sam.PORT.PINCFG0_16 = val
	case 17:
		sam.PORT.PINCFG0_17 = val
	case 18:
		sam.PORT.PINCFG0_18 = val
	case 19:
		sam.PORT.PINCFG0_19 = val
	case 20:
		sam.PORT.PINCFG0_20 = val
	case 21:
		sam.PORT.PINCFG0_21 = val
	case 22:
		sam.PORT.PINCFG0_22 = val
	case 23:
		sam.PORT.PINCFG0_23 = val
	case 24:
		sam.PORT.PINCFG0_24 = val
	case 25:
		sam.PORT.PINCFG0_25 = val
	case 26:
		sam.PORT.PINCFG0_26 = val
	case 27:
		sam.PORT.PINCFG0_27 = val
	case 28:
		sam.PORT.PINCFG0_28 = val
	case 29:
		sam.PORT.PINCFG0_29 = val
	case 30:
		sam.PORT.PINCFG0_30 = val
	case 31:
		sam.PORT.PINCFG0_31 = val
	case 32: // PB00
		sam.PORT.PINCFG1_0 = (sam.PORT.PINCFG1_0 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 33: // PB01
		sam.PORT.PINCFG1_0 = (sam.PORT.PINCFG1_0 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 34: // PB02
		sam.PORT.PINCFG1_0 = (sam.PORT.PINCFG1_0 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 35: // PB03
		sam.PORT.PINCFG1_0 = (sam.PORT.PINCFG1_0 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 36: // PB04
		sam.PORT.PINCFG1_4 = (sam.PORT.PINCFG1_4 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 37: // PB05
		sam.PORT.PINCFG1_4 = (sam.PORT.PINCFG1_4 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 38: // PB06
		sam.PORT.PINCFG1_4 = (sam.PORT.PINCFG1_4 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 39: // PB07
		sam.PORT.PINCFG1_4 = (sam.PORT.PINCFG1_4 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 40: // PB08
		sam.PORT.PINCFG1_8 = (sam.PORT.PINCFG1_8 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 41: // PB09
		sam.PORT.PINCFG1_8 = (sam.PORT.PINCFG1_8 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 42: // PB10
		sam.PORT.PINCFG1_8 = (sam.PORT.PINCFG1_8 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 43: // PB11
		sam.PORT.PINCFG1_8 = (sam.PORT.PINCFG1_8 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 44: // PB12
		sam.PORT.PINCFG1_12 = (sam.PORT.PINCFG1_12 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 45: // PB13
		sam.PORT.PINCFG1_12 = (sam.PORT.PINCFG1_12 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 46: // PB14
		sam.PORT.PINCFG1_12 = (sam.PORT.PINCFG1_12 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 47: // PB15
		sam.PORT.PINCFG1_12 = (sam.PORT.PINCFG1_12 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 48: // PB16
		sam.PORT.PINCFG1_16 = (sam.PORT.PINCFG1_16 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 49: // PB17
		sam.PORT.PINCFG1_16 = (sam.PORT.PINCFG1_16 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 50: // PB18
		sam.PORT.PINCFG1_16 = (sam.PORT.PINCFG1_16 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 51: // PB19
		sam.PORT.PINCFG1_16 = (sam.PORT.PINCFG1_16 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 52: // PB20
		sam.PORT.PINCFG1_20 = (sam.PORT.PINCFG1_20 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 53: // PB21
		sam.PORT.PINCFG1_20 = (sam.PORT.PINCFG1_20 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 54: // PB22
		sam.PORT.PINCFG1_20 = (sam.PORT.PINCFG1_20 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 55: // PB23
		sam.PORT.PINCFG1_20 = (sam.PORT.PINCFG1_20 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 56: // PB24
		sam.PORT.PINCFG1_24 = (sam.PORT.PINCFG1_24 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 57: // PB25
		sam.PORT.PINCFG1_24 = (sam.PORT.PINCFG1_24 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 58: // PB26
		sam.PORT.PINCFG1_24 = (sam.PORT.PINCFG1_24 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 59: // PB27
		sam.PORT.PINCFG1_24 = (sam.PORT.PINCFG1_24 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	case 60: // PB28
		sam.PORT.PINCFG1_28 = (sam.PORT.PINCFG1_28 &^ (0xff << 0)) | (sam.RegValue(val) << 0)
	case 61: // PB29
		sam.PORT.PINCFG1_28 = (sam.PORT.PINCFG1_28 &^ (0xff << 8)) | (sam.RegValue(val) << 8)
	case 62: // PB30
		sam.PORT.PINCFG1_28 = (sam.PORT.PINCFG1_28 &^ (0xff << 16)) | (sam.RegValue(val) << 16)
	case 63: // PB31
		sam.PORT.PINCFG1_28 = (sam.PORT.PINCFG1_28 &^ (0xff << 24)) | (sam.RegValue(val) << 24)
	}
}
