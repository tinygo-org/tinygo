//go:build esp32c3 && !m5stamp_c3

package machine

import (
	"device/esp"
	"errors"
)

// newRegI2C returns the regI2C configured for ESP32-C3: hostID=0, drefInit=1.
func newRegI2C() regI2C { return regI2C{hostID: 0, drefInit: 1} }

const (
	// ADC attenuation values for ESP32-C3 APB_SARADC.
	// 0 dB  : ~0 .. 1.1 V
	// 11 dB : ~0 .. 3.3 V (matches typical VDD)
	atten0dB  = 0
	atten11dB = 3
)

func InitADC() {
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_APB_SARADC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(0)

	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	esp.RTC_CNTL.SetSENSOR_CTRL_FORCE_XPD_SAR(1)
	esp.APB_SARADC.SetCTRL_SARADC_XPD_SAR_FORCE(1)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_XPD_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_RSTB_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_STANDBY_WAIT(100)
	esp.APB_SARADC.SetCLKM_CONF_CLK_SEL(2)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_NUM(1)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_B(0)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_A(0)
	esp.APB_SARADC.SetCLKM_CONF_CLK_EN(1)

	adcSelfCalibrate()
}

// ESP32-C3: ADC1 = GPIO0–GPIO4 (ch 0–4), ADC2 = GPIO5 (ch 0). ADC2 shares with Wi‑Fi;
// readings may be noisy when Wi‑Fi is active.
func (a ADC) Configure(config ADCConfig) error {
	if a.Pin > 5 {
		return errors.New("invalid ADC pin for ESP32-C3")
	}
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	return nil
}

func (a ADC) Get() uint16 {
	if a.Pin > 5 {
		return 0
	}
	adc1 := a.Pin <= 4
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetINT_CLR_APB_SARADC2_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	var raw uint32
	if adc1 {
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(uint32(a.Pin))
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
		for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
		}
		raw = esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA()
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)
	} else {
		// ADC2: GPIO5 = channel 0. Grant arbiter to ADC2 first, then set channel and start.
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(1)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(1)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(8) // (1<<3)|0 for ADC2 channel 0
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC2_ONETIME_SAMPLE(1)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
		for esp.APB_SARADC.GetINT_RAW_APB_SARADC2_DONE_INT_RAW() == 0 {
		}
		raw = esp.APB_SARADC.GetSAR2DATA_STATUS_APB_SARADC2_DATA()
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC2_ONETIME_SAMPLE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
	}
	return uint16(raw&0xfff) << 4
}

// adcSelfCalibration
const (
	adcCalTimesC3    = 15
	adcCalRtcMagicC3 = uint32(0xADC1C401)
	adcCalInitMinC3  = uint32(1000)
	adcCalInitMaxC3  = uint32(4096)
)

// selfCalibrate sets ADC1/ADC2 init code from RTC or runs self-calibration (GND).
// eFuse is not used: on ESP32-C3 the ADC calibration fields in BLK2 are often unprogrammed.
func adcSelfCalibrate() {
	reg := newRegI2C()
	reg.sarEnable()

	var adc1Code uint32
	if saved, ok := restoreFromRTC(); ok {
		adc1Code = saved
	} else {
		calSetupADC1()
		reg.calibrationInit(0)
		reg.calibrationPrepare(0)
		adc1Code = reg.calibrateBinarySearch(0, adcCalTimesC3, readADC1)
		if adc1Code < adcCalInitMinC3 {
			adc1Code = adcCalInitMinC3
		}
		if adc1Code > adcCalInitMaxC3 {
			adc1Code = adcCalInitMaxC3
		}
		saveToRTC(adc1Code)
		reg.calibrationFinish(0)
	}

	applyADC1Code(reg, adc1Code)
	applyADC2Code(reg, adc1Code)
}

// calSetupADC1 configures APB_SARADC for oneshot sampling on ADC1 channel 0
// with fixed attenuation. This is used only during self‑calibration.
func calSetupADC1() {
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
}

// calSetupADC2 configures APB_SARADC for oneshot sampling on ADC2 (GPIO5, ch 0).
// On C3, onetime_channel = (unit<<3)|channel → ADC2 ch0 = 8.
func calSetupADC2() {
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(8) // (1<<3)|0 for ADC2
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(1)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC2_ONETIME_SAMPLE(1)
}

// readADC1 performs a single ADC1 conversion using the APB_SARADC
// oneshot path and returns the raw 12‑bit result (0..4095).
func readADC1() uint32 {
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA() & 0xfff
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	return uint32(raw)
}

// readADC2 performs a single ADC2 conversion and returns the raw 12‑bit result (0..4095).
func readADC2() uint32 {
	esp.APB_SARADC.SetINT_CLR_APB_SARADC2_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC2_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR2DATA_STATUS_APB_SARADC2_DATA() & 0xfff
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
	return uint32(raw)
}

func restoreFromRTC() (uint32, bool) {
	if esp.RTC_CNTL.GetSTORE0() != adcCalRtcMagicC3 {
		return 0, false
	}
	code := esp.RTC_CNTL.GetSTORE1()
	if code < adcCalInitMinC3 || code > adcCalInitMaxC3 {
		return 0, false
	}
	return code, true
}

func saveToRTC(code uint32) {
	if code < adcCalInitMinC3 || code > adcCalInitMaxC3 {
		return
	}
	esp.RTC_CNTL.SetSTORE0(adcCalRtcMagicC3)
	esp.RTC_CNTL.SetSTORE1(code)
}

// applyADC1Code sets ADC1 init code and finishes calibration.
func applyADC1Code(reg regI2C, code uint32) {
	calSetupADC1()
	reg.calibrationInit(0)
	reg.calibrationPrepare(0)
	reg.setCalibrationParam(0, code)
	reg.calibrationFinish(0)
}

// applyADC2Code sets ADC2 init code and finishes calibration. On C3 eFuse V1
// there is no separate ADC2 calibration; IDF uses ADC1 init code for both units.
func applyADC2Code(reg regI2C, code uint32) {
	reg.calibrationInit(1)
	reg.calibrationPrepare(1)
	reg.setCalibrationParam(1, code)
	reg.calibrationFinish(1)
}
