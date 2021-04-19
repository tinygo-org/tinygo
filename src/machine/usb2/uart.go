package usb2

import (
	"errors"
)

var (
	ErrInvalidPort = errors.New("invalid USB port")
)

type (
	UARTConfig struct {
		BaudRate uint32
	}

	// UART represents a virtual serial (UART) device emulation using the USB
	// CDC-ACM device class driver.
	UART struct {
		port int // USB port (core index, e.g., 0-1)
		core *core
	}
)

func (uart *UART) Configure(config UARTConfig) error {

	if uart.port >= CoreCount || uart.port >= dcdCount {
		return ErrInvalidPort
	}

	// verify we have a free USB port and take ownership of it
	var st status
	uart.core, st = initCore(uart.port, class{id: classDeviceCDCACM, config: 1})
	if !st.ok() {
		return ErrInvalidPort
	}
	return nil
}
