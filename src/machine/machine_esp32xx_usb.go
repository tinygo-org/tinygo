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

// Buffered returns the number of bytes in the receive ring buffer. It also sends
// the bytes that wait in the EP1 IN FIFO, so that an echo loop does not need to flush.
func (usbdev *USB_DEVICE) Buffered() int {
	usbdev.ensureConfigured()
	if usbdev.txPending {
		usbdev.flush()
		usbdev.txPending = false
	}
	usbdev.resumeRX()
	return int(usbdev.Buffer.Used())
}

// ReadByte returns a byte from the receive ring buffer, or errNoByte when it
// is empty.
func (usbdev *USB_DEVICE) ReadByte() (byte, error) {
	usbdev.ensureConfigured()
	b, ok := usbdev.Buffer.Get()
	usbdev.resumeRX()
	if !ok {
		return 0, errNoByte
	}
	return b, nil
}

// handleInterrupt moves received bytes from the EP1 OUT FIFO to the ring buffer.
// When the ring buffer is full, it leaves the rest in the FIFO, so the controller
// stops accepting data from the host, and pauses the RX interrupt until
// resumeRX. The status is cleared only when the FIFO is empty, and the FIFO is
// checked again after that, so a packet that arrives during the clear is not lost.
func (usbdev *USB_DEVICE) handleInterrupt() {
	for {
		for usbdev.Bus.EP1_CONF.HasBits(usbEP1OutDataAvail) {
			if usbdev.Buffer.Used() == bufferSize {
				usbdev.Bus.INT_ENA.ClearBits(usbRecvPktIntEna)
				return
			}
			usbdev.Buffer.Put(byte(usbdev.Bus.EP1.Get()))
		}
		usbdev.Bus.INT_CLR.Set(usbRecvPktIntClr)
		if !usbdev.Bus.EP1_CONF.HasBits(usbEP1OutDataAvail) {
			return
		}
	}
}

// resumeRX enables the RX interrupt again after handleInterrupt paused it and
// there is room in the ring buffer. The status is still set, so the interrupt
// then drains the FIFO. Only handleInterrupt writes to the ring buffer.
func (usbdev *USB_DEVICE) resumeRX() {
	if !usbdev.Bus.INT_ENA.HasBits(usbRecvPktIntEna) && usbdev.Buffer.Used() < bufferSize {
		usbdev.Bus.INT_ENA.SetBits(usbRecvPktIntEna)
	}
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
