//go:build esp32c6

package machine

import (
	"device/esp"
	"errors"
	"machine/usb"
	"machine/usb/descriptor"
	"runtime/interrupt"
)

// USB Serial/JTAG Controller
// The ESP32-C6 has the same built-in USB Serial/JTAG hardware IP as the
// ESP32-C3. The only difference at the machine level is how the peripheral
// clock is enabled: C3 uses SYSTEM.PERIP_CLK_EN0, while C6 uses the PCR
// (Peripheral Clock Reset) block.

const cpuInterruptFromUSB = 10

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

var usbConfigured bool

// USBDevice provides a stub USB device for the ESP32-C6. The hardware
// only supports a fixed-function CDC-ACM serial port, so the programmable
// USB device features are no-ops.
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

// initUSB is intentionally empty — the interp phase evaluates init()
// functions at compile time and cannot access hardware registers.
// Actual hardware setup is deferred to the first Configure() call.
func initUSB() {}

// Configure initialises the USB Serial/JTAG controller clock, pads, and
// interrupt so that received data is buffered automatically.
func (usbdev *USB_DEVICE) Configure(config UARTConfig) error {
	if usbConfigured {
		return nil
	}
	usbConfigured = true

	// Enable the USB_DEVICE peripheral clock via PCR (C6 uses PCR instead
	// of SYSTEM.PERIP_CLK_EN0 that C3 used).
	esp.PCR.SetUSB_DEVICE_CONF_USB_DEVICE_CLK_EN(1)
	esp.PCR.SetUSB_DEVICE_CONF_USB_DEVICE_RST_EN(0)

	// Ensure internal PHY is selected and USB pads are enabled.
	usbdev.Bus.SetCONF0_PHY_SEL(0)
	usbdev.Bus.SetCONF0_USB_PAD_ENABLE(1)
	usbdev.Bus.SetCONF0_DP_PULLUP(1)

	// Clear any pending interrupts.
	usbdev.Bus.INT_CLR.Set(0xFFFFFFFF)

	// Enable the RX-packet-received interrupt.
	usbdev.Bus.SetINT_ENA_SERIAL_OUT_RECV_PKT_INT_ENA(1)

	// Map the USB peripheral interrupt to CPU interrupt cpuInterruptFromUSB.
	esp.INTERRUPT_CORE0.SetUSB_INTR_MAP(cpuInterruptFromUSB)

	_ = interrupt.New(cpuInterruptFromUSB, func(interrupt.Interrupt) {
		_USBCDC.handleInterrupt()
	}).Enable()

	return nil
}

// ensureConfigured triggers lazy initialization on first use.
func (usbdev *USB_DEVICE) ensureConfigured() {
	if !usbConfigured {
		usbdev.Configure(UARTConfig{})
	}
}

// handleInterrupt drains the hardware RX FIFO into the software ring buffer.
func (usbdev *USB_DEVICE) handleInterrupt() {
	intStatus := usbdev.Bus.INT_ST.Get()

	// Disable the RX interrupt to prevent re-triggering while we drain.
	usbdev.Bus.SetINT_ENA_SERIAL_OUT_RECV_PKT_INT_ENA(0)

	if intStatus&esp.USB_DEVICE_INT_ST_SERIAL_OUT_RECV_PKT_INT_ST != 0 {
		for usbdev.Bus.GetEP1_CONF_SERIAL_OUT_EP_DATA_AVAIL() != 0 {
			b := byte(usbdev.Bus.EP1.Get())
			usbdev.Buffer.Put(b)
		}
		usbdev.Bus.SetINT_CLR_SERIAL_OUT_RECV_PKT_INT_CLR(1)
	}

	// Re-enable the RX interrupt.
	usbdev.Bus.SetINT_ENA_SERIAL_OUT_RECV_PKT_INT_ENA(1)
}

func (usbdev *USB_DEVICE) WriteByte(c byte) error {
	usbdev.ensureConfigured()
	if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
		if usbdev.txStalled {
			return errUSBCouldNotWriteAllData
		}
		if !usbdev.flushAndWait() {
			usbdev.txStalled = true
			return errUSBCouldNotWriteAllData
		}
	}
	usbdev.txStalled = false

	usbdev.Bus.EP1.Set(uint32(c))

	if c == '\n' {
		usbdev.flush()
		usbdev.txPending = false
	} else {
		usbdev.txPending = true
	}

	return nil
}

func (usbdev *USB_DEVICE) Write(data []byte) (n int, err error) {
	usbdev.ensureConfigured()
	if len(data) == 0 {
		return 0, nil
	}

	for i, c := range data {
		if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
			if usbdev.txStalled {
				return i, errUSBCouldNotWriteAllData
			}
			if !usbdev.flushAndWait() {
				usbdev.txStalled = true
				return i, errUSBCouldNotWriteAllData
			}
		}
		usbdev.txStalled = false
		usbdev.Bus.EP1.Set(uint32(c))
	}

	usbdev.flush()
	usbdev.txPending = false
	return len(data), nil
}

// Buffered returns the number of bytes waiting in the receive ring buffer.
func (usbdev *USB_DEVICE) Buffered() int {
	usbdev.ensureConfigured()
	if usbdev.txPending {
		usbdev.flush()
		usbdev.txPending = false
	}
	return int(usbdev.Buffer.Used())
}

// ReadByte returns a byte from the receive ring buffer.
func (usbdev *USB_DEVICE) ReadByte() (byte, error) {
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

// flush signals WR_DONE to tell the hardware to send the data that has
// been written to the EP1 FIFO.
func (usbdev *USB_DEVICE) flush() {
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
}

// flushAndWait signals WR_DONE and waits for the EP1 FIFO to become
// writable again. Returns false if the FIFO is still locked after the
// timeout (no host reading).
func (usbdev *USB_DEVICE) flushAndWait() bool {
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
	for i := 0; i < 50000; i++ {
		if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() != 0 {
			return true
		}
	}
	return false
}

// FlushSerial flushes any pending USB serial TX data.
func FlushSerial() {
	if _USBCDC.txPending {
		_USBCDC.flush()
		_USBCDC.txPending = false
	}
}

// ConfigureUSBEndpoint is a no-op on ESP32-C6 — the hardware does not
// support programmable USB endpoints.
func ConfigureUSBEndpoint(desc descriptor.Descriptor, epSettings []usb.EndpointConfig, setup []usb.SetupConfig) {
}

// SendZlp is a no-op on ESP32-C6.
func SendZlp() {
}

// SendUSBInPacket is a no-op on ESP32-C6.
func SendUSBInPacket(ep uint32, data []byte) bool {
	return false
}
