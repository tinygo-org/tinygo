//go:build esp32c6

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

// newRegI2C returns the regI2C configured for ESP32-C6: hostID=0, drefInit=1.
func newRegI2C() regI2C { return regI2C{hostID: 0, drefInit: 1} }

const (
	// ADC attenuation values for ESP32-C6 APB_SARADC.
	// 0 dB  : ~0 .. 1.1 V
	// 11 dB : ~0 .. 3.3 V (matches typical VDD)
	atten0dB  = 0
	atten11dB = 3
)

// sarEnable enables the analog SAR I2C domain before any regI2C access.
// On ESP32-C6, RTC_CNTL does not exist; the SAR analog power is managed
// through the analog config registers directly.
func (r regI2C) sarEnable() {
	cfg := (*volatile.Register32)(unsafe.Pointer(anaConfigReg))
	cfg2 := (*volatile.Register32)(unsafe.Pointer(anaConfig2Reg))
	cfg.Set(cfg.Get() &^ i2cSarEnMask)
	cfg2.Set(cfg2.Get() | anaSarCfg2En)
}

func InitADC() {
	// Enable APB_SARADC clock via PCR.
	esp.PCR.SARADC_CONF.SetBits(1 << 0) // SARADC_REG_CLK_EN
	// Reset sequence
	esp.PCR.SARADC_CONF.SetBits(1 << 1)   // SARADC_RST_EN
	esp.PCR.SARADC_CONF.ClearBits(1 << 1) // clear reset

	esp.APB_SARADC.SetCTRL_SARADC_XPD_SAR_FORCE(1)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_XPD_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_RSTB_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_STANDBY_WAIT(100)
	esp.APB_SARADC.SetCLKM_CONF_CLK_SEL(2)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_NUM(1)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_B(0)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_A(0)
	esp.APB_SARADC.SetCLKM_CONF_CLK_EN(1)

	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_GATED(1)
	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_DIV(1)
}

// ESP32-C6: ADC1 = GPIO0–GPIO6 (ch 0–6). ADC2 is not available on C6.
func (a ADC) Configure(config ADCConfig) error {
	if a.Pin > 6 {
		return errors.New("invalid ADC pin for ESP32-C6")
	}
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	return nil
}

func (a ADC) Get() uint16 {
	if a.Pin > 6 {
		return 0
	}
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)

	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(uint32(a.Pin))
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA()
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)

	return uint16(raw&0xfff) << 4
}
