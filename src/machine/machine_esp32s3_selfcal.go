//go:build esp32s3

// ADC hardware self-calibration for ESP32-S3.
//
// Mapping to ESP-IDF (adc_hal_common.c, hal/esp32s3/adc_ll.h):
//   - adc_hal_self_calibration()      → ADCSelfCalibrate()
//   - adc_ll_calibration_init()       → RegI2C.ADC1CalibrationInit (DREF=4);
//                                      in IDF it is not called from self_cal, we call it explicitly.
//   - adc_ll_calibration_prepare()    → SarEnable + ADC1CalibrationPrepare (ENCAL_GND=1)
//   - adc_ll_calibration_finish()     → ADC1CalibrationFinish (ENCAL_GND=0)
//   - adc_ll_set_calibration_param()  → ADC1SetCalibrationParam()
//   - read_cal_channel()              → ADCDefaultCalibration.readADC1():
//                                      wait for meas_status==0, start 0→1, wait done, read data
//                                      (similar to adc_oneshot_ll_start + get_raw_result).
//   - Loop: 10 iterations, code 0..4096, binary search on self_cal==0; drop min/max;
//           rounding (remainder%8 < 4 without +1, otherwise +1) — same as in adc_hal_common.c.
//   - raw_check_valid: for ADC1 in IDF always true — we do not check it.
//
// Differences:
//   - RegI2C: not ROM helper but direct access to 0x6000E000 (protocol like I2C_RTC_CONFIG2).
//   - cal_setup: same SENS/atten/controller fields, but through our registers.
//   - Result is stored only in hardware for the current session (not in eFuse).
//   - eFuse V1: init_code and digi_ref are taken from eFuse — same idea as Arduino/IDF.

package machine

import (
	"device/esp"
)

const (
	adcCalTimes     = 10
	adcCalOffsetMax = uint32(4096)
	adcCalRtcMagic  = uint32(0xADC1C401)
	adcCalInitMin   = uint32(1500)
	adcCalInitMax   = uint32(4096)
)

// ADCDefaultCalibration encapsulates the self-calibration flow for ADC1
// and remembers per-chip calibration data (such as DIGI_REF) when it is
// available from eFuse.
type ADCDefaultCalibration struct {
	digiRefMv uint32
}

func (c *ADCDefaultCalibration) SelfCalibrate() {
	reg := RegI2C{}
	fuse := Fuse{}

	if vref, ok := fuse.ADC1DigiRefAtten3(); ok {
		c.digiRefMv = vref
	}

	if saved, ok := c.restoreFromRTC(); ok {
		reg.SarEnable()
		reg.ADC1CalibrationInit(0)
		reg.ADC1SetCalibrationParam(0, saved)
		return
	}

	initCode, useEfuse := fuse.ADC1InitCodeAtten3()
	c.adc1CalibrationSetup(reg)

	if useEfuse {
		c.saveToRTC(initCode)
		c.adc1CalibrateHigh(reg, initCode)
		return
	}

	finalCode := c.adc1CalibrateLow(reg)
	c.saveToRTC(finalCode)
	c.adc1CalibrateHigh(reg, finalCode)
}

func (c *ADCDefaultCalibration) GetDigiRef() uint32 {
	return c.digiRefMv
}

func (c *ADCDefaultCalibration) adc1CalibrationSetup(reg RegI2C) {
	reg.SarEnable()

	esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(0)
	esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_FORCE(0)
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD(0)
	setSensAtten1(0, attenDefault)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)

	reg.ADC1CalibrationInit(0)
	reg.ADC1CalibrationPrepare(0)
}

func (c *ADCDefaultCalibration) adc1CalibrateLow(reg RegI2C) uint32 {
	var codeList [adcCalTimes]uint32
	var codeSum uint32

	for rpt := 0; rpt < adcCalTimes; rpt++ {
		codeH := adcCalOffsetMax
		codeL := uint32(0)
		chkCode := (codeH + codeL) / 2
		reg.ADC1SetCalibrationParam(0, chkCode)
		selfCal := c.readADC1()

		for codeH-codeL > 1 {
			if selfCal == 0 {
				codeH = chkCode
			} else {
				codeL = chkCode
			}
			chkCode = (codeH + codeL) / 2
			reg.ADC1SetCalibrationParam(0, chkCode)
			selfCal = c.readADC1()
			if codeH-codeL == 1 {
				chkCode++
				reg.ADC1SetCalibrationParam(0, chkCode)
				selfCal = c.readADC1()
			}
		}
		codeList[rpt] = chkCode
		codeSum += chkCode
	}

	codeL := codeList[0]
	codeH := codeList[0]
	for i := 0; i < adcCalTimes; i++ {
		if codeList[i] < codeL {
			codeL = codeList[i]
		}
		if codeList[i] > codeH {
			codeH = codeList[i]
		}
	}
	excluded := codeH + codeL
	remaining := codeSum - excluded
	finalCode := remaining / (adcCalTimes - 2)
	if remaining%(adcCalTimes-2) >= 4 {
		finalCode++
	}

	return finalCode
}

func (c *ADCDefaultCalibration) adc1CalibrateHigh(reg RegI2C, code uint32) {
	reg.ADC1SetCalibrationParam(0, code)
	reg.ADC1CalibrationFinish(0)
	c.adc1StartWithPadForce()
}

func (c *ADCDefaultCalibration) adc1StartWithPadForce() {
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
}

// readADC1 performs one ADC1 conversion via RTC path (used during calibration).
// Internal GND is connected via ENCAL_GND, so the pin input is disconnected.
// Matches IDF: wait conversion idle (meas_status==0), then start 0→1, wait done, read data.
func (c *ADCDefaultCalibration) readADC1() uint32 {
	for esp.SENS.GetSAR_SLAVE_ADDR1_SAR_SARADC_MEAS_STATUS() != 0 {
	}
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(0)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(1)
	for esp.SENS.GetSAR_MEAS1_CTRL2_MEAS1_DONE_SAR() == 0 {
	}
	return uint32(esp.SENS.GetSAR_MEAS1_CTRL2_MEAS1_DATA_SAR() & 0xfff)
}

func (c *ADCDefaultCalibration) restoreFromRTC() (uint32, bool) {
	if esp.RTC_CNTL.GetSTORE0() != adcCalRtcMagic {
		return 0, false
	}
	code := esp.RTC_CNTL.GetSTORE1()
	if code < adcCalInitMin || code > adcCalInitMax {
		return 0, false
	}
	return code, true
}

func (c *ADCDefaultCalibration) saveToRTC(code uint32) {
	esp.RTC_CNTL.SetSTORE0(adcCalRtcMagic)
	esp.RTC_CNTL.SetSTORE1(code)
}
