//go:build esp32s3 || esp32c3

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// REGI2C master at 0x6000E000 used by ESP32-S3 and ESP32-C3.
const (
	// ANA_CONFIG / ANA_CONFIG2: enable analog SAR I2C domain.
	anaConfigReg  = uintptr(0x6000E044)
	i2cSarEnMask  = uint32(1 << 18)
	anaConfig2Reg = uintptr(0x6000E048)
	anaSarCfg2En  = uint32(1 << 16)

	// REGI2C master control register and helper masks.
	i2cMstCtrlReg     = uintptr(0x6000E000)
	i2cMstBusyBit     = uint32(1 << 25)
	i2cMstWrCntlBit   = uint32(1 << 24)
	i2cMstDataMask    = uint32(0xFF << 16)
	i2cMstDataShift   = 16
	i2cMstBusyTimeout = 10000
)

// waitIdle polls the REGI2C master BUSY bit until it clears or a
// timeout expires, matching the busy-wait helper in ESP-IDF's regi2c_ctrl.c.
func (r regI2C) waitIdle(reg *volatile.Register32) bool {
	for i := 0; i < i2cMstBusyTimeout; i++ {
		if reg.Get()&i2cMstBusyBit == 0 {
			return true
		}
	}
	return false
}

// writeMask is a software implementation of the IDF REGI2C_WRITE_MASK macro.
// It reads the current byte at regAddr on the SAR ADC I2C block, updates
// only the [msb:lsb] bitfield, and writes it back via the internal I2C master.
func (r regI2C) writeMask(regAddr, msb, lsb, data uint8) {
	reg := (*volatile.Register32)(unsafe.Pointer(i2cMstCtrlReg))
	if !r.waitIdle(reg) {
		return
	}
	reg.Set(uint32(i2cSarADC) | uint32(regAddr)<<8)
	if !r.waitIdle(reg) {
		return
	}
	cur := (reg.Get() & i2cMstDataMask) >> i2cMstDataShift
	mask := uint32(1<<(msb-lsb+1)-1) << lsb
	cur &^= mask
	cur |= uint32(data&(1<<(msb-lsb+1)-1)) << lsb
	reg.Set(uint32(i2cSarADC) | uint32(regAddr)<<8 | i2cMstWrCntlBit | (cur<<i2cMstDataShift)&i2cMstDataMask)
	r.waitIdle(reg)
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
