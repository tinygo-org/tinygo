// +build sam,atsamd21,trinket_m0

package machine

import "device/sam"

// GPIO Pins
const (
	D0  = PA08 // PWM available
	D1  = PA02
	D2  = PA09 // PWM available
	D3  = PA07 // PWM available / UART0 RX
	D4  = PA06 // PWM available / UART0 TX
	D13 = PA10 // LED
)

// Analog pins
const (
	A0 = D1
	A1 = D2
	A2 = D0
	A3 = D3
	A4 = D4
)

const (
	LED = D13
)

// UART0 aka USBCDC pins
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART1 pins
const (
	UART_TX_PIN = D4
	UART_RX_PIN = D3
)

// UART1 on the Trinket M0.
var (
	UART1 = UART{Bus: sam.SERCOM1_USART,
		Buffer: NewRingBuffer(),
		Mode:   PinSERCOM,
		IRQVal: sam.IRQ_SERCOM1,
	}
)

//go:export SERCOM1_IRQHandler
func handleUART1() {
	defaultUART1Handler()
}

// SPI pins
const (
	SPI0_SCK_PIN  = D3
	SPI0_MOSI_PIN = D4
	SPI0_MISO_PIN = D2
)

// SPI on the Trinket M0.
var (
	SPI0 = SPI{Bus: sam.SERCOM0_SPI}
)

// I2C pins
const (
	SDA_PIN = D0 // SDA
	SCL_PIN = D2 // SCL
)

// I2C on the Trinket M0.
var (
	I2C0 = I2C{Bus: sam.SERCOM2_I2CM,
		SDA:     SDA_PIN,
		SCL:     SCL_PIN,
		PinMode: PinSERCOMAlt}
)

// I2S pins
const (
	I2S_SCK_PIN = PA10
	I2S_SD_PIN  = PA08
	I2S_WS_PIN  = NoPin // TODO: figure out what this is on Trinket M0.
)
