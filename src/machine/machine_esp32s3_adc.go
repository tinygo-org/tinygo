//go:build esp32s3

// ESP32-S3: 2 SAR ADCs, 12-bit hardware; Get() returns 16-bit (raw << 4, 0..65520).
// Pin mapping: ADC1 = GPIO 1..10 (channel = GPIO-1); ADC2 = GPIO 11..20 (channel = GPIO-11).
// Get() returns raw, uncalibrated ADC values; accurate 0–3.3V mapping should be done
// either by a two-point calibration in user code or by using the eFuse-based
// calibration logic (see IDF adc_cali / our ADCSelfCalibrate implementation).
//
// Registers used (TRM / IDF):
//   SYSTEM:     PERIP_RST_EN0.APB_SARADC_RST, PERIP_CLK_EN0.APB_SARADC_CLK_EN
//   RTC_CNTL:   ANA_CONF.SAR_I2C_PU, I2C_RESET_POR_FORCE_PU
// ADC1 RTC path (oneshot, TRM/IDF):
//   SENS.SAR_MEAS1_MUX.SAR1_DIG_FORCE = 0  → ADC1 under RTC (not digital/APB)
//   SENS.SAR_MEAS1_CTRL2.MEAS1_START_FORCE = 1, SAR1_EN_PAD_FORCE = 1  → SW triggers and selects channel
//   Per conversion: set attenuation (SAR_ATTEN1), channel (SAR1_EN_PAD), then MEAS1_START_SAR 0→1; wait MEAS1_DONE_SAR; read MEAS1_DATA_SAR.
//   SENS.SAR_MEAS1_CTRL1: amp/ref (FORCE_XPD_AMP etc). SAR_MEAS1_CTRL2: MEAS1_DONE_SAR (done), MEAS1_START_SAR (start), MEAS1_DATA_SAR (12-bit result).
//   APB_SARADC: FSM_WAIT, CLKM, etc. used for clock/shared logic; ADC2 uses ARB_CTRL.

package machine

import (
	"device/esp"
	"errors"
)

var adcInitialized bool
var digiRefMv uint32

func InitADC() {
	if adcInitialized {
		return
	}

	// SYSTEM: reset and enable APB_SARADC clock so SAR registers are accessible.
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_APB_SARADC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(0)

	// SENS.SAR_PERI_CLK_GATE_CONF: enable SENS SAR peripheral clock (matches Arduino/IDF runtime state).
	esp.SENS.SetSAR_PERI_CLK_GATE_CONF_SARADC_CLK_EN(1)

	// RTC_CNTL.ANA_CONF: keep internal SAR I2C (RegI2C analog bus) powered and out of reset.
	esp.RTC_CNTL.SetANA_CONF_I2C_RESET_POR_FORCE_PD(0)
	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	esp.RTC_CNTL.SetANA_CONF_I2C_RESET_POR_FORCE_PU(1)

	// SENS.SAR_POWER: power up SAR analog block and enable SAR internal clock.
	esp.SENS.SetSAR_POWER_XPD_SAR_FORCE_XPD_SAR(3)
	esp.SENS.SetSAR_POWER_XPD_SAR_SARCLK_EN(1)

	// SENS.SAR_MEAS1_CTRL1: force ADC1 front-end amplifier and reference on in RTC oneshot mode.
	esp.SENS.SetSAR_MEAS1_CTRL1_FORCE_XPD_AMP(3)
	esp.SENS.SetSAR_MEAS1_CTRL1_AMP_RST_FB_FORCE(3)
	esp.SENS.SetSAR_MEAS1_CTRL1_AMP_SHORT_REF_FORCE(3)
	esp.SENS.SetSAR_MEAS1_CTRL1_AMP_SHORT_REF_GND_FORCE(3)

	// SENS.SAR_AMP_CTRL1/2: amplifier/reference settling timings (same as cold-boot defaults).
	esp.SENS.SetSAR_AMP_CTRL1_SAR_AMP_WAIT1(10)
	esp.SENS.SetSAR_AMP_CTRL1_SAR_AMP_WAIT2(10)
	esp.SENS.SetSAR_AMP_CTRL2_SAR_XPD_SAR_AMP_FSM_IDLE(1)
	esp.SENS.SetSAR_AMP_CTRL2_SAR_AMP_SHORT_REF_GND_FSM_IDLE(1)

	// SENS.SAR_MEAS2_CTRL1: ADC2 FSM wait timings for power-up/reset/standby.
	esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_XPD_WAIT(8)
	esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_RSTB_WAIT(8)
	esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_STANDBY_WAIT(100)
	esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_RSTB_FORCE(3)

	// SENS.SAR_MEAS1_MUX / SAR_MEAS1_CTRL2: route ADC1 to RTC controller and use SW to select channel/start.
	esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)      // 0 = controlled by RTC/SENS, not digital/APB.
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1) // SW triggers conversion.
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1) // SW selects which ADC1 pad is enabled.

	// APB_SARADC: shared FSM/clock config used by both ADC units and the ADC2 arbiter.
	esp.APB_SARADC.SetFSM_WAIT_SARADC_XPD_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_RSTB_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_STANDBY_WAIT(100)
	esp.APB_SARADC.SetCTRL_SARADC_XPD_SAR_FORCE(3)
	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_GATED(1)
	esp.APB_SARADC.SetCTRL2_SARADC_SAR1_INV(0)
	esp.APB_SARADC.SetCTRL2_SARADC_SAR2_INV(0)
	esp.APB_SARADC.SetCLKM_CONF_CLK_SEL(2)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_NUM(1)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_B(0)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_A(0)
	esp.APB_SARADC.SetCLKM_CONF_CLK_EN(1)
	esp.APB_SARADC.SetFILTER_CTRL1_FILTER_FACTOR0(0)
	esp.APB_SARADC.SetFILTER_CTRL1_FILTER_FACTOR1(0)

	adcCal := ADCDefaultCalibration{}
	adcCal.SelfCalibrate()
	digiRefMv = adcCal.GetDigiRef()

	adcInitialized = true
}

const (
	attenDefault = 3 // 11 dB, ~0..3.3 V (IDF ADC_ATTEN_DB_12)
)

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
	InitADC()

	return nil
}

func (a ADC) Get() uint16 {
	if a.Pin < 1 || a.Pin > 20 {
		return 0
	}

	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	adc1 := a.Pin <= 10
	var ch uint32
	if adc1 {
		ch = uint32(a.Pin - 1) // GPIO1→ch0 … GPIO10→ch9
		esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)
		setSensAtten1(ch, attenDefault)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD(1 << ch)
		for i := 0; i < 100; i++ {
		}

		for esp.SENS.GetSAR_SLAVE_ADDR1_SAR_SARADC_MEAS_STATUS() != 0 {
		}
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(0)
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

func (a ADC) GetVoltage() (raw uint32, v float64) {
	const samples = 4
	var sum uint32
	for i := 0; i < samples; i++ {
		sum += uint32(a.Get() >> 4)
	}
	raw = sum / samples

	// Default full-scale for 11 dB is approximately 3.3 V assuming
	// Vref ≈ 1.1 V and gain ≈ 3. If eFuse provided a per-chip DIGI_REF
	// (Vref in mV) via ADCDefaultCalibration, use it to adjust the
	// full-scale range instead.
	scale := 3.3
	if digiRefMv != 0 {
		scale = 3.0 * float64(digiRefMv) / 1000.0
	}

	v = float64(raw) / 4095.0 * scale
	return raw, v
}
