//go:build esp32

package touch

import (
	"device/esp"
	"errors"
	"machine"
	"runtime/volatile"
	"unsafe"
)

// Pad reads the capacitive touch sensor on one pin. The count drops
// when the pad is touched (ESP32 TRM v4.2 section 29.2.4, page 589).
type Pad struct {
	Pin machine.Pin
}

var ErrInvalidPin = errors.New("touch: invalid touch pin for ESP32")

var initialized bool

// Init puts the touch sensor under software control. Configure calls it
// when it has not run yet.
func Init() {
	initialized = true

	esp.SENS.SetSAR_TOUCH_ENABLE_TOUCH_PAD_WORKEN(0)
	esp.SENS.SetSAR_TOUCH_ENABLE_TOUCH_PAD_OUTEN1(0)
	esp.SENS.SetSAR_TOUCH_ENABLE_TOUCH_PAD_OUTEN2(0)

	// 2.7 V high, 0.5 V low, 1 V attenuation, as in the ESP-IDF touch examples.
	esp.RTC_IO.SetTOUCH_CFG_TOUCH_DREFH(3)
	esp.RTC_IO.SetTOUCH_CFG_TOUCH_DREFL(0)
	esp.RTC_IO.SetTOUCH_CFG_TOUCH_DRANGE(1)

	// Measure for 0x7fff cycles of the 8 MHz clock, the ESP-IDF default.
	esp.SENS.SetSAR_TOUCH_CTRL1_TOUCH_MEAS_DELAY(0x7fff)
	esp.SENS.SetSAR_TOUCH_CTRL1_TOUCH_XPD_WAIT(0xff)

	esp.SENS.SetSAR_TOUCH_CTRL2_TOUCH_START_FSM_EN(1)
	esp.SENS.SetSAR_TOUCH_CTRL2_TOUCH_START_EN(0)
	esp.SENS.SetSAR_TOUCH_CTRL2_TOUCH_START_FORCE(1)
	esp.RTC_CNTL.SetSTATE0_TOUCH_SLP_TIMER_EN(1)
}

// Configure hands the pin to the touch sensor. Only touch pads T0 to T7 are
// supported.
func (t Pad) Configure() error {
	pad, ok := padNumber(t.Pin)
	if !ok {
		return ErrInvalidPin
	}

	if !initialized {
		Init()
	}

	reg := padReg(pad)
	v := reg.Get()
	v &^= esp.RTC_GPIO_TOUCH_PAD0_FUN_SEL_Msk | esp.RTC_GPIO_TOUCH_PAD0_FUN_IE_Msk |
		esp.RTC_GPIO_TOUCH_PAD0_RUE_Msk | esp.RTC_GPIO_TOUCH_PAD0_RDE_Msk |
		esp.RTC_GPIO_TOUCH_PAD0_TIE_OPT_Msk | esp.RTC_GPIO_TOUCH_PAD0_DAC_Msk
	// DAC 7 is the fastest charge rate and the ESP-IDF default.
	v |= esp.RTC_GPIO_TOUCH_PAD0_MUX_SEL_Msk | 7<<esp.RTC_GPIO_TOUCH_PAD0_DAC_Pos
	reg.Set(v)

	// Touch pads T0 to T7 are RTC GPIO 10 to 17.
	esp.RTC_IO.SetENABLE_W1TC(1 << (pad + 10))

	return nil
}

// Get runs one measurement and returns the raw count, or 0 if the pin is not
// a touch pad or the sensor was never set up.
func (t Pad) Get() uint16 {
	pad, ok := padNumber(t.Pin)
	if !ok || !initialized {
		return 0
	}

	esp.SENS.SetSAR_TOUCH_ENABLE_TOUCH_PAD_WORKEN(1 << pad)

	esp.SENS.SetSAR_TOUCH_CTRL2_TOUCH_START_EN(0)
	esp.SENS.SetSAR_TOUCH_CTRL2_TOUCH_START_EN(1)

	for esp.SENS.GetSAR_TOUCH_CTRL2_TOUCH_MEAS_DONE() == 0 {
	}

	// Pad n is in SAR_TOUCH_OUT(n/2+1), even pads in the upper half
	// (ESP32 TRM v4.2 section 29.7.1, page 608).
	out := (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.SENS.SAR_TOUCH_OUT1), uintptr(pad/2)*4))
	if pad%2 == 0 {
		return uint16(out.Get() >> 16)
	}
	return uint16(out.Get())
}

func padReg(pad uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.RTC_IO.TOUCH_PAD0), uintptr(pad)*4))
}

// padNumber maps a pin to its touch pad (ESP32 TRM v4.2 table 122,
// page 589).
func padNumber(p machine.Pin) (uint32, bool) {
	switch p {
	case machine.GPIO4:
		return 0, true
	case machine.GPIO0:
		return 1, true
	case machine.GPIO2:
		return 2, true
	case machine.GPIO15:
		return 3, true
	case machine.GPIO13:
		return 4, true
	case machine.GPIO12:
		return 5, true
	case machine.GPIO14:
		return 6, true
	case machine.GPIO27:
		return 7, true
	}
	return 0, false
}
