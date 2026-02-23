//go:build esp32s3

// ESP32-S3: 2 SAR ADCs, 12-bit hardware; Get() returns 16-bit (raw << 4, 0..65520).
// Pin mapping: ADC1 = GPIO 1..10 (channel = GPIO-1); ADC2 = GPIO 11..20 (channel = GPIO-11).
//
// Registers used (TRM / IDF):
//   SYSTEM:     PERIP_RST_EN0.APB_SARADC_RST, PERIP_CLK_EN0.APB_SARADC_CLK_EN
//   RTC_CNTL:   ANA_CONF.SAR_I2C_PU, I2C_RESET_POR_FORCE_PU
// ADC1 RTC path (oneshot, TRM/IDF):
//   SENS.SAR_MEAS1_MUX.SAR1_DIG_FORCE = 0  → ADC1 under RTC (not digital/APB)
//   SENS.SAR_MEAS1_CTRL2.MEAS1_START_FORCE = 1, SAR1_EN_PAD_FORCE = 1  → SW triggers and selects channel
//   Per conversion: set attenuation (SAR_ATTEN1), channel (SAR1_EN_PAD), then MEAS1_START_SAR 0→1; wait MEAS1_DONE_SAR; read MEAS1_DATA_SAR.
//   SENS.SAR_MEAS1_CTRL1: amp/ref (FORCE_XPD_AMP etc). SAR_MEAS1_CTRL2: MEAS1_DONE_SAR (done), MEAS1_START_SAR (start), MEAS1_DATA_SAR (12-bit result).
//   SAR_READER1_CTRL: CLK_DIV, SAR1_CLK_GATED, SAR1_SAMPLE_NUM. No separate FSM wait for ADC1 in SVD; shared APB FSM_WAIT used.
//   APB_SARADC: FSM_WAIT, CLKM, etc. used for clock/shared logic; ADC2 uses ARB_CTRL.

package machine

import (
	"device/esp"
	"errors"
	"sync"
)

var adcOnce sync.Once

func InitADC() {
	initADC()
}

const (
	ADC0  Pin = 1
	ADC1  Pin = 2
	ADC2  Pin = 3
	ADC3  Pin = 4
	ADC4  Pin = 5
	ADC5  Pin = 6
	ADC6  Pin = 7
	ADC7  Pin = 8
	ADC8  Pin = 9
	ADC9  Pin = 10
	ADC10 Pin = 11
	ADC11 Pin = 12
	ADC12 Pin = 13
	ADC13 Pin = 14
	ADC14 Pin = 15
	ADC15 Pin = 16
	ADC16 Pin = 17
	ADC17 Pin = 18
	ADC18 Pin = 19
	ADC19 Pin = 20
)

const (
	attenDefault = 3 // 11 dB, ~0..3.3 V (IDF ADC_ATTEN_DB_12)
	adc1Delay    = 800
)

func adc1Settle() {
	for i := 0; i < adc1Delay; i++ {
	}
}

func initADCClock() {
	// SYSTEM: reset and enable APB SARADC clock
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_APB_SARADC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(0)
}

func initADC() {
	adcOnce.Do(func() {
		initADCClock()

		// RTC_CNTL.ANA_CONF: SAR I2C pull-up for analog
		esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
		esp.RTC_CNTL.SetANA_CONF_I2C_RESET_POR_FORCE_PU(1)

		// SENS.SAR_POWER: SAR power and clock
		esp.SENS.SetSAR_POWER_XPD_SAR_FORCE_XPD_SAR(3)
		esp.SENS.SetSAR_POWER_XPD_SAR_SARCLK_EN(1)

		// SENS.SAR_MEAS1_CTRL1: ADC1 amp and reference
		esp.SENS.SetSAR_MEAS1_CTRL1_FORCE_XPD_AMP(3)
		esp.SENS.SetSAR_MEAS1_CTRL1_AMP_RST_FB_FORCE(3)
		esp.SENS.SetSAR_MEAS1_CTRL1_AMP_SHORT_REF_FORCE(3)
		esp.SENS.SetSAR_MEAS1_CTRL1_AMP_SHORT_REF_GND_FORCE(3)

		// SENS.SAR_READER1_CTRL / SAR_READER2_CTRL: sample clock and count
		esp.SENS.SetSAR_READER1_CTRL_SAR_SAR1_CLK_DIV(1)
		esp.SENS.SetSAR_READER1_CTRL_SAR_SAR1_CLK_GATED(0)
		esp.SENS.SetSAR_READER1_CTRL_SAR_SAR1_SAMPLE_NUM(1)
		esp.SENS.SetSAR_READER2_CTRL_SAR_SAR2_CLK_DIV(1)
		esp.SENS.SetSAR_READER2_CTRL_SAR_SAR2_CLK_GATED(0)
		esp.SENS.SetSAR_READER2_CTRL_SAR_SAR2_SAMPLE_NUM(1)

		// SENS.SAR_MEAS2_CTRL1: ADC2 FSM wait times
		esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_XPD_WAIT(8)
		esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_RSTB_WAIT(8)
		esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_STANDBY_WAIT(100)
		esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_RSTB_FORCE(3)

		// SENS.SAR_MEAS1_MUX, SAR_MEAS1_CTRL2: ADC1 RTC controller (same path that was working)
		esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)      // 0 = RTC control
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1) // SW triggers conversion
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1) // SW selects channel

		// APB_SARADC: shared FSM, clock, invert; needed for ADC2 arbiter and SAR clock
		esp.APB_SARADC.SetFSM_WAIT_SARADC_XPD_WAIT(8)
		esp.APB_SARADC.SetFSM_WAIT_SARADC_RSTB_WAIT(8)
		esp.APB_SARADC.SetFSM_WAIT_SARADC_STANDBY_WAIT(100)
		esp.APB_SARADC.SetCTRL_SARADC_XPD_SAR_FORCE(3)
		esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_GATED(1)
		esp.APB_SARADC.SetCTRL2_SARADC_SAR1_INV(1)
		esp.APB_SARADC.SetCTRL2_SARADC_SAR2_INV(1)
		esp.APB_SARADC.SetCLKM_CONF_CLK_SEL(2)
		esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_NUM(1)
		esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_B(0)
		esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_A(0)
		esp.APB_SARADC.SetCLKM_CONF_CLK_EN(1)
		esp.APB_SARADC.SetFILTER_CTRL1_FILTER_FACTOR0(0)
		esp.APB_SARADC.SetFILTER_CTRL1_FILTER_FACTOR1(0)
	})
}

func setSensAtten1(ch, atten uint32) {
	// SENS.SAR_ATTEN1: 2 bits per channel
	v := esp.SENS.GetSAR_ATTEN1()
	v &^= 3 << (ch * 2)
	v |= (atten & 3) << (ch * 2)
	esp.SENS.SetSAR_ATTEN1(v)
}

func setSensAtten2(ch, atten uint32) {
	// SENS.SAR_ATTEN2: 2 bits per channel
	v := esp.SENS.GetSAR_ATTEN2()
	v &^= 3 << (ch * 2)
	v |= (atten & 3) << (ch * 2)
	esp.SENS.SetSAR_ATTEN2(v)
}

func (a ADC) Configure(config ADCConfig) error {
	if a.Pin < 1 || a.Pin > 20 {
		return errors.New("invalid ADC pin for ESP32-S3")
	}
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	initADC()
	return nil
}

func (a ADC) Get() uint16 {
	if a.Pin < 1 || a.Pin > 20 {
		return 0
	}
	initADC()
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	adc1 := a.Pin <= 10
	var ch uint32
	if adc1 {
		ch = uint32(a.Pin - 1) // GPIO1→ch0 … GPIO10→ch9
		// RTC controller: must be set before each conversion (adc_ll_set_controller equivalent)
		esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)
		setSensAtten1(ch, attenDefault)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD(1 << ch)
		adc1Settle()
		// One-shot: 0 then 1 with short gap (TRM: MEAS1_START_SAR active when MEAS1_START_FORCE=1)
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(0)
		for i := 0; i < 50; i++ {
		}
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(1)
		for esp.SENS.GetSAR_MEAS1_CTRL2_MEAS1_DONE_SAR() == 0 {
		}
		raw := esp.SENS.GetSAR_MEAS1_CTRL2_MEAS1_DATA_SAR() & 0xfff
		return uint16(raw) << 4
	}
	ch = uint32(a.Pin - 11) // GPIO11→ch0 … GPIO20→ch9
	// SENS.SAR_MEAS2_CTRL2: force SW control, select channel
	esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_FORCE(1)
	esp.SENS.SetSAR_MEAS2_CTRL2_SAR2_EN_PAD_FORCE(1)
	esp.SENS.SetSAR_MEAS2_CTRL2_SAR2_EN_PAD(1 << ch)
	setSensAtten2(ch, attenDefault)
	// APB_SARADC.ARB_CTRL: grant ADC2 to APB for oneshot
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(1)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(1)
	// SENS.SAR_MEAS2_CTRL2.MEAS2_START_SAR: one-shot start
	esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_SAR(0)
	esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_SAR(1)
	for esp.SENS.GetSAR_MEAS2_CTRL2_MEAS2_DONE_SAR() == 0 {
	}
	// SENS.SAR_MEAS2_CTRL2.MEAS2_DATA_SAR: 12-bit result
	raw := esp.SENS.GetSAR_MEAS2_CTRL2_MEAS2_DATA_SAR() & 0xfff
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
	return uint16(raw) << 4
}
