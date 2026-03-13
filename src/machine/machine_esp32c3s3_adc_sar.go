//go:build (esp32c3 && !m5stamp_c3) || esp32s3

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// sarEnable enables the analog SAR I2C domain before any regI2C access,
// matching the prologue in adc_ll_calibration_prepare().
func (r regI2C) sarEnable() {
	cfg := (*volatile.Register32)(unsafe.Pointer(anaConfigReg))
	cfg2 := (*volatile.Register32)(unsafe.Pointer(anaConfig2Reg))
	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	cfg.Set(cfg.Get() &^ i2cSarEnMask)
	cfg2.Set(cfg2.Get() | anaSarCfg2En)
}
