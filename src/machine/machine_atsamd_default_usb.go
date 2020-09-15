// +build atsamd21,!no_default_usb atsamd51,!no_default_usb

package machine

var (
	// UART0 is actually a USB CDC interface.
	UART0 = USBCDC{Buffer: NewRingBuffer()}
)
