//go:build baremetal && serial.usb
// +build baremetal,serial.usb

package machine

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Configure(config UARTConfig)
	Configured() bool
	Buffered() int
	ReadByte() (byte, error)
}

//go:linkname NewUSBCDC machine/usb/cdc.New
func NewUSBCDC() Serialer

// Serial is implemented via USB (USB-CDC).
var Serial = NewUSBCDC()
