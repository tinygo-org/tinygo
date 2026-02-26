//go:build esp32s3

// ADC hardware self-calibration for ESP32-S3.
//
// Соответствие IDF (adc_hal_common.c, hal/esp32s3/adc_ll.h):
//   - adc_hal_self_calibration()      → ADCSelfCalibrate()
//   - adc_ll_calibration_init()       → DefaultRegI2C.ADC1CalibrationInit (DREF=4); в IDF в self_cal не вызывается, мы вызываем явно.
//   - adc_ll_calibration_prepare()    → SarEnable + ADC1CalibrationPrepare (ENCAL_GND=1)
//   - adc_ll_calibration_finish()     → ADC1CalibrationFinish (ENCAL_GND=0)
//   - adc_ll_set_calibration_param()  → ADC1SetCalibrationParam()
//   - read_cal_channel()              → ADCDefaultCalibration.readADC1(): ожидание meas_status==0, start 0→1, ожидание done, чтение data (как adc_oneshot_ll_start + get_raw_result).
//   - Цикл: 10 итераций, код 0..4096, бинарный поиск по self_cal==0; отбрасывание min/max; округление (remainder%8 < 4 без +1, иначе +1) — как в adc_hal_common.c.
//   - raw_check_valid: для ADC1 в IDF всегда true — не проверяем.
//
// Отличия:
//   - RegI2C: не ROM, а прямая запись в 0x6000E000 (протокол как I2C_RTC_CONFIG2).
//   - cal_setup: те же SENS/atten/controller через наши регистры.
//   - Результат только в железе на сессию (не в eFuse).
//   - eFuse V1: init_code и digi_ref из eFuse — как Arduino/IDF.

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

type ADCDefaultCalibration struct{}

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

func (c *ADCDefaultCalibration) SelfCalibrate() {
	if saved, ok := c.restoreFromRTC(); ok {
		DefaultRegI2C.SarEnable()
		DefaultRegI2C.ADC1CalibrationInit(0)
		DefaultRegI2C.ADC1SetCalibrationParam(0, saved)
		return
	}

	initCode, useEfuse := DefaultFuse.ADC1InitCodeAtten3()

	DefaultRegI2C.SarEnable()

	esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(0)
	esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_FORCE(0)
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD(0)
	setSensAtten1(0, attenDefault)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)

	DefaultRegI2C.ADC1CalibrationInit(0)
	DefaultRegI2C.ADC1CalibrationPrepare(0)

	if useEfuse {
		DefaultRegI2C.ADC1SetCalibrationParam(0, initCode)
		DefaultRegI2C.ADC1CalibrationFinish(0)
		c.saveToRTC(initCode)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
		return
	}

	var codeList [adcCalTimes]uint32
	var codeSum uint32

	for rpt := 0; rpt < adcCalTimes; rpt++ {
		codeH := adcCalOffsetMax
		codeL := uint32(0)
		chkCode := (codeH + codeL) / 2
		DefaultRegI2C.ADC1SetCalibrationParam(0, chkCode)
		selfCal := c.readADC1()

		for codeH-codeL > 1 {
			if selfCal == 0 {
				codeH = chkCode
			} else {
				codeL = chkCode
			}
			chkCode = (codeH + codeL) / 2
			DefaultRegI2C.ADC1SetCalibrationParam(0, chkCode)
			selfCal = c.readADC1()
			if codeH-codeL == 1 {
				chkCode++
				DefaultRegI2C.ADC1SetCalibrationParam(0, chkCode)
				selfCal = c.readADC1()
			}
		}
		codeList[rpt] = chkCode
		codeSum += chkCode
	}

	// Exclude min and max, average remaining 8 values
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
	var finalCode uint32
	finalCode = remaining / (adcCalTimes - 2)
	if remaining%(adcCalTimes-2) >= 4 {
		finalCode++
	}

	DefaultRegI2C.ADC1SetCalibrationParam(0, finalCode)
	DefaultRegI2C.ADC1CalibrationFinish(0)
	c.saveToRTC(finalCode)

	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
}
