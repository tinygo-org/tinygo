//go:build esp32c3 && !m5stamp_c3

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// RegI2C on ESP32‑C3 exposes the internal analog I2C bus that controls
// SAR ADC trim registers. Constants below mirror the layout from
// ESP‑IDF's soc/regi2c_saradc.h and TRM (I2C_RTC_CONFIG2 block).
const (
	// i2cSarADC/i2cSarADCHostID select the SAR ADC block on the internal bus.
	i2cSarADC       = uint8(0x69)
	i2cSarADCHostID = uint8(0)

	// adc*_Dref* define the DREF (reference) bitfields for ADC1/ADC2.
	adc1DrefAddr = uint8(0x2)
	adc1DrefMSB  = uint8(6)
	adc1DrefLSB  = uint8(4)

	adc2DrefAddr = uint8(0x5)
	adc2DrefMSB  = uint8(6)
	adc2DrefLSB  = uint8(4)

	// adc*_EncalGnd* control ENCAL_GND: route internal ground to ADC input
	// during self‑calibration so that the pin is effectively disconnected.
	adc1EncalGndAddr = uint8(0x7)
	adc1EncalGndMSB  = uint8(5)
	adc1EncalGndLSB  = uint8(5)

	adc2EncalGndAddr = uint8(0x7)
	adc2EncalGndMSB  = uint8(7)
	adc2EncalGndLSB  = uint8(7)

	// adc*_InitCode* hold the INIT_CODE (offset) that hardware uses to
	// compensate ADC1/ADC2 offset error.
	adc1InitCodeHighAddr = uint8(0x1)
	adc1InitCodeHighMSB  = uint8(3)
	adc1InitCodeHighLSB  = uint8(0)
	adc1InitCodeLowAddr  = uint8(0x0)
	adc1InitCodeLowMSB   = uint8(7)
	adc1InitCodeLowLSB   = uint8(0)

	adc2InitCodeHighAddr = uint8(0x4)
	adc2InitCodeHighMSB  = uint8(3)
	adc2InitCodeHighLSB  = uint8(0)
	adc2InitCodeLowAddr  = uint8(0x3)
	adc2InitCodeLowMSB   = uint8(7)
	adc2InitCodeLowLSB   = uint8(0)

	// ANA_CONFIG/ANA_CONFIG2: enable analog SAR I2C domain before RegI2C access.
	anaConfigReg  = uintptr(0x6000E044)
	i2cSarEnMask  = uint32(1 << 18)
	anaConfig2Reg = uintptr(0x6000E048)
	anaSarCfg2En  = uint32(1 << 16)

	// I2C_RTC_CONFIG2 master control register used by RegI2C operations.
	i2cMstCtrlHost  = uintptr(0x6000E000)
	i2cMstBusyBit   = uint32(1 << 25)
	i2cMstWrCntl    = uint32(1 << 24)
	i2cMstDataMask  = uint32(0xFF << 16)
	i2cMstDataShift = 16
	i2cMstTimeout   = 10000
)

type RegI2C struct{}

var DefaultRegI2C RegI2C

// waitIdle polls the REGI2C master BUSY bit until it clears or the
// simple software timeout expires. This matches the busy‑wait helper
// used in ESP‑IDF's regi2c_ctrl.c.
func (r *RegI2C) waitIdle(reg *volatile.Register32) bool {
	for i := 0; i < i2cMstTimeout; i++ {
		if reg.Get()&i2cMstBusyBit == 0 {
			return true
		}
	}
	return false
}

// WriteMask is a software implementation of REGI2C_WRITE_MASK macro:
//  1. select block + regAddr,
//  2. read current byte,
//  3. update only [msb:lsb] bitfield,
//  4. write it back via internal I2C master.
func (r *RegI2C) WriteMask(block, hostID, regAddr, msb, lsb, data uint8) {
	if hostID != i2cSarADCHostID {
		return
	}
	reg := (*volatile.Register32)(unsafe.Pointer(i2cMstCtrlHost))
	if !r.waitIdle(reg) {
		return
	}
	reg.Set(uint32(block) | uint32(regAddr)<<8)
	if !r.waitIdle(reg) {
		return
	}
	cur := (reg.Get() & i2cMstDataMask) >> i2cMstDataShift
	mask := uint32(1<<(msb-lsb+1) - 1<<lsb)
	cur &^= mask
	cur |= uint32(data&(1<<(msb-lsb+1)-1)) << lsb
	reg.Set(uint32(block) | uint32(regAddr)<<8 | i2cMstWrCntl | (cur<<i2cMstDataShift)&i2cMstDataMask)
	r.waitIdle(reg)
}

// ReadMask is a software implementation of REGI2C_READ_MASK macro:
// it selects the SAR ADC register and returns only the [msb:lsb] field.
func (r *RegI2C) ReadMask(block, hostID, regAddr, msb, lsb uint8) uint8 {
	if hostID != i2cSarADCHostID {
		return 0
	}
	reg := (*volatile.Register32)(unsafe.Pointer(i2cMstCtrlHost))
	if !r.waitIdle(reg) {
		return 0
	}
	reg.Set(uint32(block) | uint32(regAddr)<<8)
	if !r.waitIdle(reg) {
		return 0
	}
	data := (reg.Get() & i2cMstDataMask) >> i2cMstDataShift
	return uint8((data >> lsb) & (1<<(msb-lsb+1) - 1))
}

// SarEnable enables the SAR analog I2C domain before any RegI2C access.
func (r *RegI2C) SarEnable() {
	cfg := (*volatile.Register32)(unsafe.Pointer(anaConfigReg))
	cfg2 := (*volatile.Register32)(unsafe.Pointer(anaConfig2Reg))
	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	cfg.Set(cfg.Get() &^ i2cSarEnMask)
	cfg2.Set(cfg2.Get() | anaSarCfg2En)
}

// ADC1CalibrationInit sets DREF for the selected ADC unit
// before running the self‑calibration procedure.
func (r *RegI2C) ADC1CalibrationInit(adcN uint8) {
	if adcN == 0 {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc1DrefAddr, adc1DrefMSB, adc1DrefLSB, 1)
	} else {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc2DrefAddr, adc2DrefMSB, adc2DrefLSB, 1)
	}
}

// ADC1CalibrationPrepare enables ENCAL_GND so that the ADC input
// is internally shorted to ground during self‑calibration.
func (r *RegI2C) ADC1CalibrationPrepare(adcN uint8) {
	if adcN == 0 {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 1)
	} else {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 1)
	}
}

// ADC1CalibrationFinish clears ENCAL_GND and reconnects the ADC
// input back to the external pad after self‑calibration.
func (r *RegI2C) ADC1CalibrationFinish(adcN uint8) {
	if adcN == 0 {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 0)
	} else {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 0)
	}
}

// ADC1SetCalibrationParam writes the INIT_CODE (offset trim) for
// the selected ADC unit using the RegI2C bitfields.
func (r *RegI2C) ADC1SetCalibrationParam(adcN uint8, param uint32) {
	msb := uint8(param >> 8)
	lsb := uint8(param & 0xFF)
	if adcN == 0 {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc1InitCodeHighAddr, adc1InitCodeHighMSB, adc1InitCodeHighLSB, msb)
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc1InitCodeLowAddr, adc1InitCodeLowMSB, adc1InitCodeLowLSB, lsb)
	} else {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc2InitCodeHighAddr, adc2InitCodeHighMSB, adc2InitCodeHighLSB, msb)
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc2InitCodeLowAddr, adc2InitCodeLowMSB, adc2InitCodeLowLSB, lsb)
	}
}
