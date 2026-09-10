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

type c6PWDET_Type struct {
	CONF_REG volatile.Register32 // 0x0
}

// see soc/esp32c6/register/soc/reg_base.h
const (
	c6PWDET_CONF_REG               = 0x600A0810
	c6PWDET_LL_SAR_POWER_FORCE_BIT = 1 << 24
	c6PWDET_LL_SAR_POWER_CNTL_BIT  = 1 << 23
)

var c6PWDET = (*c6PWDET_Type)(unsafe.Pointer(uintptr(c6PWDET_CONF_REG)))

// InitADC initialises the APB_SARADC and Modem/ADC peripheral on ESP32-C6.
// On C6 the clock/reset gating moved to PCR (not SYSTEM as on C3), and the
// SARADC CLKM divider configuration also lives in PCR.
func InitADC() {
	// Reset and enable the SARADC bus clock via PCR.
	esp.PCR.SetSARADC_CONF_SARADC_REG_CLK_EN(1)   // PCR.saradc_conf.saradc_reg_clk_en = 1
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_EN(1) // PCR.saradc_clkm_conf.saradc_clkm_en = 1
	esp.PCR.SetSARADC_CONF_SARADC_RST_EN(1)       // PCR.saradc_conf.saradc_rst_en = 1
	esp.PCR.SetSARADC_CONF_SARADC_RST_EN(0)       // PCR.saradc_conf.saradc_rst_en = 0
	esp.PCR.SetSARADC_CONF_SARADC_REG_RST_EN(1)   // PCR.saradc_conf.saradc_reg_rst_en = 1
	esp.PCR.SetSARADC_CONF_SARADC_REG_RST_EN(0)   // PCR.saradc_conf.saradc_reg_rst_en = 0

	modemClockModuleEnableForADC()

	// Enable REG_I2C: Enter regi2c reset mode
	esp.PMU.SetRF_PWC_PERIF_I2C_RSTB(0) // CLEAR_PERI_REG_MASK(PMU_RF_PWC_REG, PMU_PERIF_I2C_RSTB);
	// Enable REGI2C for SAR_ADC and TSENS
	esp.PMU.SetRF_PWC_XPD_PERIF_I2C(1) // SET_PERI_REG_MASK(PMU_RF_PWC_REG, PMU_XPD_PERIF_I2C);
	// Release regi2c reset mode, enter work mode
	esp.PMU.SetRF_PWC_PERIF_I2C_RSTB(1) // SET_PERI_REG_MASK(PMU_RF_PWC_REG, PMU_PERIF_I2C_RSTB);

	// Enable PWDET see hal at: sar_ctrl_ll_set_power_mode_from_pwdet(SAR_CTRL_LL_POWER_ON);
	c6PWDET.CONF_REG.SetBits(c6PWDET_LL_SAR_POWER_FORCE_BIT) // REG_SET_BIT(PWDET_CONF_REG, PWDET_LL_SAR_POWER_FORCE_BIT);
	c6PWDET.CONF_REG.SetBits(c6PWDET_LL_SAR_POWER_CNTL_BIT)  // REG_SET_BIT(PWDET_CONF_REG, PWDET_LL_SAR_POWER_CNTL_BIT);

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

// ── Modem Clock for ADC ──────────────────────────────────────────────────────

// Enable the clock for the shared ADC and Front-End (FE) controller logic (?).
// ESP-IDF initializes this during ADC setup with:
// modem_clock_module_enable(PERIPH_MODEM_ADC_COMMON_FE_MODULE);
// (see esp_hw_support/port/esp32c6/sar_periph_ctrl.c).
func modemClockModuleEnableForADC() {
	// BEGIN code for modem_clock_module_icg_map_init_all();
	for domain := modemClockDomainModemAPB; domain < modemClockDomainMax; domain++ {
		code := modemClockGetClockDomainICGBitmap(domain)
		modemClockSetClockDomainICGBitmap(domain, initialGatingMode[domain]|code)
	}
	// END code for modem_clock_module_icg_map_init_all();

	// Enable Modem Clk - ADC ( => modem_clock_device_enable(ctx, 1))
	esp.MODEM_SYSCON.SetCLK_CONF1_CLK_FE_APB_EN(1) // hw->clk_conf1.clk_fe_apb_en = 1
	esp.MODEM_SYSCON.SetCLK_CONF1_CLK_FE_80M_EN(1) // hw->clk_conf1.clk_fe_80m_en = 1
}

type c6ModemClockDomain int

const (
	modemClockDomainModemAPB c6ModemClockDomain = iota // see hal/include/hal/modem_clock_types.h
	modemClockDomainModemPeriph
	modemClockDomainWiFi
	modemClockDomainBT
	modemClockDomainModemFE
	modemClockDomainIEEE802154
	modemClockDomainLPAPB
	modemClockDomainI2CMaster
	modemClockDomainCoex
	modemClockDomainWiFiPwr
	modemClockDomainMax
)

const (
	pmuHpIcgModemCodeSleep  = 0 // see esp_hw_support/include/esp_private/esp_pmu.h
	pmuHpIcgModemCodeModem  = 1
	pmuHpIcgModemCodeActive = 2
)

// The ICG code's bit 0, 1 and 2 indicates the ICG state
// of pmu SLEEP, MODEM and ACTIVE mode respectively
const (
	icgNogatingActive = 1 << pmuHpIcgModemCodeActive // see esp_hw_support/modem_clock.c
	icgNogatingSleep  = 1 << pmuHpIcgModemCodeSleep
	icgNogatingModem  = 1 << pmuHpIcgModemCodeModem
)

// initialGatingMode represents the baseline gating configurations per domain.
// see esp_hw_support/modem_clock.c
var initialGatingMode = [modemClockDomainMax]uint32{
	modemClockDomainModemAPB:    icgNogatingActive | icgNogatingModem,
	modemClockDomainModemPeriph: icgNogatingActive,
	modemClockDomainWiFi:        icgNogatingActive | icgNogatingModem,
	modemClockDomainBT:          icgNogatingActive,
	modemClockDomainModemFE:     icgNogatingActive | icgNogatingModem,
	modemClockDomainIEEE802154:  icgNogatingActive,
	modemClockDomainLPAPB:       icgNogatingActive | icgNogatingModem,
	modemClockDomainI2CMaster:   icgNogatingActive | icgNogatingModem,
	modemClockDomainCoex:        icgNogatingActive | icgNogatingModem,
	modemClockDomainWiFiPwr:     icgNogatingActive | icgNogatingModem,
}

// see hal/esp32c6/modem_clock_hal.c
// see soc/esp32c6/include/modem/modem_syscon_struct.h for SysconDev
// see hal/esp32c6/include/hal/modem_lpcon_ll.h for LPConDev
// see hal/esp32c6/include/hal/modem_syscon_ll.h for definition of modem_syscon_ll_get_modem_apb_icg_bitmap, ...
func modemClockSetClockDomainICGBitmap(domain c6ModemClockDomain, bitmap uint32) {
	switch domain {
	case modemClockDomainModemAPB: // modem_syscon_ll_set_modem_apb_icg_bitmap
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_MODEM_APB_ST_MAP(bitmap) //hw->clk_conf_power_st.clk_modem_apb_st_map = bitmap
	case modemClockDomainModemPeriph: // modem_syscon_ll_set_modem_periph_icg_bitmap(hal->syscon_dev, bitmap);
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_MODEM_PERI_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_modem_peri_st_map
	case modemClockDomainWiFi: // modem_syscon_ll_set_wifi_icg_bitmap(hal->syscon_dev, bitmap);
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_WIFI_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_wifi_st_map
	case modemClockDomainBT: // modem_syscon_ll_set_bt_icg_bitmap(hal->syscon_dev, bitmap);
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_BT_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_bt_st_map
	case modemClockDomainModemFE: // modem_syscon_ll_set_fe_icg_bitmap(hal->syscon_dev, bitmap);
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_FE_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_fe_st_map
	case modemClockDomainIEEE802154: // modem_syscon_ll_set_ieee802154_icg_bitmap(hal->syscon_dev, bitmap);
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_ZB_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_zb_st_map
	case modemClockDomainLPAPB: // modem_lpcon_ll_set_lp_apb_icg_bitmap(hal->lpcon_dev, bitmap);
		esp.MODEM_LPCON.SetCLK_CONF_POWER_ST_CLK_LP_APB_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_lp_apb_st_map
	case modemClockDomainI2CMaster: // modem_lpcon_ll_set_i2c_master_icg_bitmap(hal->lpcon_dev, bitmap);
		esp.MODEM_LPCON.SetCLK_CONF_POWER_ST_CLK_I2C_MST_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_i2c_mst_st_map
	case modemClockDomainCoex: // modem_lpcon_ll_set_coex_icg_bitmap(hal->lpcon_dev, bitmap);
		esp.MODEM_LPCON.SetCLK_CONF_POWER_ST_CLK_COEX_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_coex_st_map
	case modemClockDomainWiFiPwr: // modem_lpcon_ll_set_wifipwr_icg_bitmap(hal->lpcon_dev, bitmap);
		esp.MODEM_LPCON.SetCLK_CONF_POWER_ST_CLK_WIFIPWR_ST_MAP(bitmap) // hw->clk_conf_power_st.clk_wifipwr_st_map
	default:
		panic("unhandled domain")
	}
}

func modemClockGetClockDomainICGBitmap(domain c6ModemClockDomain) uint32 {
	var bitmap uint32

	switch domain {
	case modemClockDomainModemAPB:
		bitmap = esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_MODEM_APB_ST_MAP()
	case modemClockDomainModemPeriph:
		bitmap = esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_MODEM_PERI_ST_MAP()
	case modemClockDomainWiFi:
		bitmap = esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_WIFI_ST_MAP()
	case modemClockDomainBT:
		bitmap = esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_BT_ST_MAP()
	case modemClockDomainModemFE:
		bitmap = esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_FE_ST_MAP()
	case modemClockDomainIEEE802154:
		bitmap = esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_ZB_ST_MAP()
	case modemClockDomainLPAPB:
		bitmap = esp.MODEM_LPCON.GetCLK_CONF_POWER_ST_CLK_LP_APB_ST_MAP()
	case modemClockDomainI2CMaster:
		bitmap = esp.MODEM_LPCON.GetCLK_CONF_POWER_ST_CLK_I2C_MST_ST_MAP()
	case modemClockDomainCoex:
		bitmap = esp.MODEM_LPCON.GetCLK_CONF_POWER_ST_CLK_COEX_ST_MAP()
	case modemClockDomainWiFiPwr:
		bitmap = esp.MODEM_LPCON.GetCLK_CONF_POWER_ST_CLK_WIFIPWR_ST_MAP()
	default:
		panic("unhandled domain")
	}
	return bitmap
}

// ── Modem Clock for ADC END ──────────────────────────────────────────────────
