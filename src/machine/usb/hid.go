package usb

import (
	"errors"
)

var (
	ErrHIDInvalidPort    = errors.New("invalid USB port")
	ErrHIDInvalidCore    = errors.New("invalid USB core")
	ErrHIDReportTransfer = errors.New("failed to transfer HID report")
)

type HID struct {
	Port int
	core *core
}

type HIDConfig struct {
	// Port is the MCU's native USB core number. If in doubt, leave it
	// uninitialized for default (0).
	Port int
}

func (hid *HID) Configure(config HIDConfig) error {

	if config.Port >= CoreCount || config.Port >= dcdCount {
		return ErrHIDInvalidPort
	}
	hid.Port = config.Port

	// verify we have a free USB port and take ownership of it
	var st status
	hid.core, st = initCore(hid.Port, class{id: classDeviceHID, config: 1})
	if !st.ok() {
		return ErrHIDInvalidPort
	}
	return nil
}

func (hid *HID) Keyboard() *Keyboard {
	return hid.core.dc.keyboard()
}
