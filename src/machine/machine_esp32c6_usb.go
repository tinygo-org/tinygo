//go:build esp32c6 || esp32h2

package machine

import (
	"device/esp"
	"runtime/interrupt"
)

// USB Serial/JTAG Controller
// The ESP32-C6 and ESP32-H2 have the same built-in USB Serial/JTAG hardware IP
// as the ESP32-C3. The only difference at the machine level is how the
// peripheral clock is enabled: C3 uses SYSTEM.PERIP_CLK_EN0, while C6 and H2
// use the PCR (Peripheral Clock Reset) block.

const cpuInterruptFromUSB = 10

// Masks for the RX interrupt and the EP1 OUT FIFO status.
const (
	usbEP1OutDataAvail = esp.USB_DEVICE_EP1_CONF_SERIAL_OUT_EP_DATA_AVAIL
	usbRecvPktIntEna   = esp.USB_DEVICE_INT_ENA_SERIAL_OUT_RECV_PKT_INT_ENA
	usbRecvPktIntClr   = esp.USB_DEVICE_INT_CLR_SERIAL_OUT_RECV_PKT_INT_CLR
)

// configure initialises the USB Serial/JTAG controller clock, pads, and
// interrupt so that received data is buffered automatically.
func (usbdev *USB_DEVICE) configure(config UARTConfig) error {
	// Enable the USB_DEVICE peripheral clock via PCR (C6 and H2 use PCR
	// instead of SYSTEM.PERIP_CLK_EN0 that C3 used).
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
