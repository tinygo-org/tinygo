//go:build qtpy_rp2040

package machine

import (
	"device/rp"
	"runtime/interrupt"
)

// Onboard crystal oscillator frequency, in MHz.
const xoscFreq = 12 // MHz

// GPIO Pins
const (
	SDA  = GPIO24
	SCL  = GPIO25
	TX   = GPIO20
	MO   = GPIO3
	MOSI = GPIO3
	MI   = GPIO4
	MISO = GPIO4
	SCK  = GPIO6
	RX   = GPIO5

	QT_SCL1 = GPIO23
	QT_SDA1 = GPIO22
)

// Analog pins
const (
	A0 = GPIO29
	A1 = GPIO28
	A2 = GPIO27
	A3 = GPIO26
)

const (
	NEOPIXEL       = GPIO12
	WS2812         = GPIO12
	NEOPIXEL_POWER = GPIO11
)

// I2C Pins.
const (
	I2C0_SDA_PIN = GPIO24
	I2C0_SCL_PIN = GPIO25

	I2C1_SDA_PIN = GPIO26
	I2C1_SCL_PIN = GPIO27

	I2C1_QT_SDA_PIN = GPIO22
	I2C1_QT_SCL_PIN = GPIO23

	SDA_PIN = GPIO24
	SCL_PIN = GPIO25
)

// SPI default pins
const (
	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO6
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO3 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO4 // Rx
	SPI0_CS      = GPIO5

	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO26
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO27 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO24 // Rx
	SPI1_CS      = GPIO25
)

// UART pins
const (
	UART0_TX_PIN = GPIO28
	UART0_RX_PIN = GPIO29
	UART1_TX_PIN = GPIO20
	UART1_RX_PIN = GPIO5
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

// UART on the RP2040
var (
	UART0  = &_UART0
	_UART0 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART0,
	}

	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART1,
	}
)

var DefaultUART = UART0

func init() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, _UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(rp.IRQ_UART1_IRQ, _UART1.handleInterrupt)
}

// USB identifiers
const (
	usb_STRING_PRODUCT      = "QT Py RP2040"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x80F7
)
