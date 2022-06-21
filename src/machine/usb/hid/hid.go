package hid

import (
	"errors"
	"machine"
)

// from usb-hid.go
var (
	ErrHIDInvalidPort    = errors.New("invalid USB port")
	ErrHIDInvalidCore    = errors.New("invalid USB core")
	ErrHIDReportTransfer = errors.New("failed to transfer HID report")
)

const (
	hidEndpoint = 4

	usb_SET_REPORT_TYPE = 33
	usb_SET_IDLE        = 10
)

type hidDevicer interface {
	Callback() bool
}

var devices [5]hidDevicer
var size int

// SetCallbackHandler sets the callback. Only the first time it is called, it
// calls machine.EnableHID for USB configuration
func SetCallbackHandler(d hidDevicer) {
	if size == 0 {
		machine.EnableHID(callback, nil, nil)
	}

	devices[size] = d
	size++
}

func callback() {
	for _, d := range devices {
		if d == nil {
			continue
		}
		if done := d.Callback(); done {
			return
		}
	}
}

func callbackSetup(setup machine.USBSetup) bool {
	ok := false
	if setup.BmRequestType == usb_SET_REPORT_TYPE && setup.BRequest == usb_SET_IDLE {
		machine.SendZlp()
		ok = true
	}
	return ok
}

// SendUSBPacket sends a HIDPacket.
func SendUSBPacket(b []byte) {
	machine.SendUSBInPacket(hidEndpoint, b)
}
