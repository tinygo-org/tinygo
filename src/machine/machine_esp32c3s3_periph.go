//go:build esp32c3 || esp32s3

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// enableI2C0PeriphClock enables the I2C0 peripheral clock via SYSTEM.
func enableI2C0PeriphClock() {
	esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT0_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_I2C_EXT0_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_I2C_EXT0_RST(0)
}

// enableLEDCPeriphClock enables the LEDC peripheral clock via SYSTEM.
func enableLEDCPeriphClock() {
	esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_LEDC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_LEDC_RST(0)
}

// sarEnable enables the analog SAR I2C domain before any regI2C access,
// matching the prologue in adc_ll_calibration_prepare().
func (r regI2C) sarEnable() {
	cfg := (*volatile.Register32)(unsafe.Pointer(anaConfigReg))
	cfg2 := (*volatile.Register32)(unsafe.Pointer(anaConfig2Reg))
	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	cfg.Set(cfg.Get() &^ i2cSarEnMask)
	cfg2.Set(cfg2.Get() | anaSarCfg2En)
}
