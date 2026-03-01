//go:build esp32s3 || (esp32c3 && !m5stamp_c3)

// Shared regI2C-based ADC calibration helpers for ESP32-S3 and ESP32-C3.
//
// The internal I2C bus ("regI2C") and SAR ADC trim register layout are
// identical across both chips; chip-specific differences (host ID, DREF
// init value, calibration iterations) are captured in the regI2C struct
// fields, keeping each target file free of duplicated low-level code.

package machine

import (
	"device/esp"
	"runtime/volatile"
	"unsafe"
)

// regI2C wraps the internal I2C bus used for SAR ADC calibration registers.
// Fields hold chip-specific parameters that differ between ESP32-S3 and ESP32-C3.
type regI2C struct {
	// hostID is the I2C_SAR_ADC_HOSTID (1 for ESP32-S3, 0 for ESP32-C3).
	hostID uint8
	// drefInit is the DREF reference value written during calibrationInit
	// (4 for ESP32-S3, 1 for ESP32-C3).
	drefInit uint8
}

// SAR ADC I2C register layout constants shared across ESP32-S3 and ESP32-C3.
// Source: ESP-IDF soc/regi2c_saradc.h
const (
	// i2cSarADC is the I2C_SAR_ADC block address on the internal bus.
	i2cSarADC = uint8(0x69)

	// DREF (reference) bitfields for ADC1 and ADC2.
	adc1DrefAddr = uint8(0x2)
	adc1DrefMSB  = uint8(6)
	adc1DrefLSB  = uint8(4)
	adc2DrefAddr = uint8(0x5)
	adc2DrefMSB  = uint8(6)
	adc2DrefLSB  = uint8(4)

	// ENCAL_GND: routes internal ground to ADC input during self-calibration.
	adc1EncalGndAddr = uint8(0x7)
	adc1EncalGndMSB  = uint8(5)
	adc1EncalGndLSB  = uint8(5)
	adc2EncalGndAddr = uint8(0x7)
	adc2EncalGndMSB  = uint8(7)
	adc2EncalGndLSB  = uint8(7)

	// INIT_CODE (offset) high/low for ADC1 and ADC2.
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

	// adcCalOffsetRange is the binary search upper bound (12-bit full scale).
	adcCalOffsetRange = uint32(4096)

	// adcCalMaxIterations is the maximum number of calibration iterations
	// supported by calibrateBinarySearch. Must be >= max(S3=10, C3=15).
	adcCalMaxIterations = 16
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

// calibrationInit sets the DREF reference for the selected ADC unit to
// the chip-specific init value before running self-calibration.
// Corresponds to adc_ll_calibration_init() in ESP-IDF.
func (r regI2C) calibrationInit(adcN uint8) {
	if adcN == 0 {
		r.writeMask(adc1DrefAddr, adc1DrefMSB, adc1DrefLSB, r.drefInit)
	} else {
		r.writeMask(adc2DrefAddr, adc2DrefMSB, adc2DrefLSB, r.drefInit)
	}
}

// calibrationPrepare enables ENCAL_GND so that the ADC input is
// internally shorted to ground during self-calibration.
// Corresponds to the ENCAL_GND part of adc_ll_calibration_prepare().
func (r regI2C) calibrationPrepare(adcN uint8) {
	if adcN == 0 {
		r.writeMask(adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 1)
	} else {
		r.writeMask(adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 1)
	}
}

// calibrationFinish clears ENCAL_GND to reconnect the ADC input to the
// external pad after self-calibration.
// Corresponds to adc_ll_calibration_finish() in ESP-IDF.
func (r regI2C) calibrationFinish(adcN uint8) {
	if adcN == 0 {
		r.writeMask(adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 0)
	} else {
		r.writeMask(adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 0)
	}
}

// setCalibrationParam writes the INIT_CODE (offset trim) for the selected
// ADC unit via the regI2C bitfields.
// Corresponds to adc_ll_set_calibration_param() in ESP-IDF.
func (r regI2C) setCalibrationParam(adcN uint8, param uint32) {
	msb := uint8(param >> 8)
	lsb := uint8(param & 0xFF)
	if adcN == 0 {
		r.writeMask(adc1InitCodeHighAddr, adc1InitCodeHighMSB, adc1InitCodeHighLSB, msb)
		r.writeMask(adc1InitCodeLowAddr, adc1InitCodeLowMSB, adc1InitCodeLowLSB, lsb)
	} else {
		r.writeMask(adc2InitCodeHighAddr, adc2InitCodeHighMSB, adc2InitCodeHighLSB, msb)
		r.writeMask(adc2InitCodeLowAddr, adc2InitCodeLowMSB, adc2InitCodeLowLSB, lsb)
	}
}

// calibrateBinarySearch runs the ADC self-calibration binary search loop.
// It performs 'iterations' rounds of binary search to find the optimal offset
// code, drops the min/max outliers, and returns the rounded mean of the
// remaining values. This matches adc_hal_self_calibration() in ESP-IDF.
//
// The readADC callback must perform a single conversion using the target's
// oneshot path (SENS or APB_SARADC) and return the raw 12-bit result.
// During calibration, ENCAL_GND is active so the ADC reads its internal ground.
func (r regI2C) calibrateBinarySearch(adcN uint8, iterations int, readADC func() uint32) uint32 {
	if iterations > adcCalMaxIterations {
		iterations = adcCalMaxIterations
	}
	var codeList [adcCalMaxIterations]uint32
	var codeSum uint32

	for rpt := 0; rpt < iterations; rpt++ {
		codeH := adcCalOffsetRange
		codeL := uint32(0)
		chkCode := (codeH + codeL) / 2
		r.setCalibrationParam(adcN, chkCode)
		selfCal := readADC()

		for codeH-codeL > 1 {
			if selfCal == 0 {
				codeH = chkCode
			} else {
				codeL = chkCode
			}
			chkCode = (codeH + codeL) / 2
			r.setCalibrationParam(adcN, chkCode)
			selfCal = readADC()
			if codeH-codeL == 1 {
				chkCode++
				r.setCalibrationParam(adcN, chkCode)
				selfCal = readADC()
			}
		}
		codeList[rpt] = chkCode
		codeSum += chkCode
	}

	// Drop min and max outliers, then average with IDF-style rounding.
	codeMin := codeList[0]
	codeMax := codeList[0]
	for i := 0; i < iterations; i++ {
		if codeList[i] < codeMin {
			codeMin = codeList[i]
		}
		if codeList[i] > codeMax {
			codeMax = codeList[i]
		}
	}
	remaining := codeSum - codeMax - codeMin
	divisor := uint32(iterations - 2)
	finalCode := remaining / divisor
	if remaining%divisor >= 4 {
		finalCode++
	}

	return finalCode
}
