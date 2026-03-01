//go:build esp32s3

// ESP32-S3: 2 SAR ADCs, 12-bit hardware; Get() returns 0..65520 (scaled from 12-bit).
// Pin mapping: ADC1 = GPIO 1..10 (channel = GPIO-1); ADC2 = GPIO 11..20 (channel = GPIO-11).
// Get() returns raw, uncalibrated ADC values; accurate 0–3.3V mapping should be done
// either by a two-point calibration in user code or by using the eFuse-based
// calibration logic (see IDF adc_cali / our ADCSelfCalibrate implementation).
//
// Registers used (TRM / IDF):
//   SYSTEM:     PERIP_RST_EN0.APB_SARADC_RST, PERIP_CLK_EN0.APB_SARADC_CLK_EN
//   RTC_CNTL:   ANA_CONF.SAR_I2C_PU, I2C_RESET_POR_FORCE_PU
// ADC1 RTC path (oneshot, TRM/IDF):
//   SENS.SAR_MEAS1_MUX.SAR1_DIG_FORCE = 0  → ADC1 under RTC (not digital/APB)
//   SENS.SAR_MEAS1_CTRL2.MEAS1_START_FORCE = 1, SAR1_EN_PAD_FORCE = 1  → SW triggers and selects channel
//   Per conversion: set attenuation (SAR_ATTEN1), channel (SAR1_EN_PAD), then MEAS1_START_SAR 0→1; wait MEAS1_DONE_SAR; read MEAS1_DATA_SAR.
//   SENS.SAR_MEAS1_CTRL1: amp/ref (FORCE_XPD_AMP etc). SAR_MEAS1_CTRL2: MEAS1_DONE_SAR (done), MEAS1_START_SAR (start), MEAS1_DATA_SAR (12-bit result).
//   APB_SARADC: FSM_WAIT, CLKM, etc. used for clock/shared logic; ADC2 uses ARB_CTRL.

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

// newRegI2C returns the regI2C configured for ESP32-S3: hostID=1, drefInit=4.
func newRegI2C() regI2C { return regI2C{hostID: 1, drefInit: 4} }

var adcDigiRefMv uint32

func InitADC() {
	// SYSTEM: reset and enable APB_SARADC clock so SAR registers are accessible.
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(1)
	esp.SYSTEM.SetPERIP_CLK_EN0_APB_SARADC_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN0_APB_SARADC_RST(0)

	// SENS.SAR_PERI_CLK_GATE_CONF: enable SENS SAR peripheral clock (matches Arduino/IDF runtime state).
	esp.SENS.SetSAR_PERI_CLK_GATE_CONF_SARADC_CLK_EN(1)

	// RTC_CNTL.ANA_CONF: keep internal SAR I2C (regI2C analog bus) powered and out of reset.
	esp.RTC_CNTL.SetANA_CONF_I2C_RESET_POR_FORCE_PD(0)
	esp.RTC_CNTL.SetANA_CONF_SAR_I2C_PU(1)
	esp.RTC_CNTL.SetANA_CONF_I2C_RESET_POR_FORCE_PU(1)

	// SENS.SAR_POWER: power up SAR analog block and enable SAR internal clock.
	esp.SENS.SetSAR_POWER_XPD_SAR_FORCE_XPD_SAR(3)
	esp.SENS.SetSAR_POWER_XPD_SAR_SARCLK_EN(1)

	// SENS.SAR_MEAS1_CTRL1: force ADC1 front-end amplifier and reference on in RTC oneshot mode.
	esp.SENS.SetSAR_MEAS1_CTRL1_FORCE_XPD_AMP(3)
	esp.SENS.SetSAR_MEAS1_CTRL1_AMP_RST_FB_FORCE(3)
	esp.SENS.SetSAR_MEAS1_CTRL1_AMP_SHORT_REF_FORCE(3)
	esp.SENS.SetSAR_MEAS1_CTRL1_AMP_SHORT_REF_GND_FORCE(3)

	// SENS.SAR_AMP_CTRL1/2: amplifier/reference settling timings (same as cold-boot defaults).
	esp.SENS.SetSAR_AMP_CTRL1_SAR_AMP_WAIT1(10)
	esp.SENS.SetSAR_AMP_CTRL1_SAR_AMP_WAIT2(10)
	esp.SENS.SetSAR_AMP_CTRL2_SAR_XPD_SAR_AMP_FSM_IDLE(1)
	esp.SENS.SetSAR_AMP_CTRL2_SAR_AMP_SHORT_REF_GND_FSM_IDLE(1)

	// ADC2 uses the same InitADC() as ADC1 (shared APB_SARADC clock/FSM).
	// SENS.SAR_MEAS2_CTRL1: ADC2 FSM wait timings for power-up/reset/standby.
	esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_XPD_WAIT(8)
	esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_RSTB_WAIT(8)
	esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_STANDBY_WAIT(100)
	esp.SENS.SetSAR_MEAS2_CTRL1_SAR_SAR2_RSTB_FORCE(3)

	// SENS.SAR_MEAS1_MUX / SAR_MEAS1_CTRL2: route ADC1 to RTC controller and use SW to select channel/start.
	esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)      // 0 = controlled by RTC/SENS, not digital/APB.
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1) // SW triggers conversion.
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1) // SW selects which ADC1 pad is enabled.

	// APB_SARADC: shared FSM/clock config used by both ADC units and the ADC2 arbiter.
	esp.APB_SARADC.SetFSM_WAIT_SARADC_XPD_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_RSTB_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_STANDBY_WAIT(100)
	esp.APB_SARADC.SetCTRL_SARADC_XPD_SAR_FORCE(3)
	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_GATED(1)
	esp.APB_SARADC.SetCTRL2_SARADC_SAR1_INV(0)
	esp.APB_SARADC.SetCTRL2_SARADC_SAR2_INV(0)
	esp.APB_SARADC.SetCLKM_CONF_CLK_SEL(2)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_NUM(1)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_B(0)
	esp.APB_SARADC.SetCLKM_CONF_CLKM_DIV_A(0)
	esp.APB_SARADC.SetCLKM_CONF_CLK_EN(1)
	esp.APB_SARADC.SetFILTER_CTRL1_FILTER_FACTOR0(0)
	esp.APB_SARADC.SetFILTER_CTRL1_FILTER_FACTOR1(0)

	adcSelfCalibrate()
	adcDigiRefMv = getDigiRef()
}

const (
	attenDefault = 3 // 11 dB, ~0..3.3 V (IDF ADC_ATTEN_DB_12)
)

func setSensAtten1(ch, atten uint32) {
	// SENS.SAR_ATTEN1: 2 bits per channel
	v := esp.SENS.GetSAR_ATTEN1()
	v &^= 3 << (ch * 2)
	v |= (atten & 3) << (ch * 2)
	esp.SENS.SetSAR_ATTEN1(v)
}

func setSensAtten2(ch, atten uint32) {
	// SENS.SAR_ATTEN2: 2 bits per channel
	v := esp.SENS.GetSAR_ATTEN2()
	v &^= 3 << (ch * 2)
	v |= (atten & 3) << (ch * 2)
	esp.SENS.SetSAR_ATTEN2(v)
}

func (a ADC) Configure(config ADCConfig) error {
	if a.Pin < 1 || a.Pin > 20 {
		return errors.New("invalid ADC pin for ESP32-S3")
	}
	a.Pin.Configure(PinConfig{Mode: PinAnalog})

	return nil
}

func (a ADC) Get() uint16 {
	if a.Pin < 1 || a.Pin > 20 {
		return 0
	}

	var ch uint32
	var raw uint32
	if a.Pin <= 10 {
		ch = uint32(a.Pin - 1) // GPIO1→ch0 … GPIO10→ch9
		esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)
		setSensAtten1(ch, attenDefault)
		esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD(1 << ch)
		for esp.SENS.GetSAR_SLAVE_ADDR1_SAR_SARADC_MEAS_STATUS() != 0 {
		}
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(0)
		esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(1)
		for esp.SENS.GetSAR_MEAS1_CTRL2_MEAS1_DONE_SAR() == 0 {
		}
		raw = esp.SENS.GetSAR_MEAS1_CTRL2_MEAS1_DATA_SAR()
	} else {
		ch = uint32(a.Pin - 11) // GPIO11→ch0 … GPIO20→ch9
		// SENS.SAR_MEAS2_CTRL2: force SW control, select channel
		esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_FORCE(1)
		esp.SENS.SetSAR_MEAS2_CTRL2_SAR2_EN_PAD_FORCE(1)
		esp.SENS.SetSAR_MEAS2_CTRL2_SAR2_EN_PAD(1 << ch)
		setSensAtten2(ch, attenDefault)
		// APB_SARADC.ARB_CTRL: grant ADC2 to APB for oneshot
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(1)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(1)
		// SENS.SAR_MEAS2_CTRL2.MEAS2_START_SAR: one-shot start
		esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_SAR(0)
		esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_SAR(1)
		for esp.SENS.GetSAR_MEAS2_CTRL2_MEAS2_DONE_SAR() == 0 {
		}
		raw = esp.SENS.GetSAR_MEAS2_CTRL2_MEAS2_DATA_SAR()
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_APB_FORCE(0)
		esp.APB_SARADC.SetARB_CTRL_ADC_ARB_GRANT_FORCE(0)
	}

	return uint16(raw&0xfff) << 4
}

func (a ADC) GetVoltage() (raw uint32, v float64) {
	const samples = 4
	var sum uint32
	for i := 0; i < samples; i++ {
		sum += uint32(a.Get())
	}
	raw = sum / samples

	// Default full-scale for 11 dB is approximately 3.3 V assuming
	// Vref ≈ 1.1 V and gain ≈ 3. If eFuse provided a per-chip DIGI_REF
	// (Vref in mV) via adcCalibration, use it to adjust the
	// full-scale range instead.
	scale := 3.3
	if adcDigiRefMv != 0 {
		scale = 3.0 * float64(adcDigiRefMv) / 1000.0
	}

	v = float64(raw) / 65520.0 * scale
	return raw, v
}

// ADC hardware self-calibration for ESP32-S3.
//
// Mapping to ESP-IDF (adc_hal_common.c, hal/esp32s3/adc_ll.h):
//   - adc_hal_self_calibration()      → ADCSelfCalibrate()
//   - adc_ll_calibration_init()       → regI2C.calibrationInit (DREF=4);
//                                      in IDF it is not called from self_cal, we call it explicitly.
//   - adc_ll_calibration_prepare()    → SarEnable + calibrationPrepare (ENCAL_GND=1)
//   - adc_ll_calibration_finish()     → calibrationFinish (ENCAL_GND=0)
//   - adc_ll_set_calibration_param()  → setCalibrationParam()
//   - read_cal_channel()              → adcCalibration.readADC1():
//                                      wait for meas_status==0, start 0→1, wait done, read data
//                                      (similar to adc_oneshot_ll_start + get_raw_result).
//   - Loop: 10 iterations, code 0..4096, binary search on self_cal==0; drop min/max;
//           rounding (remainder%8 < 4 without +1, otherwise +1) — same as in adc_hal_common.c.
//   - raw_check_valid: for ADC1 in IDF always true — we do not check it.
//
// Differences:
//   - regI2C: not ROM helper but direct access to 0x6000E000 (protocol like I2C_RTC_CONFIG2).
//   - cal_setup: same SENS/atten/controller fields, but through our registers.
//   - Result is stored only in hardware for the current session (not in eFuse).
//   - eFuse V1: init_code and digi_ref are taken from eFuse — same idea as Arduino/IDF.

const (
	adcCalTimes     = 10
	adcCalRtcMagic  = uint32(0xADC1C401)
	adcCalInitMin   = uint32(2000)
	adcCalInitMax   = uint32(3900)
	adcDigiRefMinMv = uint32(920)
	adcDigiRefMaxMv = uint32(1150)
)

// adcCalibration encapsulates the self-calibration flow for ADC1
// and remembers per-chip calibration data (such as DIGI_REF) when it is
// available from eFuse.
func adcSelfCalibrate() {
	reg := newRegI2C()
	f := fuse{}

	if vref, ok := f.adc1DigiRefAtten3(); ok {
		adcDigiRefMv = vref
	}

	if saved, ok := restoreFromRTC(); ok {
		reg.sarEnable()
		reg.calibrationInit(0)
		adc1CalibrateHigh(reg, saved)
		return
	}

	initCode, useEfuse := f.adc1InitCodeAtten3()
	adc1CalibrationSetup(reg)

	if useEfuse {
		saveToRTC(initCode)
		adc1CalibrateHigh(reg, initCode)
		return
	}

	finalCode := reg.calibrateBinarySearch(0, adcCalTimes, readADC1)
	saveToRTC(finalCode)
	adc1CalibrateHigh(reg, finalCode)
}

func getDigiRef() uint32 {
	return adcDigiRefMv
}

func adc1CalibrationSetup(reg regI2C) {
	reg.sarEnable()

	esp.SENS.SetSAR_MEAS1_MUX_SAR1_DIG_FORCE(0)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(0)
	esp.SENS.SetSAR_MEAS2_CTRL2_MEAS2_START_FORCE(0)
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD(0)
	setSensAtten1(0, attenDefault)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)

	reg.calibrationInit(0)
	reg.calibrationPrepare(0)
}

func adc1CalibrateHigh(reg regI2C, code uint32) {
	reg.setCalibrationParam(0, code)
	reg.calibrationFinish(0)
	adc1StartWithPadForce()
}

func adc1StartWithPadForce() {
	esp.SENS.SetSAR_MEAS1_CTRL2_SAR1_EN_PAD_FORCE(1)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_FORCE(1)
}

// readADC1 performs one ADC1 conversion via RTC path (used during calibration).
// Internal GND is connected via ENCAL_GND, so the pin input is disconnected.
// Matches IDF: wait conversion idle (meas_status==0), then start 0→1, wait done, read data.
func readADC1() uint32 {
	for esp.SENS.GetSAR_SLAVE_ADDR1_SAR_SARADC_MEAS_STATUS() != 0 {
	}
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(0)
	esp.SENS.SetSAR_MEAS1_CTRL2_MEAS1_START_SAR(1)
	for esp.SENS.GetSAR_MEAS1_CTRL2_MEAS1_DONE_SAR() == 0 {
	}
	return uint32(esp.SENS.GetSAR_MEAS1_CTRL2_MEAS1_DATA_SAR() & 0xfff)
}

func restoreFromRTC() (uint32, bool) {
	if esp.RTC_CNTL.GetSTORE0() != adcCalRtcMagic {
		return 0, false
	}
	code := esp.RTC_CNTL.GetSTORE1()
	if code < adcCalInitMin || code > adcCalInitMax {
		return 0, false
	}
	return code, true
}

func saveToRTC(code uint32) {
	esp.RTC_CNTL.SetSTORE0(adcCalRtcMagic)
	esp.RTC_CNTL.SetSTORE1(code)
}

// fuse
const (
	// Base address for eFuse controller (EFUSE_BLKx region in TRM).
	efuseBase = uintptr(0x60007000)

	// EFUSE_*_REG offsets mirror ESP-IDF's efuse_reg.h layout.
	efuseClkReg        = efuseBase + 0x1c8
	efuseConfReg       = efuseBase + 0x1cc
	efuseCmdReg        = efuseBase + 0x1d4
	efuseDacConfReg    = efuseBase + 0x1e8
	efuseWrTimConf1Reg = efuseBase + 0x1f4
	efuseWrTimConf2Reg = efuseBase + 0x1f8
	efuseRdData4Reg    = efuseBase + 0x6c // EFUSE_RD_WR_DIS_REG / RD_DATA4
	efuseRdData5Reg    = efuseBase + 0x70 // EFUSE_RD_REPEAT_DATA1_REG / RD_DATA5
	efuseRdData7Reg    = efuseBase + 0x78 // EFUSE_RD_REPEAT_DATA3_REG / RD_DATA7

	// Read opcode and clock enable bit used by EFUSE HAL (see efuse_ll).
	efuseReadOpCode   = uint32(0x5AA5)
	efuseClkEnBit     = uint32(1 << 16)
	efuseBlkVersionV1 = 1 // EFUSE_BLK_VERSION major version = 1

	// SYSTEM_PERIP_CLK_EN0 register and EFUSE clock gate bit.
	systemPeripClkEn0   = uintptr(0x600C0018)
	systemEfuseClkEnBit = uint32(1 << 14)
)

type fuse struct{}

// adc1InitCodeAtten3 extracts the ADC1 INIT_CODE (offset trim) for
// attenuation index 3 (typically 11 dB) from EFUSE_BLK2. This mirrors
// the logic used by ESP-IDF's ADC calibration HAL for ESP32-S3.
//
// The code is built from four differential eFuse fields (diff0..diff3)
// and constant offsets (1850, 90, 70) as described in Espressif's
// internal calibration formulas.
func (f *fuse) adc1InitCodeAtten3() (uint32, bool) {
	for try := 0; try < 2; try++ {
		f.triggerReadSequence()
		data4, data5, blkVer := f.readBlock2Data4Data5()
		if blkVer != efuseBlkVersionV1 {
			continue
		}
		diff0 := (data4 >> 21) & 0xFF
		diff1 := (data4 >> 29) | ((data5 & 7) << 3)
		diff2 := (data5 >> 3) & 0x3F
		diff3 := (data5 >> 9) & 0x3F
		icode0 := diff0 + 1850
		icode1 := diff1 + icode0 + 90
		icode2 := diff2 + icode1
		icode3 := diff3 + icode2 + 70
		if icode3 >= adcCalInitMin && icode3 <= adcCalInitMax {
			return icode3, true
		}
	}
	return 0, false
}

// adc1DigiRefAtten3 reads the digital reference (DIGI_REF) for
// ADC1 at attenuation index 3 from EFUSE_BLK2 / RD_DATA7. This is
// similar to what the ESP-IDF ADC calibration HAL uses when present.
func (f *fuse) adc1DigiRefAtten3() (uint32, bool) {
	f.triggerReadSequence()
	_, _, blkVer := f.readBlock2Data4Data5()
	if blkVer != efuseBlkVersionV1 {
		return 0, false
	}
	data7 := f.readBlock2Data7()
	diff3 := (data7 >> 1) & 0xFF
	digiRef := diff3 + 900
	if digiRef < adcDigiRefMinMv || digiRef > adcDigiRefMaxMv {
		return 0, false
	}
	return digiRef, true
}

// triggerReadSequence performs one eFuse read operation using the
// controller's timing/opcode sequence. This roughly corresponds to
// the low-level logic in the ESP-IDF eFuse HAL (see efuse_ll_* in
// the IDF sources and the "eFuse Manager" docs:
// https://docs.espressif.com/projects/esp-idf/en/latest/esp32s3/api-reference/system/efuse.html).
func (f *fuse) triggerReadSequence() {
	clk := (*volatile.Register32)(unsafe.Pointer(systemPeripClkEn0))
	clk.Set(clk.Get() | systemEfuseClkEnBit)
	efuseClk := (*volatile.Register32)(unsafe.Pointer(efuseClkReg))
	efuseClk.Set(efuseClk.Get() | efuseClkEnBit)
	dac := (*volatile.Register32)(unsafe.Pointer(efuseDacConfReg))
	dac.Set(0x28 | (0xFF << 9))
	(*volatile.Register32)(unsafe.Pointer(efuseWrTimConf1Reg)).Set(0x3000 << 8)
	(*volatile.Register32)(unsafe.Pointer(efuseWrTimConf2Reg)).Set(0x190)
	(*volatile.Register32)(unsafe.Pointer(efuseConfReg)).Set(efuseReadOpCode)
	cmd := (*volatile.Register32)(unsafe.Pointer(efuseCmdReg))
	cmd.Set(1)
	for cmd.Get()&1 != 0 {
	}
}

// readBlock2Data4Data5 reads the EFUSE_BLK2 data words that contain
// ADC calibration and version information. It returns RD_DATA4,
// RD_DATA5 and the decoded block version (BLK_VERSION).
//
// Layout is derived from the ESP32-S3 TRM and IDF eFuse tables.
func (f *fuse) readBlock2Data4Data5() (data4, data5 uint32, blkVer uint8) {
	data4 = (*volatile.Register32)(unsafe.Pointer(efuseRdData4Reg)).Get()
	data5 = (*volatile.Register32)(unsafe.Pointer(efuseRdData5Reg)).Get()
	blkVer = uint8(data4 & 3)
	return data4, data5, blkVer
}

// readBlock2Data7 reads RD_DATA7 from EFUSE_BLK2, which for ADC
// calibration contains additional reference (DIGI_REF) data fields.
func (f *fuse) readBlock2Data7() uint32 {
	return (*volatile.Register32)(unsafe.Pointer(efuseRdData7Reg)).Get()
}

// readAdcCalibBlock2 triggers an eFuse read and returns the raw
// EFUSE_BLK2 words used for ADC calibration (RD_DATA4/5) along
// with the decoded block version. This is a small helper similar
// in spirit to the internal IDF helpers around EFUSE_BLK2.
func (f *fuse) readAdcCalibBlock2() (data4, data5 uint32, blkVer uint8) {
	f.triggerReadSequence()
	return f.readBlock2Data4Data5()
}
