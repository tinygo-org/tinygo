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

	if uart.port >= CoreCount || uart.port >= dciCount ||
		uart.port >= descCDCACMConfigCount {
		return ErrInvalidPort
	}

	// use default configuration index (1-based index; 0=invalid)
	// uart.config = configDeviceCDCACMConfigurationIndex

	// modify the global basic configuration struct configDeviceCDCACM for our USB
	// port and configuration index.
	//
	// these settings are copied into the real CDC-ACM object, using interface
	// deviceClassDriver, via initialization method (*deviceCDCACM).init().

	// change baud rate from default, if provided
	//if config.BaudRate != 0 {
	//	configDeviceCDCACM[uart.port][uart.config-1].lineCodingBaudRate =
	//		config.BaudRate
	//}

	// verify we have a free USB port and take ownership of it
	var st status
	uart.core, st = initCore(uart.port, modeDevice)
	if !st.ok() {
		return ErrInvalidPort
	}
	return nil
}
