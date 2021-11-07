package usb

import (
	"errors"
)

var (
	ErrUARTInvalidPort = errors.New("invalid USB port")
	ErrUARTInvalidCore = errors.New("invalid USB core")
	ErrUARTEmptyBuffer = errors.New("USB receive buffer empty")
	ErrUARTWriteFailed = errors.New("USB write failure")
)

// UART represents a virtual serial (UART) device emulation using the USB
// CDC-ACM device class driver.
type UART struct {
	Port int
	core *core
}

type UARTConfig struct {
	// Port is the MCU's native USB core number. If in doubt, leave it
	// uninitialized for default (0).
	Port int
}

func (uart *UART) Configure(config UARTConfig) error {

	if config.Port >= CoreCount || config.Port >= dcdCount {
		return ErrUARTInvalidPort
	}
	uart.Port = config.Port

	// verify we have a free USB port and take ownership of it
	var st status
	uart.core, st = initCore(uart.Port, class{id: classDeviceCDCACM, config: 1})
	if !st.ok() {
		return ErrUARTInvalidPort
	}
	return nil
}

// Buffered returns the number of bytes currently stored in the RX buffer.
func (uart UART) Buffered() int {
	return uart.core.dc.uartAvailable()
}

// ReadByte reads a single byte from the RX buffer.
// If there is no data in the buffer, returns an error.
func (uart UART) ReadByte() (byte, error) {
	n, ok := uart.core.dc.uartReadByte()
	if !ok {
		return 0, ErrUARTEmptyBuffer
	}
	return n, nil
}

// Read from the RX buffer.
func (uart UART) Read(data []byte) (n int, err error) {
	return uart.core.dc.uartRead(data), nil
}

// WriteByte writes a single byte of data to the UART interface.
func (uart UART) WriteByte(c byte) error {
	if !uart.core.dc.uartWriteByte(c) {
		return ErrUARTWriteFailed
	}
	return nil
}

// Write data to the UART.
func (uart UART) Write(data []byte) (n int, err error) {
	return uart.core.dc.uartWrite(data), nil
}
