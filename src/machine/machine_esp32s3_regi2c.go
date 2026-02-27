//go:build esp32s3

// RegI2C — внутренний I2C для SAR ADC (протокол как ESP32-S2 I2C_RTC_CONFIG2, регистр 0x6000E000).
// Source: idf-source/components/soc/esp32s3/include/soc/regi2c_saradc.h

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

const (
	// I2C_SAR_ADC / I2C_SAR_ADC_HOSTID in regi2c_saradc.h
	i2cSarADC       = uint8(0x69) // I2C_SAR_ADC
	i2cSarADCHostID = uint8(1)    // I2C_SAR_ADC_HOSTID

	// ADC_SAR1_DREF_ADDR(_MSB/_LSB)
	adc1DrefAddr = uint8(0x2) // ADC_SAR1_DREF_ADDR
	adc1DrefMSB  = uint8(6)   // ADC_SAR1_DREF_ADDR_MSB
	adc1DrefLSB  = uint8(4)   // ADC_SAR1_DREF_ADDR_LSB

	// ADC_SAR2_DREF_ADDR(_MSB/_LSB)
	adc2DrefAddr = uint8(0x5) // ADC_SAR2_DREF_ADDR
	adc2DrefMSB  = uint8(6)   // ADC_SAR2_DREF_ADDR_MSB
	adc2DrefLSB  = uint8(4)   // ADC_SAR2_DREF_ADDR_LSB

	// ADC_SAR1_ENCAL_GND_ADDR(_MSB/_LSB)
	adc1EncalGndAddr = uint8(0x7) // ADC_SAR1_ENCAL_GND_ADDR
	adc1EncalGndMSB  = uint8(5)   // ADC_SAR1_ENCAL_GND_ADDR_MSB
	adc1EncalGndLSB  = uint8(5)   // ADC_SAR1_ENCAL_GND_ADDR_LSB

	// ADC_SAR2_ENCAL_GND_ADDR(_MSB/_LSB)
	adc2EncalGndAddr = uint8(0x7) // ADC_SAR2_ENCAL_GND_ADDR
	adc2EncalGndMSB  = uint8(7)   // ADC_SAR2_ENCAL_GND_ADDR_MSB
	adc2EncalGndLSB  = uint8(7)   // ADC_SAR2_ENCAL_GND_ADDR_LSB

	// ADC_SAR1_INITIAL_CODE_HIGH/LOW_ADDR(_MSB/_LSB)
	adc1InitCodeHighAddr = uint8(0x1) // ADC_SAR1_INITIAL_CODE_HIGH_ADDR
	adc1InitCodeHighMSB  = uint8(3)   // ADC_SAR1_INITIAL_CODE_HIGH_ADDR_MSB
	adc1InitCodeHighLSB  = uint8(0)   // ADC_SAR1_INITIAL_CODE_HIGH_ADDR_LSB
	adc1InitCodeLowAddr  = uint8(0x0) // ADC_SAR1_INITIAL_CODE_LOW_ADDR
	adc1InitCodeLowMSB   = uint8(7)   // ADC_SAR1_INITIAL_CODE_LOW_ADDR_MSB
	adc1InitCodeLowLSB   = uint8(0)   // ADC_SAR1_INITIAL_CODE_LOW_ADDR_LSB

	// ADC_SAR2_INITIAL_CODE_HIGH/LOW_ADDR(_MSB/_LSB)
	adc2InitCodeHighAddr = uint8(0x4) // ADC_SAR2_INITIAL_CODE_HIGH_ADDR
	adc2InitCodeHighMSB  = uint8(3)   // ADC_SAR2_INITIAL_CODE_HIGH_ADDR_MSB
	adc2InitCodeHighLSB  = uint8(0)   // ADC_SAR2_INITIAL_CODE_HIGH_ADDR_LSB
	adc2InitCodeLowAddr  = uint8(0x3) // ADC_SAR2_INITIAL_CODE_LOW_ADDR
	adc2InitCodeLowMSB   = uint8(7)   // ADC_SAR2_INITIAL_CODE_LOW_ADDR_MSB
	adc2InitCodeLowLSB   = uint8(0)   // ADC_SAR2_INITIAL_CODE_LOW_ADDR_LSB

	// Analog config registers for RegI2C block (RTC/ANA config in TRM).
	anaConfigReg  = uintptr(0x6000E044)
	i2cSarEnMask  = uint32(1 << 18)
	anaConfig2Reg = uintptr(0x6000E048)
	anaSarCfg2En  = uint32(1 << 16)

	// REGI2C master control register and helper masks.
	i2cMstCtrlHost1   = uintptr(0x6000E000)
	i2cMstBusyBit     = uint32(1 << 25)
	i2cMstWrCntlBit   = uint32(1 << 24)
	i2cMstDataMask    = uint32(0xFF << 16)
	i2cMstDataShift   = 16
	i2cMstBusyTimeout = 10000
)

type RegI2C struct{}

// waitIdle mimics the IDF regi2c busy-wait helper (see regi2c_ctrl.c).
// It polls the REGI2C master control register until the BUSY bit clears
// or a small timeout expires, to avoid writing while a previous transfer
// is still in progress.
func (r *RegI2C) waitIdle(reg *volatile.Register32) bool {
	for i := 0; i < i2cMstBusyTimeout; i++ {
		if reg.Get()&i2cMstBusyBit == 0 {
			return true
		}
	}
	return false
}

// WriteMask is a software implementation of the REGI2C_WRITE_MASK macro
// from IDF (see soc/regi2c_saradc.h). It:
//   - selects the RegI2C SAR ADC block + register address,
//   - reads the current byte,
//   - updates only the [msb:lsb] bitfield,
//   - writes the new value back via the internal I2C master.
func (r *RegI2C) WriteMask(block, hostID, regAddr, msb, lsb, data uint8) {
	if hostID != i2cSarADCHostID {
		return
	}
	reg := (*volatile.Register32)(unsafe.Pointer(i2cMstCtrlHost1))
	if !r.waitIdle(reg) {
		return
	}
	reg.Set(uint32(block) | uint32(regAddr)<<8)
	if !r.waitIdle(reg) {
		return
	}
	cur := (reg.Get() & i2cMstDataMask) >> i2cMstDataShift
	mask := uint32(1<<(msb-lsb+1)-1) << lsb
	cur &^= mask
	cur |= uint32(data&(1<<(msb-lsb+1)-1)) << lsb
	reg.Set(uint32(block) | uint32(regAddr)<<8 | i2cMstWrCntlBit | (cur<<i2cMstDataShift)&i2cMstDataMask)
	r.waitIdle(reg)
}

// ReadMask is a software implementation of REGI2C_READ_MASK from IDF.
// It selects the SAR ADC RegI2C address, reads the current byte and
// returns only the requested [msb:lsb] bitfield.
func (r *RegI2C) ReadMask(block, hostID, regAddr, msb, lsb uint8) uint8 {
	if hostID != i2cSarADCHostID {
		return 0
	}
	reg := (*volatile.Register32)(unsafe.Pointer(i2cMstCtrlHost1))
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

// SarEnable enables the analog SAR I2C domain before any RegI2C access,
// matching the prologue in adc_ll_calibration_prepare() (sets ANA_SAR_CFG2_EN).
func (r *RegI2C) SarEnable() {
	cfg := (*volatile.Register32)(unsafe.Pointer(anaConfigReg))
	cfg2 := (*volatile.Register32)(unsafe.Pointer(anaConfig2Reg))
	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	cfg.Set(cfg.Get() &^ i2cSarEnMask)
	cfg2.Set(cfg2.Get() | anaSarCfg2En)
}

// ADC1CalibrationInit corresponds to adc_ll_calibration_init() for ESP32-S3:
// it sets the DREF field to 4 for the selected ADC unit, which is the
// reference index used by Espressif's calibration flow.
func (r *RegI2C) ADC1CalibrationInit(adcN uint8) {
	if adcN == 0 {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc1DrefAddr, adc1DrefMSB, adc1DrefLSB, 4)
	} else {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc2DrefAddr, adc2DrefMSB, adc2DrefLSB, 4)
	}
}

// ADC1CalibrationPrepare corresponds to the ENCAL_GND part of
// adc_ll_calibration_prepare(): it temporarily routes the internal
// ground reference into the SAR input so that self-calibration can
// measure offset with the pin disconnected.
func (r *RegI2C) ADC1CalibrationPrepare(adcN uint8) {
	if adcN == 0 {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 1)
	} else {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 1)
	}
}

// ADC1CalibrationFinish corresponds to adc_ll_calibration_finish():
// it clears ENCAL_GND so that ADC input is again connected to the pad.
func (r *RegI2C) ADC1CalibrationFinish(adcN uint8) {
	if adcN == 0 {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 0)
	} else {
		r.WriteMask(i2cSarADC, i2cSarADCHostID, adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 0)
	}
}

// ADC1SetCalibrationParam corresponds to adc_ll_set_calibration_param():
// it writes the 9-bit initial code (offset) into the high/low INIT_CODE
// RegI2C registers for the selected ADC unit.
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
