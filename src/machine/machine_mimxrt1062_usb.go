// +build mimxrt1062

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
	uart usb2.UART
}

// Configure the embedded usb.UART with our receiver's settings and the given
// UART configuration. This provides compatibility with machine.UART.
func (cdc *USBCDC) Configure(config UARTConfig) {
	cdc.uart.Configure(usb2.UARTConfig{BaudRate: config.BaudRate})
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (cdc USBCDC) Buffered() int {
	return cdc.uart.Buffered()
}

// ReadByte reads a single byte from the RX buffer.
// If there is no data in the buffer, returns an error.
func (cdc USBCDC) ReadByte() (byte, error) {
	return cdc.uart.ReadByte()
}

// Read from the RX buffer.
func (cdc USBCDC) Read(data []byte) (n int, err error) {
	return cdc.uart.Read(data)
}

// WriteByte writes a single byte of data to the UART interface.
func (cdc USBCDC) WriteByte(c byte) error {
	return cdc.uart.WriteByte(c)
}

// Write data to the UART.
func (cdc USBCDC) Write(data []byte) (n int, err error) {
	return cdc.uart.Write(data)
}
