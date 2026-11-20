//go:build esp32p4

package machine

import (
	"device/esp"
	"errors"
	"machine/usb"
	"machine/usb/descriptor"
	"runtime/interrupt"
)

// USB Serial/JTAG controller. The ROM has already set up the clocks and pads.

const cpuInterruptFromUSB = 10

const (
	usbEP1WrDone      = esp.USB_DEVICE_EP1_CONF_USB_SERIAL_JTAG_WR_DONE
	usbEP1InDataFree  = esp.USB_DEVICE_EP1_CONF_USB_SERIAL_JTAG_SERIAL_IN_EP_DATA_FREE
	usbEP1OutDataAvai = esp.USB_DEVICE_EP1_CONF_USB_SERIAL_JTAG_SERIAL_OUT_EP_DATA_AVAIL
	usbRecvPktIntEna  = esp.USB_DEVICE_INT_ENA_USB_SERIAL_JTAG_SERIAL_OUT_RECV_PKT_INT_ENA
	usbRecvPktIntClr  = esp.USB_DEVICE_INT_CLR_USB_SERIAL_JTAG_SERIAL_OUT_RECV_PKT_INT_CLR
)

type USB_DEVICE struct {
	Bus       *esp.USB_DEVICE_Type
	Buffer    *RingBuffer
	txPending bool // unflushed data in the EP1 TX FIFO
	txStalled bool // no host is reading, cleared when the FIFO becomes writable
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

// USBDevice is a stub. The hardware only has a fixed CDC-ACM serial port.
type USBDevice struct {
	initcomplete         bool
	InitEndpointComplete bool
}

var USBDev = &USBDevice{}

func (dev *USBDevice) SetStallEPIn(ep uint32)    {}
func (dev *USBDevice) SetStallEPOut(ep uint32)   {}
func (dev *USBDevice) ClearStallEPIn(ep uint32)  {}
func (dev *USBDevice) ClearStallEPOut(ep uint32) {}
func (dev *USBDevice) Attach()                   {}
func (dev *USBDevice) Detach()                   {}

func initUSB() {}

var usbConfigured bool

// Configure enables the receive interrupt. It runs on first use, because
// initUSB runs at compile time in the interp pass.
func (usbdev *USB_DEVICE) Configure(config UARTConfig) error {
	if usbConfigured {
		return nil
	}
	usbConfigured = true

	usbdev.Bus.INT_CLR.Set(0xffffffff)
	usbdev.Bus.INT_ENA.SetBits(usbRecvPktIntEna)

	// The interrupt matrix takes the CLIC number, which is the line plus 16.
	esp.INTERRUPT_CORE0.USB_DEVICE_INT_MAP.Set(cpuInterruptFromUSB + 16)
	interrupt.New(cpuInterruptFromUSB, func(interrupt.Interrupt) {
		_USBCDC.handleInterrupt()
	}).Enable()
	return nil
}

// handleInterrupt moves received bytes from the FIFO into the ring buffer.
func (usbdev *USB_DEVICE) handleInterrupt() {
	for usbdev.Bus.EP1_CONF.HasBits(usbEP1OutDataAvai) {
		if !usbdev.Buffer.Put(byte(usbdev.Bus.EP1.Get())) {
			break
		}
	}
	usbdev.Bus.INT_CLR.Set(usbRecvPktIntClr)
}

func (usbdev *USB_DEVICE) WriteByte(c byte) error {
	if !usbdev.waitWritable() {
		return errUSBCouldNotWriteAllData
	}
	usbdev.Bus.EP1.Set(uint32(c))
	if c == '\n' {
		usbdev.flush()
	} else {
		usbdev.txPending = true
	}
	return nil
}

func (usbdev *USB_DEVICE) Write(data []byte) (n int, err error) {
	for i, c := range data {
		if !usbdev.waitWritable() {
			return i, errUSBCouldNotWriteAllData
		}
		usbdev.Bus.EP1.Set(uint32(c))
	}
	if len(data) > 0 {
		usbdev.flush()
	}
	return len(data), nil
}

// waitWritable returns false if the EP1 FIFO stays full because no host reads it.
func (usbdev *USB_DEVICE) waitWritable() bool {
	if usbdev.Bus.EP1_CONF.HasBits(usbEP1InDataFree) {
		usbdev.txStalled = false
		return true
	}
	if usbdev.txStalled {
		return false
	}
	usbdev.Bus.EP1_CONF.Set(usbEP1WrDone)
	for i := 0; i < 50000; i++ {
		if usbdev.Bus.EP1_CONF.HasBits(usbEP1InDataFree) {
			return true
		}
	}
	usbdev.txStalled = true
	return false
}

// Buffered returns the number of received bytes in the ring buffer.
func (usbdev *USB_DEVICE) Buffered() int {
	usbdev.Configure(UARTConfig{})
	if usbdev.txPending {
		usbdev.flush()
	}
	return int(usbdev.Buffer.Used())
}

func (usbdev *USB_DEVICE) ReadByte() (byte, error) {
	usbdev.Configure(UARTConfig{})
	b, ok := usbdev.Buffer.Get()
	if !ok {
		return 0, nil
	}
	return b, nil
}

func (usbdev *USB_DEVICE) DTR() bool {
	return false
}

func (usbdev *USB_DEVICE) RTS() bool {
	return false
}

// flush sends the data in the EP1 FIFO to the host.
func (usbdev *USB_DEVICE) flush() {
	usbdev.Bus.EP1_CONF.Set(usbEP1WrDone)
	usbdev.txPending = false
}

func ConfigureUSBEndpoint(desc descriptor.Descriptor, epSettings []usb.EndpointConfig, setup []usb.SetupConfig) {
}

func SendZlp() {
}

func SendUSBInPacket(ep uint32, data []byte) bool {
	return false
}
