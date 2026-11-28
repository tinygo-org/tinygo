//go:build esp32 || esp32s3

// Package touch reads capacitive touch sensors on the ESP32 and ESP32-S3.
package touch

import (
	"device/esp"
	"errors"
	"machine"
	"runtime/volatile"
	"unsafe"
)

// Pad reads the capacitive touch sensor on one pin. The count drops when the
// pad is touched on the ESP32 (ESP32 TRM v4.2 section 29.2.4, page 589) and
// rises on the ESP32-S3.
type Pad struct {
	Pin machine.Pin
}

var ErrInvalidPin = errors.New("touch: invalid touch pin")

var initialized bool

// Configure hands the pin to the touch sensor. The ESP32 supports touch pads
// T0 to T7, the ESP32-S3 supports T1 to T14 on GPIO1 to GPIO14.
func (t Pad) Configure() error {
	pad, ok := padNumber(t.Pin)
	if !ok {
		return ErrInvalidPin
	}

	if !initialized {
		initialized = true
		initSensor()
	}

	configurePad(t.Pin, pad)
	return nil
}

// Get returns the raw count, or 0 if the pin is not a touch pad, the sensor
// was never set up, or the first ESP32-S3 scan has not finished.
func (t Pad) Get() uint32 {
	pad, ok := padNumber(t.Pin)
	if !ok || !initialized {
		return 0
	}
	return readPad(pad)
}

func padReg(pad uint32) *volatile.Register32 {
	return (*volatile.Register32)(unsafe.Add(unsafe.Pointer(&esp.RTC_IO.TOUCH_PAD0), uintptr(pad)*4))
}
