// +build mimxrt1062

// Compatibility wrapper for legacy type USBCDC, which provides USB CDC-ACM
// device class emulation for serial UART communication.
// This functionality is being replaced by a platform-agnostic type usb.UART in
// package "machine/usb".

package machine

import (
	"machine/usb2"
)

// USBCDC is the legacy TinyGo type used to implement USB CDC-ACM device class
// emulation for serial UART communication. It is retained here as a temporary
// wrapper for type usb.UART from new package "machine/usb", which is still in
// active development. Once that package has stabilized a bit, this type should
// be removed and usb.UART should be used directly instead.
type USBCDC struct {
	port uint8
	buff *RingBuffer
	uart usb2.UART
}

// Configure the embedded usb.UART with our receiver's settings and the given
// UART configuration. This provides compatibility with machine.UART.
func (cdc *USBCDC) Configure(config UARTConfig) {
	cdc.uart.Configure(usb2.UARTConfig{BaudRate: config.BaudRate})
}
