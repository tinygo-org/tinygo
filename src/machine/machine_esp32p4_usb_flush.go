//go:build esp32p4 && serial.usb

package machine

// FlushSerial sends any pending USB serial data.
func FlushSerial() {
	if _USBCDC.txPending {
		_USBCDC.flush()
	}
}
