//go:build usb.cdc
// +build usb.cdc

package usb

import (
	"errors"
)

var (
	ErrCDCInvalidPort = errors.New("invalid port")
	ErrCDCEmptyBuffer = errors.New("buffer empty")
)

// CDC represents a virtual UART serial device emulation using the USB
// CDC-ACM device class driver.
type CDC struct {
	// Port is the MCU's native USB core number. If in doubt, leave it
	// uninitialized for default (0).
	Port int
	core *core
}

type CDCConfig struct {
	BusSpeed Speed
}

func (cdc *CDC) Configure(config CDCConfig) error {

	c := class{id: classDeviceCDC, config: 1}

	// verify we have a free USB port and take ownership of it
	var st status
	cdc.core, st = initCore(cdc.Port, config.BusSpeed, c)
	if !st.ok() {
		return ErrCDCInvalidPort
	}
	return nil
}

func (cdc *CDC) Ready() bool {
	return cdc.core.dc.cdcReady()
}

// Buffered returns the number of bytes currently stored in the Rx buffer.
func (cdc *CDC) Buffered() int {
	for !cdc.Ready() {
	}
	return cdc.core.dc.cdcAvailable()
}

// ReadByte reads a single byte from the Rx buffer.
// If there is no data in the buffer, returns an error.
func (cdc *CDC) ReadByte() (byte, error) {
	for !cdc.Ready() {
	}
	n, ok := cdc.core.dc.cdcReadByte()
	if !ok {
		return 0, ErrCDCEmptyBuffer
	}
	return n, nil
}

// Read from the Rx buffer.
func (cdc *CDC) Read(data []byte) (n int, err error) {
	for !cdc.Ready() {
	}
	return cdc.core.dc.cdcRead(data)
}

// WriteByte writes a single byte of data to the virtual UART interface.
func (cdc *CDC) WriteByte(c byte) error {
	for !cdc.Ready() {
	}
	return cdc.core.dc.cdcWriteByte(c)
}

// Write data to the virtual UART.
func (cdc *CDC) Write(data []byte) (n int, err error) {
	for !cdc.Ready() {
	}
	return cdc.core.dc.cdcWrite(data)
}
