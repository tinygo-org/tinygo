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

// InitADC initialises the APB_SARADC and Modem/ADC peripheral on ESP32-C6.
// On C6 the clock/reset gating moved to PCR (not SYSTEM as on C3), and the
// SARADC CLKM divider configuration also lives in PCR.
func InitADC() {
	esp.PCR.SetSARADC_CONF_SARADC_CLK_EN(1)

	// Reset and enable the SARADC bus clock via PCR.
	esp.PCR.SetSARADC_CONF_SARADC_REG_CLK_EN(1)
	esp.PCR.SetSARADC_CONF_SARADC_RST_EN(1)
	esp.PCR.SetSARADC_CONF_SARADC_RST_EN(0)
	esp.PCR.SetSARADC_CONF_SARADC_REG_RST_EN(1)
	esp.PCR.SetSARADC_CONF_SARADC_REG_RST_EN(0)

	modemClockModuleEnableForADC()

	// Enable REG_I2C: Enter regi2c reset mode
	esp.PMU.SetRF_PWC_PERIF_I2C_RSTB(0)
	// Enable REGI2C for SAR_ADC and TSENS
	esp.PMU.SetRF_PWC_XPD_PERIF_I2C(1)
	// Release regi2c reset mode, enter work mode
	esp.PMU.SetRF_PWC_PERIF_I2C_RSTB(1)

	// Select clock source 1 (PLL_F80M)
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_SEL(1)
	// At boot we have the CLKM values (4, 0, 0) but in ADC one-shot mode
	// the CLKM values are changed to (15, 1, 0) (source: esp-idf hal/adc_oneshot_hal.c)
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_DIV_NUM(15) // see ADC_LL_CLKM_DIV_NUM_DEFAULT hal/esp32c6/include/hal/adc_ll.h
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_DIV_B(1)
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_DIV_A(0)
	esp.PCR.SetSARADC_CLKM_CONF_SARADC_CLKM_EN(1)

	// Power up the SAR ADC in software  mode (3), no FSM config needed, FSM_REG defaults to 0x00FF0808
	esp.APB_SARADC.SetCTRL_SARADC_XPD_SAR_FORCE(3)
	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_GATED(1)
	esp.APB_SARADC.SetCTRL_SARADC_SAR_CLK_DIV(1)

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

	// Clear stale DONE
	esp.APB_SARADC.SetINT_CLR_APB_SARADC1_DONE_INT_CLR(1)

	// Disable oneshot conversion trigger for ADC unit
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)

	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_CHANNEL(uint32(a.Pin))
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_ATTEN(atten11dB)

	// Select ADC1 one-time mode
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(1)

	// Trigger one-time conversion
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	// No delay needed here because adc_ctrl_clk is fast (>= APB_CLK_FREQ/8).
	// ESP-IDF evaluates this condition and calls esp_rom_delay_us(0).
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(1)

	for esp.APB_SARADC.GetINT_RAW_APB_SARADC1_DONE_INT_RAW() == 0 {
	}
	raw := esp.APB_SARADC.GetSAR1DATA_STATUS_APB_SARADC1_DATA()
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC_ONETIME_START(0)
	esp.APB_SARADC.SetONETIME_SAMPLE_SARADC1_ONETIME_SAMPLE(0)

	return uint16(raw&0xfff) << 4
}

// LP_I2C_ANA_MST I2C0_CTRL bit-field shifts.
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

// Enables the modem FE clock for the ADC.
// Ported from s_sar_power_acquire() in esp_hw_support/port/esp32c6/sar_periph_ctrl.c.
func modemClockModuleEnableForADC() {
	initModemClocks()

	// Enable Modem Clk - ADC ( => see modem_clock_device_enable(ctx, 1))
	esp.MODEM_SYSCON.SetCLK_CONF1_CLK_FE_APB_EN(1)
	esp.MODEM_SYSCON.SetCLK_CONF1_CLK_FE_80M_EN(1)
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

// Sets the default modem ICG bitmaps.
// Ported from modem_clock_module_icg_map_init_all() in esp_hw_support/modem_clock.c.
func initModemClocks() {
	for domain := modemClockDomainModemAPB; domain < modemClockDomainMax; domain++ {
		bitmap := getInitialGatingMode(domain)
		modemClockSetClockDomainICGBitmap(domain, bitmap)
	}
}

// Ported from initial_gating_mode in esp_hw_support/modem_clock.c.
func getInitialGatingMode(domain c6ModemClockDomain) uint32 {
	switch domain {
	case modemClockDomainModemAPB:
		return esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_MODEM_APB_ST_MAP() | icgNogatingActive | icgNogatingModem
	case modemClockDomainModemPeriph:
		return esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_MODEM_PERI_ST_MAP() | icgNogatingActive
	case modemClockDomainWiFi:
		return esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_WIFI_ST_MAP() | icgNogatingActive | icgNogatingModem
	case modemClockDomainBT:
		return esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_BT_ST_MAP() | icgNogatingActive
	case modemClockDomainModemFE:
		return esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_FE_ST_MAP() | icgNogatingActive | icgNogatingModem
	case modemClockDomainIEEE802154:
		return esp.MODEM_SYSCON.GetCLK_CONF_POWER_ST_CLK_ZB_ST_MAP() | icgNogatingActive
	case modemClockDomainLPAPB:
		return esp.MODEM_LPCON.GetCLK_CONF_POWER_ST_CLK_LP_APB_ST_MAP() | icgNogatingActive | icgNogatingModem
	case modemClockDomainI2CMaster:
		return esp.MODEM_LPCON.GetCLK_CONF_POWER_ST_CLK_I2C_MST_ST_MAP() | icgNogatingActive | icgNogatingModem
	case modemClockDomainCoex:
		return esp.MODEM_LPCON.GetCLK_CONF_POWER_ST_CLK_COEX_ST_MAP() | icgNogatingActive | icgNogatingModem
	case modemClockDomainWiFiPwr:
		return esp.MODEM_LPCON.GetCLK_CONF_POWER_ST_CLK_WIFIPWR_ST_MAP() | icgNogatingActive | icgNogatingModem
	}
	return 0
}

// Ported from hal/esp32c6/modem_clock_hal.c.
func modemClockSetClockDomainICGBitmap(domain c6ModemClockDomain, bitmap uint32) {
	switch domain {
	case modemClockDomainModemAPB:
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_MODEM_APB_ST_MAP(bitmap)
	case modemClockDomainModemPeriph:
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_MODEM_PERI_ST_MAP(bitmap)
	case modemClockDomainWiFi:
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_WIFI_ST_MAP(bitmap)
	case modemClockDomainBT:
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_BT_ST_MAP(bitmap)
	case modemClockDomainModemFE:
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_FE_ST_MAP(bitmap)
	case modemClockDomainIEEE802154:
		esp.MODEM_SYSCON.SetCLK_CONF_POWER_ST_CLK_ZB_ST_MAP(bitmap)
	case modemClockDomainLPAPB:
		esp.MODEM_LPCON.SetCLK_CONF_POWER_ST_CLK_LP_APB_ST_MAP(bitmap)
	case modemClockDomainI2CMaster:
		esp.MODEM_LPCON.SetCLK_CONF_POWER_ST_CLK_I2C_MST_ST_MAP(bitmap)
	case modemClockDomainCoex:
		esp.MODEM_LPCON.SetCLK_CONF_POWER_ST_CLK_COEX_ST_MAP(bitmap)
	case modemClockDomainWiFiPwr:
		esp.MODEM_LPCON.SetCLK_CONF_POWER_ST_CLK_WIFIPWR_ST_MAP(bitmap)
	}
}
