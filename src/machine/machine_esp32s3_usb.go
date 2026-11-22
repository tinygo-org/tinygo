//go:build esp32s3 && !usbotg

package machine

import (
	"device/esp"
	"runtime/interrupt"
)

// USB Serial/JTAG Controller
// See esp32-s3_technical_reference_manual_en.pdf
//
// The ESP32-S3 has a built-in USB Serial/JTAG controller that provides a
// CDC-ACM serial port. The USB protocol and enumeration are handled entirely
// in hardware; software only reads/writes the EP1 FIFO.

const cpuInterruptFromUSB = 8

// usbHandleInterrupt is the top-level interrupt handler passed to
// interrupt.New. It must be a plain function (not a closure) because
// interrupt.New is a compiler intrinsic that does not support closures.
func usbHandleInterrupt(interrupt.Interrupt) {
	_USBCDC.handleInterrupt()
}

// configure initialises the USB Serial/JTAG controller.
func (usbdev *USB_DEVICE) configure(config UARTConfig) error {
	// Enable the USB_DEVICE peripheral clock.
	esp.SYSTEM.SetPERIP_CLK_EN1_USB_DEVICE_CLK_EN(1)
	esp.SYSTEM.SetPERIP_RST_EN1_USB_DEVICE_RST(0)

	// Clear any pending interrupts, then enable the RX interrupt.
	usbdev.Bus.INT_CLR.Set(0xFFFFFFFF)
	usbdev.Bus.SetINT_ENA_SERIAL_OUT_RECV_PKT_INT_ENA(1)

	// Map the USB_DEVICE peripheral interrupt to CPU interrupt line.
	esp.INTERRUPT_CORE0.SetUSB_DEVICE_INT_MAP(cpuInterruptFromUSB)

	_ = interrupt.New(cpuInterruptFromUSB, usbHandleInterrupt).Enable()

	return nil
}

// handleInterrupt is called from the CPU interrupt vector when the USB
// peripheral raises an interrupt.  Disable INT_ENA to prevent the
// level-triggered interrupt from re-asserting immediately (data may
// still be in the FIFO).  Buffered() re-enables after draining.
func (usbdev *USB_DEVICE) handleInterrupt() {
	usbdev.Bus.SetINT_ENA_SERIAL_OUT_RECV_PKT_INT_ENA(0)
	usbdev.Bus.SetINT_CLR_SERIAL_OUT_RECV_PKT_INT_CLR(1)
}

func (usbdev *USB_DEVICE) WriteByte(c byte) error {
	usbdev.ensureConfigured()
	if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() == 0 {
		// FIFO locked by a pending USB transfer.
		if usbdev.txStalled {
			// Previously failed — skip the expensive spin and drop
			// the byte. When a host reconnects SERIAL_IN_EP_DATA_FREE
			// goes back to 1, clearing the stall on the next call.
			return errUSBCouldNotWriteAllData
		}
		// First time the FIFO is full: wait briefly for the host to
		// read the previous packet.
		if !usbdev.flushAndWait() {
			usbdev.txStalled = true
			return errUSBCouldNotWriteAllData
		}
	}
	usbdev.txStalled = false

	// Use EP1.Set() (direct store) instead of SetEP1_RDWR_BYTE which
	// does a read-modify-write — the read side-effect pops a byte from
	// the RX FIFO.
	usbdev.Bus.EP1.Set(uint32(c))

	// Only signal WR_DONE on newline to batch bytes into a single USB
	// packet. The FIFO-full path above also flushes when the 64-byte
	// FIFO fills up.
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
// It drains any data sitting in the hardware FIFO and re-enables the
// peripheral-level USB interrupt (which the ISR disables to prevent a
// level-triggered interrupt storm).
func (usbdev *USB_DEVICE) Buffered() int {
	usbdev.ensureConfigured()
	// Flush any pending TX data so callers like echo loops don't
	// need to explicitly flush after WriteByte.
	if usbdev.txPending {
		usbdev.flush()
		usbdev.txPending = false
	}
	// Drain the hardware FIFO into the ring buffer.
	for usbdev.Bus.GetEP1_CONF_SERIAL_OUT_EP_DATA_AVAIL() != 0 {
		b := byte(usbdev.Bus.EP1.Get())
		usbdev.Buffer.Put(b)
	}
	// Clear pending flags and re-enable the RX interrupt at the peripheral level.
	usbdev.Bus.INT_CLR.Set(0xFFFFFFFF)
	usbdev.Bus.SetINT_ENA_SERIAL_OUT_RECV_PKT_INT_ENA(1)
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

// flush signals WR_DONE to tell the hardware to send the data that has
// been written to the EP1 FIFO. Returns immediately without waiting.
func (usbdev *USB_DEVICE) flush() {
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
}

// FlushSerial flushes any pending USB serial TX data. Called from the
// runtime (e.g. before sleeping) to ensure data from print() without
// a trailing newline gets sent promptly.
func FlushSerial() {
	if _USBCDC.txPending {
		_USBCDC.flush()
		_USBCDC.txPending = false
	}
}

// flushAndWait signals WR_DONE and waits for the EP1 FIFO to become
// writable again. The timeout covers a few USB frames so that data gets
// through when a host is connected. Returns false if the FIFO is still
// locked after the timeout (no host reading).
func (usbdev *USB_DEVICE) flushAndWait() bool {
	usbdev.Bus.SetEP1_CONF_WR_DONE(1)
	for i := 0; i < 50000; i++ {
		if usbdev.Bus.GetEP1_CONF_SERIAL_IN_EP_DATA_FREE() != 0 {
			return true
		}
	}
	return false
}
