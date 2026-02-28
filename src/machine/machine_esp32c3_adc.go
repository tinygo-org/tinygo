//go:build esp32c3 && !m5stamp_c3

package machine

import (
	"device/esp"
	"errors"
)

var adcInitialized bool

const (
	// ADC attenuation values for ESP32-C3 APB_SARADC.
	// 0 dB  : ~0 .. 1.1 V
	// 11 dB : ~0 .. 3.3 V (matches typical VDD)
	atten0dB  = 0
	atten11dB = 3
)

func InitADC() {
	if adcInitialized {
		return
	}

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

	var c adcC3Calibration
	c.SelfCalibrate()

	adcInitialized = true
}

func (a ADC) Configure(config ADCConfig) error {
	if a.Pin > 5 {
		return errors.New("invalid ADC pin for ESP32-C3")
	}
	InitADC()
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	return nil
}

func (a ADC) Get() uint16 {
	if a.Pin > 5 {
		return 0
	}
	adc1 := a.Pin <= 4
	ch := uint32(a.Pin)
	if !adc1 {
		ch = 0
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(1)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(1)
	}
	// Use 11 dB attenuation so the effective range is approximately 0..3.3 V.
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(ch)
	if adc1 {
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
	} else {
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC2_ONETIME_SAMPLE(1)
	}
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetINT_CLR_APB_SARADC2_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 && esp.APB_SARADC.GetINT_RAW_APB_SARADC2_DONE_INT_RAW() == 0 {
	}
	var raw uint32
	if adc1 {
		raw = esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA()
	} else {
		raw = esp.APB_SARADC.GetSAR2DATA_STATUS_APB_SARADC2_DATA()
	}
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	if adc1 {
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)
	} else {
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC2_ONETIME_SAMPLE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
	}
	//raw &= 0xfff
	return uint16(raw) << 4
}
