//go:build esp32c3 && !m5stamp_c3

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

const (
	// ADC attenuation values for ESP32-C3 APB_SARADC.
	// 0 dB  : ~0 .. 1.1 V
	// 11 dB : ~0 .. 3.3 V (matches typical VDD)
	atten0dB  = 0
	atten11dB = 3
)

func InitADC() {
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_APB_SARADC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(0)

	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	esp.RTC_CNTL.SetSENSOR_CTRL_FORCE_XPD_SAR(1)
	esp.APB_SARADC.SetCTRL_SARADC_XPD_SAR_FORCE(1)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_XPD_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_RSTB_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_STANDBY_WAIT(100)
	esp.APB_SARADC.SetCLKM_CONF_CLK_SEL(2)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_NUM(1)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_B(0)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_A(0)
	esp.APB_SARADC.SetCLKM_CONF_CLK_EN(1)

	var c adcSelfCalibration
	c.calibrate()
}

// ESP32-C3: ADC1 = GPIO0–GPIO4 (ch 0–4), ADC2 = GPIO5 (ch 0). ADC2 shares with Wi‑Fi;
// readings may be noisy when Wi‑Fi is active.
func (a ADC) Configure(config ADCConfig) error {
	if a.Pin > 5 {
		return errors.New("invalid ADC pin for ESP32-C3")
	}
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	return nil
}

func (a ADC) Get() uint16 {
	if a.Pin > 5 {
		return 0
	}
	adc1 := a.Pin <= 4
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetINT_CLR_APB_SARADC2_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	var raw uint32
	if adc1 {
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(uint32(a.Pin))
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
		for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
		}
		raw = esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA()
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)
	} else {
		// ADC2: GPIO5 = channel 0. Grant arbiter to ADC2 first, then set channel and start.
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(1)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(1)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(8) // (1<<3)|0 for ADC2 channel 0
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC2_ONETIME_SAMPLE(1)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
		for esp.APB_SARADC.GetINT_RAW_APB_SARADC2_DONE_INT_RAW() == 0 {
		}
		raw = esp.APB_SARADC.GetSAR2DATA_STATUS_APB_SARADC2_DATA()
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
		esp.APB_SARADC.SetONETIME_SAMPLE_SARADC2_ONETIME_SAMPLE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
	}
	return uint16(raw&0xfff) << 4
}

// adcSelfCalibration
const (
	adcCalTimesC3       = 15
	adcCalOffsetRangeC3 = uint32(4096)
	adcCalRtcMagicC3    = uint32(0xADC1C401)
	adcCalInitMinC3     = uint32(1000)
	adcCalInitMaxC3     = uint32(4096)
	adcGndOffsetCompC3  = uint32(0)
)

type adcSelfCalibration struct {
	digiRefMv uint32
}

// calibrate sets ADC1/ADC2 init code from RTC or runs self-calibration (GND).
// eFuse is not used: on ESP32-C3 the ADC calibration fields in BLK2 are often unprogrammed.
func (c *adcSelfCalibration) calibrate() {
	reg := regI2C{}
	reg.sarEnable()

	var adc1Code uint32
	if saved, ok := c.restoreFromRTC(); ok {
		adc1Code = saved
	} else {
		c.calSetupADC1()
		reg.adc1CalibrationInit(0)
		reg.adc1CalibrationPrepare(0)
		adc1Code = c.calibrateUnit(reg, 0, c.readADC1)
		c.saveToRTC(adc1Code)
		reg.adc1CalibrationFinish(0)
	}

	c.applyADC1Code(reg, adc1Code)
	c.applyADC2Code(reg, adc1Code)
}

// calSetupADC1 configures APB_SARADC for oneshot sampling on ADC1 channel 0
// with fixed attenuation. This is used only during self‑calibration.
func (c *adcSelfCalibration) calSetupADC1() {
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
}

// calSetupADC2 configures APB_SARADC for oneshot sampling on ADC2 (GPIO5, ch 0).
// On C3, onetime_channel = (unit<<3)|channel → ADC2 ch0 = 8.
func (c *adcSelfCalibration) calSetupADC2() {
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(8) // (1<<3)|0 for ADC2
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(1)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC2_ONETIME_SAMPLE(1)
}

// readADC1 performs a single ADC1 conversion using the APB_SARADC
// oneshot path and returns the raw 12‑bit result (0..4095).
func (c *adcSelfCalibration) readADC1() uint32 {
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA() & 0xfff
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	return uint32(raw)
}

// readADC2 performs a single ADC2 conversion and returns the raw 12‑bit result (0..4095).
func (c *adcSelfCalibration) readADC2() uint32 {
	esp.APB_SARADC.SetINT_CLR_APB_SARADC2_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC2_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR2DATA_STATUS_APB_SARADC2_DATA() & 0xfff
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
	esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
	return uint32(raw)
}

func (c *adcSelfCalibration) restoreFromRTC() (uint32, bool) {
	if esp.RTC_CNTL.GetSTORE0() != adcCalRtcMagicC3 {
		return 0, false
	}
	code := esp.RTC_CNTL.GetSTORE1()
	if code < adcCalInitMinC3 || code > adcCalInitMaxC3 {
		return 0, false
	}
	return code, true
}

func (c *adcSelfCalibration) saveToRTC(code uint32) {
	if code < adcCalInitMinC3 || code > adcCalInitMaxC3 {
		return
	}
	esp.RTC_CNTL.SetSTORE0(adcCalRtcMagicC3)
	esp.RTC_CNTL.SetSTORE1(code)
}

// applyADC1Code sets ADC1 init code and finishes calibration.
func (c *adcSelfCalibration) applyADC1Code(reg regI2C, code uint32) {
	c.calSetupADC1()
	reg.adc1CalibrationInit(0)
	reg.adc1CalibrationPrepare(0)
	reg.adc1SetCalibrationParam(0, code)
	reg.adc1CalibrationFinish(0)
}

// applyADC2Code sets ADC2 init code and finishes calibration. On C3 eFuse V1
// there is no separate ADC2 calibration; IDF uses ADC1 init code for both units.
func (c *adcSelfCalibration) applyADC2Code(reg regI2C, code uint32) {
	reg.adc1CalibrationInit(1)
	reg.adc1CalibrationPrepare(1)
	reg.adc1SetCalibrationParam(1, code)
	reg.adc1CalibrationFinish(1)
}

func (c *adcSelfCalibration) calibrateUnit(reg regI2C, adcN uint8, readADC func() uint32) uint32 {
	var codeList [adcCalTimesC3]uint32
	var codeSum uint32

	for rpt := 0; rpt < adcCalTimesC3; rpt++ {
		codeH := adcCalOffsetRangeC3
		codeL := uint32(0)
		chkCode := (codeH + codeL) / 2
		reg.adc1SetCalibrationParam(adcN, chkCode)
		selfCal := readADC()

		for codeH-codeL > 1 {
			if selfCal == 0 {
				codeH = chkCode
			} else {
				codeL = chkCode
			}
			chkCode = (codeH + codeL) / 2
			reg.adc1SetCalibrationParam(adcN, chkCode)
			selfCal = readADC()
			if codeH-codeL == 1 {
				chkCode++
				reg.adc1SetCalibrationParam(adcN, chkCode)
				selfCal = readADC()
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
	if finalCode < adcCalInitMinC3 {
		finalCode = adcCalInitMinC3
	}
	if finalCode > adcCalInitMaxC3 {
		finalCode = adcCalInitMaxC3
	}

	reg.adc1SetCalibrationParam(adcN, finalCode)
	return finalCode
}

// regi2c

// regI2C on ESP32‑C3 exposes the internal analog I2C bus that controls
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

	// ANA_CONFIG/ANA_CONFIG2: enable analog SAR I2C domain before regI2C access.
	anaConfigReg  = uintptr(0x6000E044)
	i2cSarEnMask  = uint32(1 << 18)
	anaConfig2Reg = uintptr(0x6000E048)
	anaSarCfg2En  = uint32(1 << 16)

	// I2C_RTC_CONFIG2 master control register used by regI2C operations.
	i2cMstCtrlHost  = uintptr(0x6000E000)
	i2cMstBusyBit   = uint32(1 << 25)
	i2cMstWrCntl    = uint32(1 << 24)
	i2cMstDataMask  = uint32(0xFF << 16)
	i2cMstDataShift = 16
	i2cMstTimeout   = 10000
)

type regI2C struct{}

// sarEnable enables the SAR analog I2C domain before any regI2C access.
func (r *regI2C) sarEnable() {
	cfg := (*volatile.Register32)(unsafe.Pointer(anaConfigReg))
	cfg2 := (*volatile.Register32)(unsafe.Pointer(anaConfig2Reg))
	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	cfg.Set(cfg.Get() &^ i2cSarEnMask)
	cfg2.Set(cfg2.Get() | anaSarCfg2En)
}

// adc1CalibrationInit sets DREF for the selected ADC unit
// before running the self‑calibration procedure.
func (r *regI2C) adc1CalibrationInit(adcN uint8) {
	if adcN == 0 {
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc1DrefAddr, adc1DrefMSB, adc1DrefLSB, 1)
	} else {
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc2DrefAddr, adc2DrefMSB, adc2DrefLSB, 1)
	}
}

// adc1CalibrationPrepare enables ENCAL_GND so that the ADC input
// is internally shorted to ground during self‑calibration.
func (r *regI2C) adc1CalibrationPrepare(adcN uint8) {
	if adcN == 0 {
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 1)
	} else {
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 1)
	}
}

// adc1CalibrationFinish clears ENCAL_GND and reconnects the ADC
// input back to the external pad after self‑calibration.
func (r *regI2C) adc1CalibrationFinish(adcN uint8) {
	if adcN == 0 {
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 0)
	} else {
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 0)
	}
}

// adc1SetCalibrationParam writes the INIT_CODE (offset trim) for
// the selected ADC unit using the regI2C bitfields.
func (r *regI2C) adc1SetCalibrationParam(adcN uint8, param uint32) {
	msb := uint8(param >> 8)
	lsb := uint8(param & 0xFF)
	if adcN == 0 {
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc1InitCodeHighAddr, adc1InitCodeHighMSB, adc1InitCodeHighLSB, msb)
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc1InitCodeLowAddr, adc1InitCodeLowMSB, adc1InitCodeLowLSB, lsb)
	} else {
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc2InitCodeHighAddr, adc2InitCodeHighMSB, adc2InitCodeHighLSB, msb)
		r.writeMask(i2cSarADC, i2cSarADCHostID, adc2InitCodeLowAddr, adc2InitCodeLowMSB, adc2InitCodeLowLSB, lsb)
	}
}

// waitIdle polls the REGI2C master BUSY bit until it clears or the
// simple software timeout expires. This matches the busy‑wait helper
// used in ESP‑IDF's regi2c_ctrl.c.
func (r *regI2C) waitIdle(reg *volatile.Register32) bool {
	for i := 0; i < i2cMstTimeout; i++ {
		if reg.Get()&i2cMstBusyBit == 0 {
			return true
		}
	}
	return false
}

// writeMask is a software implementation of REGI2C_WRITE_MASK macro:
//  1. select block + regAddr,
//  2. read current byte,
//  3. update only [msb:lsb] bitfield,
//  4. write it back via internal I2C master.
func (r *regI2C) writeMask(block, hostID, regAddr, msb, lsb, data uint8) {
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
	mask := uint32(1<<(msb-lsb+1)-1) << lsb
	cur &^= mask
	cur |= uint32(data&(1<<(msb-lsb+1)-1)) << lsb
	reg.Set(uint32(block) | uint32(regAddr)<<8 | i2cMstWrCntl | (cur<<i2cMstDataShift)&i2cMstDataMask)
	r.waitIdle(reg)
}
