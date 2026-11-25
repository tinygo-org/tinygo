//go:build arduino_uno_r4_wifi && serial.usb

package machine

// USB identifiers
const (
	usb_STRING_PRODUCT      = "UNO R4 WiFi"
	usb_STRING_MANUFACTURER = "Arduino"
)

var (
	usb_VID uint16 = 0x2341
	usb_PID uint16 = 0x1002
)

// P4_08 routes the USB connector to the RA4M1 when high, otherwise to the ESP32-S3.
// See configure_usb_mux in ArduinoCore-renesas variants/UNOWIFIR4/variant.cpp.
func init() {
	P4_08.Configure(PinConfig{Mode: PinOutput})
	P4_08.High()
}
