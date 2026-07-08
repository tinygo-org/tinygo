//go:build esp32c6

package machine

import (
	"device/esp"
	"errors"
	"runtime/volatile"
	"unsafe"
)

// newRegI2C returns the regI2C configured for ESP32-C6: hostID=0, drefInit=1.
// I2C_SAR_ADC_HOSTID = 0 per soc/esp32c6/include/soc/regi2c_saradc.h.
func newRegI2C() regI2C { return regI2C{hostID: 0, drefInit: 1} }

const (
	// ADC attenuation values for ESP32-C6 APB_SARADC.
	// 0 dB  : ~0 .. 1.1 V
	// 11 dB : ~0 .. 3.3 V (matches typical VDD)
	atten0dB  = 0
	atten11dB = 3
)

// InitADC initialises the APB_SARADC peripheral on ESP32-C6.
// On C6 the clock/reset gating moved to PCR (not SYSTEM as on C3), and the
// SARADC CLKM divider configuration also lives in PCR.
func InitADC() {
	// Reset and enable the SARADC bus clock via PCR.
	esp.PCR.SetSARADC_CONF_SARADC_RST_EN(1)
	esp.PCR.SetSARADC_CONF_SARADC_CLK_EN(1)
	esp.PCR.SetSARADC_CONF_SARADC_RST_EN(0)

	// Select clock source 2 (PLL_F80M), divider = 1, no fractional.
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_SEL(2)
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_DIV_NUM(1)
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_DIV_B(0)
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_DIV_A(0)
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_EN(1)

	// Power up the SAR ADC and configure FSM timing (same register layout as C3).
	esp.APB_SARADC.SetCTRL_SARADC_XPD_SAR_FORCE(1)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_XPD_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_RSTB_WAIT(8)
	esp.APB_SARADC.SetFSM_WAIT_SARADC_STANDBY_WAIT(100)

	adcSelfCalibrate()
}

// ESP32-C6 ADC pin mapping: ADC1 = GPIO0–GPIO6 (ch 0–6). There is no ADC2.
// (The machine_esp32c6.go file defines ADC0..ADC6 as GPIO0..GPIO6.)
func (a ADC) Configure(config ADCConfig) error {
	if a.Pin > 6 {
		return errors.New("invalid ADC pin for ESP32-C6")
	}
	a.Pin.Configure(PinConfig{Mode: PinAnalog})
	return nil
}

// Get performs a single ADC1 conversion and returns a 16-bit value.
// The raw 12-bit result (0..4095) is left-shifted by 4 to fill 16 bits.
func (a ADC) Get() uint16 {
	if a.Pin > 6 {
		return 0
	}
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(uint32(a.Pin))
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA()
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)
	return uint16(raw&0xfff) << 4
}

// ── regI2C: internal I2C-bus (LP_I2C_ANA_MST) for SAR ADC calibration ───────
//
// On ESP32-C6 the "REGI2C" master moved from the embedded SENS/APB_SARADC
// controller (0x6000_E000) used on C3/S3 to the dedicated LP_I2C_ANA_MST
// peripheral at 0x600b_2400. The SAR ADC block address and register layout
// (DREF, ENCAL_GND, INIT_CODE) remain identical to C3.
//
// LP_I2C_ANA_MST.I2C0_CTRL bit layout (25-bit command field):
//   [7:0]  = slave block address (0x69 for I2C_SAR_ADC)
//   [15:8] = register address within the block
//   [23:16]= write data (8 bits)
//   [24]   = WR_CNTL: 0=read, 1=write
//   [25]   = BUSY (read-only, set by hardware while processing)
//
// Source: components/esp_rom/patches/esp_rom_regi2c_esp32c6.c in esp-idf

// regI2C wraps the internal I2C bus used for SAR ADC calibration registers.
// Fields hold chip-specific parameters.
type regI2C struct {
	// hostID is the I2C_SAR_ADC_HOSTID (0 for ESP32-C6, matching regi2c_saradc.h).
	hostID uint8
	// drefInit is the DREF reference value written during calibrationInit (1 for C6).
	drefInit uint8
}

// SAR ADC I2C register layout — identical to ESP32-C3 / ESP32-S3.
// Source: soc/esp32c6/include/soc/regi2c_saradc.h
const (
	i2cSarADC = uint8(0x69)

	adc1DrefAddr = uint8(0x2)
	adc1DrefMSB  = uint8(6)
	adc1DrefLSB  = uint8(4)
	adc2DrefAddr = uint8(0x5)
	adc2DrefMSB  = uint8(6)
	adc2DrefLSB  = uint8(4)

	adc1EncalGndAddr = uint8(0x7)
	adc1EncalGndMSB  = uint8(5)
	adc1EncalGndLSB  = uint8(5)
	adc2EncalGndAddr = uint8(0x7)
	adc2EncalGndMSB  = uint8(7)
	adc2EncalGndLSB  = uint8(7)

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

	// adcCalOffsetRange is the binary search upper bound (12-bit full scale).
	adcCalOffsetRange = uint32(4096)
	// adcCalMaxIterations caps binary search iterations.
	adcCalMaxIterations = 16
)

// LP_I2C_ANA_MST I2C0_CTRL bit-field shifts (see file header comment).
const (
	c6SlaveIDShift = 0  // bits [7:0]
	c6AddrShift    = 8  // bits [15:8]
	c6DataShift    = 16 // bits [23:16]
	c6WrCntlShift  = 24 // bit [24]
	c6BusyBit      = uint32(1 << 25)

	// c6SarI2CDeviceEn is BIT(7) in LP_I2C_ANA_MST.DEVICE_EN for I2C_SAR_ADC (0x69).
	c6SarI2CDeviceEn = uint32(1 << 7)
)

// ANA_CONFIG / ANA_CONFIG2 register addresses and bits for the internal SAR I2C
// domain on ESP32-C6. These differ from C3's SENS block (0x6000_E044/048).
// Source: soc/esp32c6/include/soc/regi2c_defs.h
const (
	c6AnaConfigReg  = uintptr(0x600AF81C) // clear ANA_I2C_SAR_FORCE_PD (bit 18)
	c6AnaConfig2Reg = uintptr(0x600AF820) // set   ANA_I2C_SAR_FORCE_PU  (bit 16)
	c6SarForcePD    = uint32(1 << 18)
	c6SarForcePU    = uint32(1 << 16)
)

// sarEnable powers up the internal SAR I2C domain and enables the LP_I2C_ANA_MST
// clock and SAR slave device before any regI2C access.
// Matches regi2c_ctrl_ll_i2c_saradc_enable() + regi2c_enable_block(REGI2C_SAR_I2C).
func (r regI2C) sarEnable() {
	cfg := (*volatile.Register32)(unsafe.Pointer(c6AnaConfigReg))
	cfg2 := (*volatile.Register32)(unsafe.Pointer(c6AnaConfig2Reg))
	cfg.Set(cfg.Get() &^ c6SarForcePD)
	cfg2.Set(cfg2.Get() | c6SarForcePU)

	// Enable the LP_I2C_ANA_MST master clock (MODEM_LPCON.CLK_CONF bit 2).
	esp.MODEM_LPCON.SetCLK_CONF_CLK_I2C_MST_EN(1)
	// Enable the master's own clock gate (LP_I2C_ANA_MST.DATE bit 28).
	esp.LP_I2C_ANA_MST.SetDATE_LP_I2C_ANA_MAST_I2C_MAT_CLK_EN(1)
	// Enable the SAR ADC slave device (DEVICE_EN bit 7).
	dev := esp.LP_I2C_ANA_MST.GetDEVICE_EN_LP_I2C_ANA_MAST_I2C_DEVICE_EN()
	esp.LP_I2C_ANA_MST.SetDEVICE_EN_LP_I2C_ANA_MAST_I2C_DEVICE_EN(dev | c6SarI2CDeviceEn)
}

// writeMask implements the REGI2C_WRITE_MASK macro for ESP32-C6 via LP_I2C_ANA_MST.
// It reads the current byte at regAddr, updates the [msb:lsb] bitfield, and writes
// it back. Matches esp_rom_regi2c_write_mask() in esp_rom_regi2c_esp32c6.c.
func (r regI2C) writeMask(regAddr, msb, lsb, data uint8) {
	ctrl := &esp.LP_I2C_ANA_MST.I2C0_CTRL
	rdata := &esp.LP_I2C_ANA_MST.I2C0_DATA

	// Issue a read command: slave_id | (reg_addr << 8), no WR_CNTL bit.
	readCmd := (uint32(i2cSarADC) << c6SlaveIDShift) | (uint32(regAddr) << c6AddrShift)
	volatile.StoreUint32(&ctrl.Reg, readCmd)
	for volatile.LoadUint32(&ctrl.Reg)&c6BusyBit != 0 {
	}
	cur := volatile.LoadUint32(&rdata.Reg) & 0xFF

	// Modify the [msb:lsb] bitfield.
	mask := uint32(1<<(msb-lsb+1)-1) << lsb
	cur &^= mask
	cur |= uint32(data&(1<<(msb-lsb+1)-1)) << lsb

	// Issue a write command: slave_id | (reg_addr<<8) | WR_CNTL | (data<<16).
	writeCmd := (uint32(i2cSarADC) << c6SlaveIDShift) |
		(uint32(regAddr) << c6AddrShift) |
		(uint32(1) << c6WrCntlShift) |
		((cur & 0xFF) << c6DataShift)
	volatile.StoreUint32(&ctrl.Reg, writeCmd)
	for volatile.LoadUint32(&ctrl.Reg)&c6BusyBit != 0 {
	}
}

// calibrationInit sets the DREF reference for the selected ADC unit.
func (r regI2C) calibrationInit(adcN uint8) {
	if adcN == 0 {
		r.writeMask(adc1DrefAddr, adc1DrefMSB, adc1DrefLSB, r.drefInit)
	} else {
		r.writeMask(adc2DrefAddr, adc2DrefMSB, adc2DrefLSB, r.drefInit)
	}
}

// calibrationPrepare enables ENCAL_GND so the ADC input is shorted to ground.
func (r regI2C) calibrationPrepare(adcN uint8) {
	if adcN == 0 {
		r.writeMask(adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 1)
	} else {
		r.writeMask(adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 1)
	}
}

// calibrationFinish clears ENCAL_GND to reconnect the ADC input to the pad.
func (r regI2C) calibrationFinish(adcN uint8) {
	if adcN == 0 {
		r.writeMask(adc1EncalGndAddr, adc1EncalGndMSB, adc1EncalGndLSB, 0)
	} else {
		r.writeMask(adc2EncalGndAddr, adc2EncalGndMSB, adc2EncalGndLSB, 0)
	}
}

// setCalibrationParam writes the INIT_CODE (offset trim) for the selected ADC unit.
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
// It performs 'iterations' rounds, drops the min/max outliers, and returns
// the rounded mean of the remaining values. Matches adc_hal_self_calibration().
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

// ── Self-calibration ──────────────────────────────────────────────────────────

const (
	adcCalTimesC6    = 15
	adcCalRtcMagicC6 = uint32(0xADC1C601) // magic distinguishes C6 from C3
	adcCalInitMinC6  = uint32(1000)
	adcCalInitMaxC6  = uint32(4096)
)

// adcSelfCalibrate runs a self-calibration for ADC1 (the only ADC unit on C6).
// The calibration code is cached in LP_AON scratch registers to survive sleep.
// eFuse calibration is not used: the fields are often unprogrammed.
func adcSelfCalibrate() {
	reg := newRegI2C()
	reg.sarEnable()

	var adc1Code uint32
	if saved, ok := c6RestoreFromLP(); ok {
		adc1Code = saved
	} else {
		c6CalSetupADC1()
		reg.calibrationInit(0)
		reg.calibrationPrepare(0)
		adc1Code = reg.calibrateBinarySearch(0, adcCalTimesC6, readADC1)
		if adc1Code < adcCalInitMinC6 {
			adc1Code = adcCalInitMinC6
		}
		if adc1Code > adcCalInitMaxC6 {
			adc1Code = adcCalInitMaxC6
		}
		c6SaveToLP(adc1Code)
		reg.calibrationFinish(0)
	}

	c6ApplyADC1Code(reg, adc1Code)
}

// c6CalSetupADC1 configures APB_SARADC for oneshot ADC1 ch0 with fixed attenuation.
func c6CalSetupADC1() {
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)
}

// readADC1 performs a single ADC1 conversion and returns the raw 12-bit result.
func readADC1() uint32 {
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)
	for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA() & 0xfff
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	return uint32(raw)
}

// c6RestoreFromLP reads the saved calibration code from LP_AON scratch registers.
// On C6, LP_AON replaces the C3's RTC_CNTL for scratch storage.
func c6RestoreFromLP() (uint32, bool) {
	if esp.LP_AON.GetSTORE0() != adcCalRtcMagicC6 {
		return 0, false
	}
	code := esp.LP_AON.GetSTORE1()
	if code < adcCalInitMinC6 || code > adcCalInitMaxC6 {
		return 0, false
	}
	return code, true
}

// c6SaveToLP stores the calibration code in LP_AON scratch registers.
func c6SaveToLP(code uint32) {
	if code < adcCalInitMinC6 || code > adcCalInitMaxC6 {
		return
	}
	esp.LP_AON.SetSTORE0(adcCalRtcMagicC6)
	esp.LP_AON.SetSTORE1(code)
}

// c6ApplyADC1Code sets ADC1 init code and finishes calibration.
// ESP32-C6 has no ADC2 so only ADC1 (adcN=0) needs to be configured.
func c6ApplyADC1Code(reg regI2C, code uint32) {
	c6CalSetupADC1()
	reg.calibrationInit(0)
	reg.calibrationPrepare(0)
	reg.setCalibrationParam(0, code)
	reg.calibrationFinish(0)
}
