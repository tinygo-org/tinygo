// +build rp2040

package machine

import (
	"device/rp"
)

func InitADC() {
	// reset ADC
	rp.RESETS.RESET.SetBits(rp.RESETS_RESET_ADC)
	rp.RESETS.RESET.ClearBits(rp.RESETS_RESET_ADC)
	for !rp.RESETS.RESET_DONE.HasBits(rp.RESETS_RESET_ADC) {
	}

	// enable ADC
	rp.ADC.CS.Set(rp.ADC_CS_EN)

	waitForReady()
}

// Configure configures a ADC pin to be able to be used to read data.
func (a ADC) Configure(config ADCConfig) {
	switch a.Pin {
	case GP26, GP27, GP28, GP29:
		a.Pin.Configure(PinConfig{Mode: PinAnalog})
	default:
		// invalid ADC
		return
	}
}

func (a ADC) Get() uint16 {
	rp.ADC.CS.SetBits(uint32(a.getADCChannel()) << rp.ADC_CS_AINSEL_Pos)
	rp.ADC.CS.SetBits(rp.ADC_CS_START_ONCE)

	waitForReady()

	// rp2040 uses 12-bit sampling, so scale to 16-bit
	return uint16(rp.ADC.RESULT.Get() << 4)
}

func waitForReady() {
	for !rp.ADC.CS.HasBits(rp.ADC_CS_READY) {
	}
}

func (a ADC) getADCChannel() uint8 {
	switch a.Pin {
	case GP26:
		return 0
	case GP27:
		return 1
	case GP28:
		return 2
	case GP29:
		return 3
	default:
		return 0
	}
}
