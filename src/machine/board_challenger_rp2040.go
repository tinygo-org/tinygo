//go:build challenger_rp2040

package machine

import (
	"device/rp"
	"runtime/interrupt"
)

const (
	LED = GPIO24

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// GPIO Pins
const (
	D5  = GPIO2
	D6  = GPIO3
	D9  = GPIO4
	D10 = GPIO5
	D11 = GPIO6
	D12 = GPIO7
	D13 = GPIO8
)

// Analog pins
const (
	A0 = ADC0
	A1 = ADC1
	A2 = ADC2
	A3 = ADC3
)

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
	SPI0_SCK_PIN = GPIO22
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO23 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO20 // Rx

	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO10
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO11 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO12 // Rx
)

// LoRa default pins
const (
	LORA_CS    = GPIO9
	LORA_SCK   = GPIO10
	LORA_SDO   = GPIO11
	LORA_SDI   = GPIO12
	LORA_RESET = GPIO13
	LORA_DIO0  = GPIO14
	LORA_DIO1  = GPIO15
	LORA_DIO2  = GPIO18
)

// UART pins
const (
	UART0_TX_PIN = GPIO16
	UART0_RX_PIN = GPIO17
	UART1_TX_PIN = GPIO4
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
	usb_STRING_PRODUCT      = "Challenger 2040 LoRa"
	usb_STRING_MANUFACTURER = "iLabs"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x1023
)
