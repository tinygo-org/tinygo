// +build sam,atsamd51,atsamd51j19

// Peripheral abstraction layer for the atsamd51.
//
// Datasheet:
// http://ww1.microchip.com/downloads/en/DeviceDoc/SAM_D5xE5x_Family_Data_Sheet_DS60001507F.pdf
//
package machine

import "device/sam"

const HSRAM_SIZE = 0x00030000

// InitPWM initializes the PWM interface.
func InitPWM() {
	// turn on timer clocks used for PWM
	sam.MCLK.APBBMASK.SetBits(sam.MCLK_APBBMASK_TCC0_ | sam.MCLK_APBBMASK_TCC1_)
	sam.MCLK.APBCMASK.SetBits(sam.MCLK_APBCMASK_TCC2_ | sam.MCLK_APBCMASK_TCC3_)
	sam.MCLK.APBDMASK.SetBits(sam.MCLK_APBDMASK_TCC4_)

	//use clock generator 0
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_TCC0].Set((sam.GCLK_PCHCTRL_GEN_GCLK0 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_TCC2].Set((sam.GCLK_PCHCTRL_GEN_GCLK0 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
	sam.GCLK.PCHCTRL[sam.PCHCTRL_GCLK_TCC4].Set((sam.GCLK_PCHCTRL_GEN_GCLK0 << sam.GCLK_PCHCTRL_GEN_Pos) |
		sam.GCLK_PCHCTRL_CHEN)
}

// getTimer returns the timer to be used for PWM on this pin
func (pwm PWM) getTimer() *sam.TCC_Type {
	switch pwm.Pin {
	case PA14:
		return sam.TCC2
	case PA15:
		return sam.TCC2
	case PA16:
		return sam.TCC1
	case PA17:
		return sam.TCC1
	case PA18:
		return sam.TCC1
	case PA19:
		return sam.TCC1
	case PA20:
		return sam.TCC0
	case PA21:
		return sam.TCC0
	case PA22:
		return sam.TCC0
	case PA23:
		return sam.TCC0
	case PB12:
		return sam.TCC3
	case PB13:
		return sam.TCC3
	case PB14:
		return sam.TCC4
	case PB15:
		return sam.TCC4
	case PB16:
		return sam.TCC0
	case PB17:
		return sam.TCC0
	case PB31:
		return sam.TCC4
	default:
		return nil // not supported on this pin
	}
}
