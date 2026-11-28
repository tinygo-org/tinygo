//go:build esp32s3

package touch

import (
	"device/esp"
	"machine"
	"runtime/volatile"
	"unsafe"
)

// initSensor sets up the touch sensor to scan the configured pads in the
// background.
func initSensor() {
	stopScan()
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_RESET(1)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_RESET(0)

	// Values from the ESP-IDF legacy touch_hal_init and
	// hal/touch_sensor_legacy_types.h for ESP32-S3.
	esp.RTC_CNTL.SetTOUCH_CTRL1_TOUCH_MEAS_NUM(500)
	esp.RTC_CNTL.SetTOUCH_CTRL1_TOUCH_SLEEP_CYCLES(0xf)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_XPD_WAIT(0xff)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_DBIAS(1)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_DREFH(3)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_DREFL(0)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_DRANGE(2)
	esp.RTC_CNTL.SetTOUCH_SCAN_CTRL_TOUCH_INACTIVE_CONNECTION(1)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_CLKGATE_EN(1)
	esp.SENS.SetSAR_TOUCH_CHN_ST_SAR_TOUCH_CHANNEL_CLR(0x7fff)
	esp.SENS.SetSAR_TOUCH_CONF_SAR_TOUCH_DATA_SEL(0)

	// Start each scan from the RTC timer.
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_START_FORCE(0)
}

func configurePad(pin machine.Pin, pad uint32) {
	pin.Configure(machine.PinConfig{Mode: machine.PinAnalog})
	padReg(pad).ClearBits(esp.RTC_IO_TOUCH_PAD0_MUX_SEL_Msk | esp.RTC_IO_TOUCH_PAD0_XPD_Msk)

	// Charge speed 7 is the ESP-IDF default.
	if pad < 10 {
		esp.RTC_CNTL.TOUCH_DAC.ReplaceBits(7, 7, uint8(29-pad*3))
	} else {
		esp.RTC_CNTL.TOUCH_DAC1.ReplaceBits(7, 7, uint8(29-(pad-10)*3))
	}

	stopScan()
	esp.RTC_CNTL.SetTOUCH_SCAN_CTRL_TOUCH_SCAN_PAD_MAP(esp.RTC_CNTL.GetTOUCH_SCAN_CTRL_TOUCH_SCAN_PAD_MAP() | 1<<pad)
	esp.SENS.SetSAR_TOUCH_CONF_SAR_TOUCH_OUTEN(esp.SENS.GetSAR_TOUCH_CONF_SAR_TOUCH_OUTEN() | 1<<pad)
	startScan()
}

// readPad returns the count of the latest background scan.
func readPad(pad uint32) uint32 {
	out := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.SENS.SAR_TOUCH_STATUS1), uintptr(pad-1)*4))
	return out.Get() & 0x3fffff
}

func stopScan() {
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_START_EN(0)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_SLP_TIMER_EN(0)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_TIMER_FORCE_DONE(3)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_TIMER_FORCE_DONE(0)
}

func startScan() {
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_TIMER_FORCE_DONE(3)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_TIMER_FORCE_DONE(0)
	esp.RTC_CNTL.SetTOUCH_CTRL2_TOUCH_SLP_TIMER_EN(1)
}

func padNumber(p machine.Pin) (uint32, bool) {
	if p >= machine.GPIO1 && p <= machine.GPIO14 {
		return uint32(p), true
	}
	return 0, false
}
