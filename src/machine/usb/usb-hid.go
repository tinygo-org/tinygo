//go:build baremetal && usb.hid
// +build baremetal,usb.hid

package usb

import (
	"errors"
)

var (
	ErrHIDInvalidPort    = errors.New("invalid USB port")
	ErrHIDInvalidCore    = errors.New("invalid USB core")
	ErrHIDReportTransfer = errors.New("failed to transfer HID report")
)

// HID represents a virtual keyboard/mouse/joystick device (with a serial data
// Rx/Tx interface) using the USB HID device class driver.
type HID struct {
	// Port is the MCU's native USB core number. If in doubt, leave it
	// uninitialized for default (0).
	Port int
	core *core
}

type HIDConfig struct {
	BusSpeed Speed
}

func (hid *HID) Configure(config HIDConfig) error {

	c := class{id: classDeviceHID, config: 1}

	// verify we have a free USB port and take ownership of it
	var st status
	hid.core, st = initCore(hid.Port, config.BusSpeed, c)
	if !st.ok() {
		return ErrHIDInvalidPort
	}
	return nil
}

func (hid *HID) Ready() bool {
	return hid.core.dc.keyboard().ready()
}

func (hid *HID) Keyboard() *Keyboard {
	return hid.core.dc.keyboard()
}
