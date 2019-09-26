// +build sam,atsamd21,circuitplay_express

package machine

import "device/sam"

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0  = PB09
	D1  = PB08
	D2  = PB02
	D3  = PB03
	D4  = PA28
	D5  = PA14
	D6  = PA05
	D7  = PA15
	D8  = PB23
	D9  = PA06
	D10 = PA07
	D11 = NoPin // does not seem to exist
	D12 = PA02
	D13 = PA17 // PWM available
)

// Analog Pins
const (
	A0  = PA02 // PWM available, also ADC/AIN[0]
	A1  = PA05 // ADC/AIN[5]
	A2  = PA06 // PWM available, also ADC/AIN[6]
	A3  = PA07 // PWM available, also ADC/AIN[7]
	A4  = PB03 // PORTB
	A5  = PB02 // PORTB
	A6  = PB09 // PORTB
	A7  = PB08 // PORTB
	A8  = PA11 // ADC/AIN[19]
	A9  = PA09 // ADC/AIN[17]
	A10 = PA04
)

const (
	LED       = D13
	NEOPIXELS = D8

	BUTTONA = D4
	BUTTONB = D5
	SLIDER  = D7 // built-in slide switch

	BUTTON  = BUTTONA
	BUTTON1 = BUTTONB

	LIGHTSENSOR = A8
	TEMPSENSOR  = A9
	PROXIMITY   = A10
)

// USBCDC pins (logical UART0)
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART0 pins (logical UART1)
const (
	UART_TX_PIN = PB08 // PORTB
	UART_RX_PIN = PB09 // PORTB
)

// UART1 on the Circuit Playground Express.
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

// I2C pins
const (
	SDA_PIN = PB02 // I2C0 external
	SCL_PIN = PB03 // I2C0 external

	SDA1_PIN = PA00 // I2C1 internal
	SCL1_PIN = PA01 // I2C1 internal
)

// I2C on the Circuit Playground Express.
var (
	// external device
	I2C0 = I2C{Bus: sam.SERCOM5_I2CM,
		SDA:     SDA_PIN,
		SCL:     SCL_PIN,
		PinMode: PinSERCOM}
	// internal device
	I2C1 = I2C{Bus: sam.SERCOM1_I2CM,
		SDA:     SDA1_PIN,
		SCL:     SCL1_PIN,
		PinMode: PinSERCOMAlt}
)

// SPI pins (internal flash)
const (
	SPI0_SCK_PIN  = PA21 // SCK: SERCOM3/PAD[3]
	SPI0_MOSI_PIN = PA20 // MOSI: SERCOM3/PAD[2]
	SPI0_MISO_PIN = PA16 // MISO: SERCOM3/PAD[0]
)

// SPI on the Circuit Playground Express.
var (
	SPI0 = SPI{Bus: sam.SERCOM3_SPI,
		SERCOM: 3,
		SCK:     SPI0_SCK_PIN,
		MOSI:    SPI0_MOSI_PIN,
		MISO:    SPI0_MISO_PIN,
		DOpad:   spiTXPad2SCK3,
		DIpad:   sercomRXPad0,
		PinMode: PinSERCOMAlt}
)

// I2S pins
const (
	I2S_SCK_PIN = PA10
	I2S_SD_PIN  = PA08
	I2S_WS_PIN  = NoPin // no WS, instead uses SCK to sync
)

// I2S on the Circuit Playground Express.
var (
	I2S0 = I2S{Bus: sam.I2S}
)
