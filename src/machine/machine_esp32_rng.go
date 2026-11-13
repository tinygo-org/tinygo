//go:build esp32

package machine

import (
	"device"
	"device/esp"
	"runtime/volatile"
	_ "unsafe"
)

// DPORT_WIFI_CLK_RNG_EN, see ESP-IDF components/soc/esp32/include/soc/dport_reg.h.
const dportWifiClkRNGEn = 1 << 15

// entropySeedSize must match seedSize in src/crypto/rand/rand_esp32.go.
const entropySeedSize = 128

var entropySeed [entropySeedSize]byte

// entropySeedReady is only read with volatile loads so that interp does not run the crypto/rand init at build time.
var entropySeedReady uint8

// GetRNG returns 32 bits from the hardware random number generator.
// It is only truly random while Wi-Fi or Bluetooth is on, crypto/rand uses a seeded DRBG instead.
func GetRNG() (uint32, error) {
	// Wait at least 16 APB cycles between results (ESP-IDF hw_random.c).
	return readRNG(CPUFrequency() / peripheralClock * 16), nil
}

func readRNG(cycles uint32) uint32 {
	v := esp.RNG.DATA.Get()
	start := cpuCycles()
	for cpuCycles()-start < cycles {
		v ^= esp.RNG.DATA.Get()
	}
	return v
}

func cpuCycles() uint32 {
	return uint32(device.AsmFull("rsr {}, CCOUNT", nil))
}

// takeEntropySeed hands the startup seed to crypto/rand once and then erases it.
//
//go:linkname takeEntropySeed crypto/rand.takeEntropySeed
func takeEntropySeed(dst []byte) bool {
	if volatile.LoadUint8(&entropySeedReady) == 0 || len(dst) != len(entropySeed) {
		return false
	}
	copy(dst, entropySeed[:])
	clear(entropySeed[:])
	volatile.StoreUint8(&entropySeedReady, 0)
	return true
}

// collectEntropy runs at startup before any peripheral is set up. It reads the RNG while SAR ADC noise feeds it.
// Port of bootloader_random_enable, bootloader_fill_random and bootloader_random_disable in ESP-IDF v5.3.
//
//go:linkname collectEntropy runtime.collectEntropy
func collectEntropy() {
	testMux := esp.RTC_CNTL.TEST_MUX.Get()
	startForce := esp.SENS.SAR_START_FORCE.Get()
	measWait2 := esp.SENS.SAR_MEAS_WAIT2.Get()
	readCtrl := esp.SENS.SAR_READ_CTRL.Get()
	readCtrl2 := esp.SENS.SAR_READ_CTRL2.Get()
	saradcCtrl := esp.APB_CTRL.APB_SARADC_CTRL.Get()
	saradcFSM := esp.APB_CTRL.APB_SARADC_FSM.Get()
	patt1 := esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB1.Get()
	patt2 := esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB2.Get()
	patt3 := esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB3.Get()
	patt4 := esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB4.Get()
	i2sClk := esp.DPORT.GetPERIP_CLK_EN_I2S0_CLK_EN()

	esp.DPORT.WIFI_CLK_EN.SetBits(dportWifiClkRNGEn)

	esp.RTC_CNTL.SetTEST_MUX_DTEST_RTC(2)
	esp.RTC_CNTL.SetTEST_MUX_ENT_RTC(1)
	esp.SENS.SetSAR_START_FORCE_SAR2_EN_TEST(1)

	esp.DPORT.SetPERIP_CLK_EN_I2S0_CLK_EN(1)
	esp.SENS.SetSAR_START_FORCE_ULP_CP_FORCE_START_TOP(0)
	esp.SENS.SetSAR_START_FORCE_ULP_CP_START_TOP(0)

	// 0xAD selects the ADC2 test channel, 12 bit, 3 dB.
	esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB1.Set(0xADADADAD)
	esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB2.Set(0xADADADAD)
	esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB3.Set(0xADADADAD)
	esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB4.Set(0xADADADAD)
	esp.SENS.SetSAR_MEAS_WAIT2_FORCE_XPD_SAR(3)
	esp.SENS.SetSAR_READ_CTRL_SAR1_DIG_FORCE(1)
	esp.SENS.SetSAR_READ_CTRL2_SAR2_DIG_FORCE(1)

	esp.APB_CTRL.SetAPB_SARADC_CTRL_SARADC_SAR2_MUX(1)
	esp.APB_CTRL.SetAPB_SARADC_CTRL_SARADC_SAR_CLK_DIV(4)
	esp.APB_CTRL.SetAPB_SARADC_FSM_SARADC_RSTB_WAIT(8)
	esp.APB_CTRL.SetAPB_SARADC_FSM_SARADC_START_WAIT(10)
	esp.APB_CTRL.SetAPB_SARADC_CTRL_SARADC_WORK_MODE(0)
	esp.APB_CTRL.SetAPB_SARADC_CTRL_SARADC_SAR_SEL(1)
	esp.APB_CTRL.SetAPB_SARADC_CTRL_SARADC_DATA_SAR_SEL(0)
	esp.I2S0.SetSAMPLE_RATE_CONF_RX_BCK_DIV_NUM(20)
	esp.APB_CTRL.SetAPB_SARADC_CTRL_SARADC_DATA_TO_I2S(1)

	esp.I2S0.SetCONF2_CAMERA_EN(0)
	esp.I2S0.SetCONF2_LCD_EN(1)
	esp.I2S0.SetCONF2_DATA_ENABLE(1)
	esp.I2S0.SetCONF2_DATA_ENABLE_TEST_EN(1)
	esp.I2S0.SetCONF_RX_START(1)

	// One word per 64 us, the ESP32 pace of RNG_CPU_WAIT_CYCLE_NUM in bootloader_random.c.
	cycles := CPUFrequency() / 1000000 * 64
	for i := 0; i < len(entropySeed); i += 4 {
		v := readRNG(cycles)
		entropySeed[i] = byte(v)
		entropySeed[i+1] = byte(v >> 8)
		entropySeed[i+2] = byte(v >> 16)
		entropySeed[i+3] = byte(v >> 24)
	}
	volatile.StoreUint8(&entropySeedReady, 1)

	esp.I2S0.SetCONF_RX_START(0)
	esp.I2S0.SetCONF_RX_RESET(1)
	esp.I2S0.SetCONF_RX_RESET(0)
	esp.I2S0.SetCONF2_CAMERA_EN(0)
	esp.I2S0.SetCONF2_LCD_EN(0)
	esp.I2S0.SetCONF2_DATA_ENABLE_TEST_EN(0)
	esp.I2S0.SetCONF2_DATA_ENABLE(0)
	esp.DPORT.SetPERIP_RST_EN_I2S0_RST(1)
	esp.DPORT.SetPERIP_RST_EN_I2S0_RST(0)
	esp.DPORT.SetPERIP_CLK_EN_I2S0_CLK_EN(i2sClk)

	esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB1.Set(patt1)
	esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB2.Set(patt2)
	esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB3.Set(patt3)
	esp.APB_CTRL.APB_SARADC_SAR2_PATT_TAB4.Set(patt4)
	esp.APB_CTRL.APB_SARADC_FSM.Set(saradcFSM)
	esp.APB_CTRL.APB_SARADC_CTRL.Set(saradcCtrl)
	esp.SENS.SAR_READ_CTRL2.Set(readCtrl2)
	esp.SENS.SAR_READ_CTRL.Set(readCtrl)
	esp.SENS.SAR_MEAS_WAIT2.Set(measWait2)
	esp.SENS.SAR_START_FORCE.Set(startForce)
	esp.RTC_CNTL.TEST_MUX.Set(testMux)
}
