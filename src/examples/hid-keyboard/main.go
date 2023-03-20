// to override the USB Manufacturer or Product names:
//
// tinygo flash -target circuitplay-express -ldflags="-X main.usbManufacturer='TinyGopher Labs' -X main.usbProduct='GopherKeyboard'" examples/hid-keyboard
//
// you can also override the VID/PID. however, only set this if you know what you are doing,
// since changing it can make it difficult to reflash some devices.
package main

import (
	"machine"
	"machine/usb"
	"machine/usb/hid/keyboard"
	"strconv"
	"time"
)

var usbVID, usbPID string
var usbManufacturer, usbProduct string

func main() {
	button := machine.BUTTON
	button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	kb := keyboard.Port()

	for {
		if !button.Get() {
			kb.Write([]byte("tinygo"))
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func init() {
	if usbVID != "" {
		vid, _ := strconv.ParseUint(usbVID, 0, 16)
		usb.VendorID = uint16(vid)
	}

	if usbPID != "" {
		pid, _ := strconv.ParseUint(usbPID, 0, 16)
		usb.ProductID = uint16(pid)
	}

	if usbManufacturer != "" {
		usb.Manufacturer = usbManufacturer
	}

	if usbProduct != "" {
		usb.Product = usbProduct
	}
}
