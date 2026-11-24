//go:build esp32c3 || esp32c6 || esp32h2 || esp32p4 || (esp32s3 && !usbotg)

package machine

import (
	"device/esp"
	"errors"
	"machine/usb"
	"machine/usb/descriptor"
)

type USB_DEVICE struct {
	Bus       *esp.USB_DEVICE_Type
	Buffer    *RingBuffer
	txPending bool // unflushed data in the EP1 TX FIFO
	txStalled bool // set when flushAndWait fails (no host reading); cleared when FIFO becomes writable
}

var (
	_USBCDC = &USB_DEVICE{
		Bus:    esp.USB_DEVICE,
		Buffer: NewRingBuffer(),
	}

	USBCDC Serialer = _USBCDC
)

var (
	errUSBWrongSize            = errors.New("USB: invalid write size")
	errUSBCouldNotWriteAllData = errors.New("USB: could not write all data")
)

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Configure(config UARTConfig) error
	Buffered() int
	ReadByte() (byte, error)
	DTR() bool
	RTS() bool
}

// USBDevice is a stub USB device. The USB Serial/JTAG controller only has a
// fixed-function CDC-ACM serial port, so the programmable USB device features are no-ops.
type USBDevice struct {
	initcomplete         bool
	InitEndpointComplete bool
}

var USBDev = &USBDevice{}

func (dev *USBDevice) SetStallEPIn(ep uint32)    {}
func (dev *USBDevice) SetStallEPOut(ep uint32)   {}
func (dev *USBDevice) ClearStallEPIn(ep uint32)  {}
func (dev *USBDevice) ClearStallEPOut(ep uint32) {}

// Attach and Detach are no-ops: the USB Serial/JTAG controller has no
// software-controlled soft-connect, it is always attached to the bus.
func (dev *USBDevice) Attach() {}
func (dev *USBDevice) Detach() {}

// initUSB is empty because the interp pass runs init at compile time and
// cannot access the hardware. Configure does the setup on first use.
func initUSB() {}

var usbConfigured bool

// Configure configures the USB Serial/JTAG controller. Only the first call
// has an effect.
func (usbdev *USB_DEVICE) Configure(config UARTConfig) error {
	if usbConfigured {
		return nil
	}
	usbConfigured = true
	return usbdev.configure(config)
}

// ensureConfigured triggers lazy initialization on first use.
func (usbdev *USB_DEVICE) ensureConfigured() {
	if !usbConfigured {
		usbdev.Configure(UARTConfig{})
	}
}

// ReadByte returns a byte from the receive ring buffer, or errNoByte when it
// is empty.
func (usbdev *USB_DEVICE) ReadByte() (byte, error) {
	usbdev.ensureConfigured()
	b, ok := usbdev.Buffer.Get()
	if !ok {
		return 0, errNoByte
	}
	return b, nil
}

func (usbdev *USB_DEVICE) DTR() bool {
	return false
}

func (usbdev *USB_DEVICE) RTS() bool {
	return false
}

// ConfigureUSBEndpoint, SendZlp and SendUSBInPacket are no-ops. The USB
// Serial/JTAG controller does not support programmable USB endpoints.
func ConfigureUSBEndpoint(desc descriptor.Descriptor, epSettings []usb.EndpointConfig, setup []usb.SetupConfig) {
}

func SendZlp() {
}

func SendUSBInPacket(ep uint32, data []byte) bool {
	return false
}
