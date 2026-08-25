//go:build esp32

// ESP32 (Xtensa): SAR ADC1, 12-bit hardware; Get() returns 0..65520 (scaled
// from 12-bit).
//
// Pin mapping is not contiguous, so a lookup is used instead of arithmetic:
// ADC1 channels 0..7 are GPIO36, 37, 38, 39, 32, 33, 34, 35.
//
// ADC1 is driven by the RTC controller under software control: SAR1_DIG_FORCE
// selects the RTC controller, MEAS1_START_FORCE and SAR1_EN_PAD_FORCE hand
// channel selection and triggering to software, then MEAS1_START_SAR 0->1
// starts a conversion and MEAS1_DATA_SAR holds the 12-bit result.
//
// Get() returns raw, uncalibrated values. Unlike the ESP32-C3/S3/C6 drivers
// there is no eFuse or self-calibration step here; accurate voltage mapping
// should be done with a two-point calibration in user code.
//
// ADC2 is deliberately not implemented. On the ESP32 it is shared with the
// Wi-Fi radio and cannot be used reliably while the radio is active.

package machine

import (
	"device/esp"
	"errors"
)

const (
	// 11 dB, the widest input range the SAR offers (IDF ADC_ATTEN_DB_12).
	adcAtten11dB = 3

	// SAR1_BIT_WIDTH / SAR1_SAMPLE_BIT value selecting 12-bit conversions.
	adcWidth12Bit = 3

	// Divider for the ADC's own clock, matching the IDF default for ADC1.
	adcClkDiv = 2
)

var (
	ErrInvalidADCPin = errors.New("invalid ADC pin for ESP32")
)

// InitADC powers up SAR ADC1 and puts it under software control.
func InitADC() {
	// The SAR front end is shared with the hall sensor and an internal
	// amplifier. Both must be off or channel readings pick up their bias.
	esp.SENS.SetSAR_TOUCH_CTRL1_XPD_HALL_FORCE(1)
	esp.SENS.SetSAR_TOUCH_CTRL1_HALL_PHASE_FORCE(1)
	esp.RTC_IO.SetHALL_SENS_XPD_HALL(0)

	esp.SENS.SetSAR_MEAS_WAIT2_FORCE_XPD_AMP(2) // 2 = force power down
	esp.SENS.SetSAR_MEAS_CTRL_AMP_RST_FB_FSM(0)
	esp.SENS.SetSAR_MEAS_CTRL_AMP_SHORT_REF_FSM(0)
	esp.SENS.SetSAR_MEAS_CTRL_AMP_SHORT_REF_GND_FSM(0)
	esp.SENS.SetSAR_MEAS_WAIT1_SAR_AMP_WAIT1(1)
	esp.SENS.SetSAR_MEAS_WAIT1_SAR_AMP_WAIT2(1)
	esp.SENS.SetSAR_MEAS_WAIT2_SAR_AMP_WAIT3(1)

	// Drive ADC1 from the RTC controller rather than the digital/DMA path.
	// The RTC domain is always clocked, so no clock gate has to be opened.
	esp.SENS.SetSAR_READ_CTRL_SAR1_DIG_FORCE(0)
	esp.SENS.SetSAR_MEAS_START1_MEAS1_START_FORCE(1)
	esp.SENS.SetSAR_MEAS_START1_SAR1_EN_PAD_FORCE(1)

	esp.SENS.SetSAR_MEAS_WAIT2_FORCE_XPD_SAR(3) // 3 = force power up

	esp.SENS.SetSAR_START_FORCE_SAR1_BIT_WIDTH(adcWidth12Bit)
	esp.SENS.SetSAR_READ_CTRL_SAR1_SAMPLE_BIT(adcWidth12Bit)
	esp.SENS.SetSAR_READ_CTRL_SAR1_CLK_DIV(adcClkDiv)

	// ADC1 returns the sample inverted; this flips it so readings rise with
	// the input voltage.
	esp.SENS.SetSAR_READ_CTRL_SAR1_DATA_INV(1)
}

// Configure routes the pin to the SAR ADC and sets its attenuation. It returns
// an error if the pin has no ADC1 channel. ADCConfig is accepted for API
// compatibility but its fields are not used; attenuation is fixed at 11 dB.
func (a ADC) Configure(config ADCConfig) error {
	ch, ok := adc1Channel(a.Pin)
	if !ok {
		return ErrInvalidADCPin
	}

	configureADCPad(a.Pin)
	setSensAtten1(ch, adcAtten11dB)

	return nil
}

// Get runs a single conversion and returns the result scaled from the 12-bit
// hardware value to the full 16-bit range, so values run 0..65520. It returns
// 0 if the pin has no ADC1 channel.
func (a ADC) Get() uint16 {
	ch, ok := adc1Channel(a.Pin)
	if !ok {
		return 0
	}

	// SAR1_EN_PAD is a one-hot mask, not a channel index.
	esp.SENS.SetSAR_MEAS_START1_SAR1_EN_PAD(1 << ch)

	// The touch and ULP state machines share the SAR over an internal bus;
	// wait for a conversion already in flight to finish.
	for esp.SENS.GetSAR_SLAVE_ADDR1_MEAS_STATUS() != 0 {
	}

	// The conversion is triggered by the 0->1 edge, so the bit has to be
	// taken low first in case a previous call left it high.
	esp.SENS.SetSAR_MEAS_START1_MEAS1_START_SAR(0)
	esp.SENS.SetSAR_MEAS_START1_MEAS1_START_SAR(1)

	for esp.SENS.GetSAR_MEAS_START1_MEAS1_DONE_SAR() == 0 {
	}

	raw := esp.SENS.GetSAR_MEAS_START1_MEAS1_DATA_SAR()

	return uint16(raw&0xfff) << 4
}

// adc1Channel maps a pin to its ADC1 channel. The ESP32's ADC pins are not
// contiguous, so this cannot be computed from the pin number.
func adc1Channel(p Pin) (uint32, bool) {
	switch p {
	case GPIO36:
		return 0, true
	case GPIO37:
		return 1, true
	case GPIO38:
		return 2, true
	case GPIO39:
		return 3, true
	case GPIO32:
		return 4, true
	case GPIO33:
		return 5, true
	case GPIO34:
		return 6, true
	case GPIO35:
		return 7, true
	}
	return 0, false
}

// configureADCPad hands the pad to the RTC domain and takes it out of digital
// mode, so the SAR sees the analog level.
//
// The pads are spread over three different RTC_IO registers with unrelated
// field names, so each group is handled separately. GPIO34-39 have no internal
// pull resistors, which is why only the GPIO32/33 group disables them.
func configureADCPad(p Pin) {
	switch p {
	case GPIO36:
		esp.RTC_IO.SetSENSOR_PADS_SENSE1_MUX_SEL(1)
		esp.RTC_IO.SetSENSOR_PADS_SENSE1_FUN_SEL(0)
		esp.RTC_IO.SetSENSOR_PADS_SENSE1_FUN_IE(0)
	case GPIO37:
		esp.RTC_IO.SetSENSOR_PADS_SENSE2_MUX_SEL(1)
		esp.RTC_IO.SetSENSOR_PADS_SENSE2_FUN_SEL(0)
		esp.RTC_IO.SetSENSOR_PADS_SENSE2_FUN_IE(0)
	case GPIO38:
		esp.RTC_IO.SetSENSOR_PADS_SENSE3_MUX_SEL(1)
		esp.RTC_IO.SetSENSOR_PADS_SENSE3_FUN_SEL(0)
		esp.RTC_IO.SetSENSOR_PADS_SENSE3_FUN_IE(0)
	case GPIO39:
		esp.RTC_IO.SetSENSOR_PADS_SENSE4_MUX_SEL(1)
		esp.RTC_IO.SetSENSOR_PADS_SENSE4_FUN_SEL(0)
		esp.RTC_IO.SetSENSOR_PADS_SENSE4_FUN_IE(0)
	case GPIO32:
		esp.RTC_IO.SetXTAL_32K_PAD_X32P_MUX_SEL(1)
		esp.RTC_IO.SetXTAL_32K_PAD_X32P_FUN_SEL(0)
		esp.RTC_IO.SetXTAL_32K_PAD_X32P_FUN_IE(0)
		esp.RTC_IO.SetXTAL_32K_PAD_X32P_RUE(0)
		esp.RTC_IO.SetXTAL_32K_PAD_X32P_RDE(0)
	case GPIO33:
		esp.RTC_IO.SetXTAL_32K_PAD_X32N_MUX_SEL(1)
		esp.RTC_IO.SetXTAL_32K_PAD_X32N_FUN_SEL(0)
		esp.RTC_IO.SetXTAL_32K_PAD_X32N_FUN_IE(0)
		esp.RTC_IO.SetXTAL_32K_PAD_X32N_RUE(0)
		esp.RTC_IO.SetXTAL_32K_PAD_X32N_RDE(0)
	case GPIO34:
		// ADC_PAD_ADC1 is a pad name: it is GPIO34, ADC1 channel 6.
		esp.RTC_IO.SetADC_PAD_ADC1_MUX_SEL(1)
		esp.RTC_IO.SetADC_PAD_ADC1_FUN_SEL(0)
		esp.RTC_IO.SetADC_PAD_ADC1_FUN_IE(0)
	case GPIO35:
		// ADC_PAD_ADC2 is a pad name: it is GPIO35, ADC1 channel 7.
		esp.RTC_IO.SetADC_PAD_ADC2_MUX_SEL(1)
		esp.RTC_IO.SetADC_PAD_ADC2_FUN_SEL(0)
		esp.RTC_IO.SetADC_PAD_ADC2_FUN_IE(0)
	default:
		return
	}

	ch, ok := adcRTCGPIO(p)

	if !ok {
		return
	}

	// Take the RTC output driver off the pad so nothing fights the input.
	esp.RTC_IO.SetENABLE_W1TC(1 << ch)
}

// adcRTCGPIO maps a pin to its index within the RTC GPIO block, which is
// numbered independently of the main GPIO matrix.
func adcRTCGPIO(p Pin) (uint32, bool) {
	switch p {
	case GPIO36:
		return 0, true
	case GPIO37:
		return 1, true
	case GPIO38:
		return 2, true
	case GPIO39:
		return 3, true
	case GPIO34:
		return 4, true
	case GPIO35:
		return 5, true
	case GPIO33:
		return 8, true
	case GPIO32:
		return 9, true
	}
	return 0, false
}

// setSensAtten1 sets the 2-bit attenuation field for one ADC1 channel. The
// generated code exposes SAR_ATTEN1 only as a whole register, so the
// read-modify-write is done here.
func setSensAtten1(ch, atten uint32) {
	v := esp.SENS.GetSAR_ATTEN1()
	v &^= 3 << (ch * 2)
	v |= (atten & 3) << (ch * 2)
	esp.SENS.SetSAR_ATTEN1(v)
}
