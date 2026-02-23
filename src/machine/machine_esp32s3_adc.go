//go:build esp32s3

// ESP32-S3: 2 SAR ADC, 12-bit, result scaled to 16-bit.
// Маппинг: ADC1 — GPIO 1..10 (ch = GPIO-1, SAR1_EN_PAD(1<<ch)); ADC2 — GPIO 11..20 (ch = GPIO-11).
// Т.е. GPIO1=ch0, GPIO2=ch1, GPIO3=ch2 … GPIO10=ch9; GPIO11=ch0 … GPIO20=ch9.

package machine

import (
	"device/esp"
	"errors"
	"sync"
)

var adcOnce sync.Once

func initADCClock() {
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_APB_SARADC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(0)
}

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
	atten0dB       = 0
	attenDefault   = 3 // 12dB, ~0–3.3V (IDF ADC_ATTEN_DB_12)
	debugADC       = false
	adc1Delay      = 200
	adcDummyCount  = 2
	adcWaitTimeout = 100000
)

func adc1Settle() {
	for i := 0; i < adc1Delay; i++ {
	}
}

func adc1Reset() {
	esp.APB_SARADC.SetCTRL_SARADC_START(0)
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetCTRL_SARADC_SAR1_PATT_P_CLEAR(1)
	esp.APB_SARADC.SetDMA_CONF_APB_ADC_RESET_FSM(1)
	esp.APB_SARADC.SetDMA_CONF_APB_ADC_RESET_FSM(0)
	esp.APB_SARADC.SetFILTER_CTRL0_FILTER_RESET(1)
	esp.APB_SARADC.SetFILTER_CTRL0_FILTER_RESET(0)
	for i := 0; i < 300; i++ {
	}
}

func initADC() {
	adcOnce.Do(func() {
		initADCClock()
		esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
		esp.RTC_CNTL.SetANA_CONF_I2C_RESET_POR_FORCE_PU(1)
		esp.SENS.SetSAR_POWER_XPD_SAR_FORCE_XPD_SAR(3)
		esp.SENS.SetSAR_POWER_XPD_SAR_SARCLK_EN(1)
		esp.SENS.SetSAR_MEAS1_CTRL1_FORCE_XPD_AMP(3)
		esp.SENS.SetSAR_MEAS1_CTRL1_AMP_RST_FB_FORCE(3)
		esp.SENS.SetSAR_MEAS1_CTRL1_AMP_SHORT_REF_FORCE(3)
		esp.SENS.SetSAR_MEAS1_CTRL1_AMP_SHORT_REF_GND_FORCE(3)
		esp.SENS.SetSAR_READER1_CTRL_SAR_SAR1_CLK_DIV(1)
		esp.SENS.SetSAR_READER1_CTRL_SAR_SAR1_CLK_GATED(0)
		esp.SENS.SetSAR_READER1_CTRL_SAR_SAR1_SAMPLE_NUM(1)
		esp.SENS.SetSAR_READER2_CTRL_SAR_SAR2_CLK_DIV(1)
		esp.SENS.SetSAR_READER2_CTRL_SAR_SAR2_CLK_GATED(0)
		esp.SENS.SetSAR_READER2_CTRL_SAR_SAR2_SAMPLE_NUM(1)
		esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_XPD_WAIT(8)
		esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_RSTB_WAIT(8)
		esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_STANDBY_WAIT(100)
		esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_RSTB_FORCE(3)
		esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(1)
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
		if debugADC {
			println("ADC S3: init done (ADC1=DIG/APB, ADC2=RTC)")
		}
	})
}

func setSensAtten1(ch, atten uint32) {
	v := esp.SENS.GetSAR_ATTEN1()
	v &^= 3 << (ch * 2)
	v |= (atten & 3) << (ch * 2)
	esp.SENS.SetSAR_ATTEN1(v)
}

func setSensAtten2(ch, atten uint32) {
	v := esp.SENS.GetSAR_ATTEN2()
	v &^= 3 << (ch * 2)
	v |= (atten & 3) << (ch * 2)
	esp.SENS.SetSAR_ATTEN2(v)
}

func (a ADC) Configure(config ADCConfig) error {
	if a.Pin < 1 || a.Pin > 20 {
		return errors.New("invalid ADC pin for ESP32-S3")
	}
	initADC()
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	if a.Pin <= 10 {
		adc1Reset()
	}
	return nil
}

func (a ADC) Get() uint16 {
	if a.Pin < 1 || a.Pin > 20 {
		if debugADC {
			println("ADC S3: invalid pin ", a.Pin)
		}
		return 0
	}
	initADC()
	adc1 := a.Pin <= 10
	var ch uint32
	if adc1 {
		ch = uint32(a.Pin - 1) // GPIO1→0, GPIO2→1, … GPIO10→9
		if debugADC {
			println("ADC S3: pin=", a.Pin, " ADC1 ch=", ch)
		}
		setSensAtten1(ch, attenDefault)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD(1 << ch)
		adc1Settle()
		esp.APB_SARADC.SetCTRL2_SARADC_MEAS_NUM_LIMIT(1)
		esp.APB_SARADC.SetCTRL2_SARADC_MAX_MEAS_NUM(1)
		esp.APB_SARADC.SetCTRL_SARADC_WORK_MODE(0)
		esp.APB_SARADC.SetCTRL_SARADC_SAR_SEL(0)
		esp.APB_SARADC.SetCTRL_SARADC_SAR1_PATT_LEN(0)
		esp.APB_SARADC.SetSAR1_PATT_TAB1_SARADC_SAR1_PATT_TAB1(((attenDefault & 3) << 4) | (ch & 0xF))
		esp.APB_SARADC.SetCTRL_SARADC_START(0)
		esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
		esp.APB_SARADC.SetCTRL_SARADC_SAR1_PATT_P_CLEAR(1)
		adc1Settle()
		var raw uint32
		for i := 0; i <= adcDummyCount; i++ {
			esp.APB_SARADC.SetCTRL_SARADC_START_FORCE(1)
			esp.APB_SARADC.SetCTRL_SARADC_START(1)
			timeout := adcWaitTimeout
			for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
				timeout--
				if timeout == 0 {
					esp.APB_SARADC.SetCTRL_SARADC_START(0)
					return 0
				}
			}
			raw = esp.APB_SARADC.GetAPB_SARADC1_DATA_STATUS_APB_SARADC1_DATA() & 0xfff
			esp.APB_SARADC.SetCTRL_SARADC_START(0)
			esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
			esp.APB_SARADC.SetCTRL_SARADC_SAR1_PATT_P_CLEAR(1)
		}
		if debugADC {
			println("ADC S3: ADC1 raw=", raw)
		}
		return uint16(raw) << 4
	}
	ch = uint32(a.Pin - 11) // GPIO11→0 … GPIO20→9
	if debugADC {
		println("ADC S3: pin=", a.Pin, " ADC2 ch=", ch)
	}
	esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_FORCE(1)
	esp.SENS.SetSAR_MEAS2_CTRL2_SAR2_EN_PAD_FORCE(1)
	esp.SENS.SetSAR_MEAS2_CTRL2_SAR2_EN_PAD(1 << ch)
	setSensAtten2(ch, attenDefault)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(1)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(1)
	var raw uint32
	for i := 0; i <= adcDummyCount; i++ {
		esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_SAR(0)
		esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_SAR(1)
		timeout := adcWaitTimeout
		for esp.SENS.GetSAR_MEAS2_CTRL2_MEAS2_DONE_SAR() == 0 {
			timeout--
			if timeout == 0 {
				esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
				esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
				return 0
			}
		}
		raw = esp.SENS.GetSAR_MEAS2_CTRL2_MEAS2_DATA_SAR() & 0xfff
	}
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
	if debugADC {
		println("ADC S3: ADC2 raw=", raw)
	}
	return uint16(raw) << 4
}
