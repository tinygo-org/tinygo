//go:build feather_rp2040

package machine

import (
	"device/rp"
	"runtime/interrupt"
)

// Onboard crystal oscillator frequency, in MHz.
const xoscFreq = 12 // MHz

// GPIO Pins
const (
	D4  = GPIO6
	D5  = GPIO7
	D6  = GPIO8
	D9  = GPIO9
	D10 = GPIO10
	D11 = GPIO11
	D12 = GPIO12
	D13 = GPIO13
	D24 = GPIO24
	D25 = GPIO25
)

// Analog pins
const (
	A0 = GPIO26
	A1 = GPIO27
	A2 = GPIO28
	A3 = GPIO29
)

const LED = GPIO13

// I2C Pins.
const (
	I2C0_SDA_PIN = GPIO24
	I2C0_SCL_PIN = GPIO25

	I2C1_SDA_PIN = GPIO2
	I2C1_SCL_PIN = GPIO3

	SDA_PIN = I2C1_SDA_PIN
	SCL_PIN = I2C1_SCL_PIN
)

// SPI default pins
const (
	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO18
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO19 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO20 // Rx

	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO10
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO11 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO12 // Rx
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART1_TX_PIN = GPIO8
	UART1_RX_PIN = GPIO9
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

// UART on the RP2040
var (
	UART0  = &_UART0
	_UART0 = UART{
		UARTCommon: NewUARTCommon(),
		Bus:        rp.UART0,
	}

	UART1  = &_UART1
	_UART1 = UART{
		UARTCommon: NewUARTCommon(),
		Bus:        rp.UART1,
	}
)

var DefaultUART = UART0

func init() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, _UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(rp.IRQ_UART1_IRQ, _UART1.handleInterrupt)
}

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Feather RP2040"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x80F1
)
