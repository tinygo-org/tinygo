//go:build esp32c3 && !m5stamp_c3

package machine

import "device/esp"

const (
	// adcCalTimesC3 is the number of calibration iterations used to
	// collect INIT_CODE candidates before discarding min/max and averaging.
	adcCalTimesC3 = 10

	// adcCalOffsetRangeC3 is the search range for the offset code.
	// It matches the 12‑bit SAR ADC range used in the IDF self‑cal algorithm.
	adcCalOffsetRangeC3 = uint32(4096)
)

// adcC3Calibration implements the ESP32‑C3 ADC self‑calibration flow
// similar to adc_hal_self_calibration() in ESP‑IDF: internal GND is
// routed to ADC1, and a binary search is performed over INIT_CODE.
type adcC3Calibration struct{}

// calSetup configures APB_SARADC for oneshot sampling on ADC1 channel 0
// with fixed attenuation. This is used only during self‑calibration.
func (c *adcC3Calibration) calSetup() {
	// Match the default runtime attenuation so the calibration code is
	// applied to the same operating range (11 dB → ~0..3.3 V).
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
}

// readADC1 performs a single ADC1 conversion using the APB_SARADC
// oneshot path and returns the raw 12‑bit result (0..4095).
func (c *adcC3Calibration) readADC1() uint32 {
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	for i := 0; i < 10; i++ {
	}
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA() & 0xfff
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	return uint32(raw)
}

// SelfCalibrate runs the hardware offset calibration for ADC1:
//  1. enables the SAR analog I2C domain and ENCAL_GND via RegI2C,
//  2. performs a binary search over INIT_CODE in [0..adcCalOffsetRangeC3),
//  3. repeats the search adcCalTimesC3 times, discards min/max,
//  4. writes the averaged INIT_CODE back to the SAR ADC trim registers.
func (c *adcC3Calibration) SelfCalibrate() {
	reg := RegI2C{}

	// Optional: log eFuse calibration data (INIT_CODE / DIGI_REF) if present.
	var fuse Fuse
	if initCode, ok := fuse.ADC1InitCodeAtten3(); ok {
		println("ESP32-C3 eFuse ADC1 INIT_CODE atten3:", initCode)
	} else {
		println("ESP32-C3 eFuse ADC1 INIT_CODE atten3: not present")
	}
	if digi, ok := fuse.ADC1DigiRefAtten3(); ok {
		println("ESP32-C3 eFuse ADC1 DIGI_REF atten3 digi:", digi)
	} else {
		println("ESP32-C3 eFuse ADC1 DIGI_REF atten3: not present")
	}

	c.calSetup()

	reg.SarEnable()
	reg.ADC1CalibrationInit(0)
	reg.ADC1CalibrationPrepare(0)

	var codeList [adcCalTimesC3]uint32
	var codeSum uint32

	for rpt := 0; rpt < adcCalTimesC3; rpt++ {
		codeH := adcCalOffsetRangeC3
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
	for i := 0; i < adcCalTimesC3; i++ {
		if codeList[i] < codeL {
			codeL = codeList[i]
		}
		if codeList[i] > codeH {
			codeH = codeList[i]
		}
	}
	excluded := codeH + codeL
	remaining := codeSum - excluded
	finalCode := remaining / (adcCalTimesC3 - 2)
	if remaining%(adcCalTimesC3-2) >= 4 {
		finalCode++
	}

	reg.ADC1SetCalibrationParam(0, finalCode)
	reg.ADC1CalibrationFinish(0)
}
