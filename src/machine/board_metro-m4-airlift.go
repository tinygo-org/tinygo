// +build sam,atsamd51,metro_m4_airlift

package machine

import "device/sam"

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0 = PA23 // UART0 RX/PWM available
	D1 = PA22 // UART0 TX/PWM available
	D2 = PB17 // PWM available
	D3 = PB16 // PWM available
	D4 = PB13 // PWM available
	D5 = PB14 // PWM available
	D6 = PB15 // PWM available
	D7 = PB12 // PWM available

	D8  = PA21 // PWM available
	D9  = PA20 // PWM available
	D10 = PA18 // can be used for PWM or UART1 TX
	D11 = PA19 // can be used for PWM or UART1 RX
	D12 = PA17 // PWM available
	D13 = PA16 // PWM available

	D40 = PB22 // built-in neopixel
)

// Analog pins
const (
	A0 = PA02 // ADC/AIN[0]
	A1 = PA05 // ADC/AIN[2]
	A2 = PB06 // ADC/AIN[3]
	A3 = PB00 // ADC/AIN[4] // NOTE: different between "airlift" and non-airlift versions
	A4 = PB08 // ADC/AIN[5]
	A5 = PB09 // ADC/AIN[10]
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
	UART_TX_PIN = D1
	UART_RX_PIN = D0
)

// I2C pins
const (
	SDA_PIN = PB02 // SDA: SERCOM2/PAD[0]
	SCL_PIN = PB03 // SCL: SERCOM2/PAD[1]
)

// I2C on the Feather M4.
var (
	I2C0 = I2C{Bus: sam.SERCOM2_I2CM,
		SDA:     SDA_PIN,
		SCL:     SCL_PIN,
		PinMode: PinSERCOM}
)

// SPI pins
const (
	SPI0_SCK_PIN  = PA13 // SCK: SERCOM1/PAD[1]
	SPI0_MOSI_PIN = PA12 // MOSI: SERCOM1/PAD[3]
	SPI0_MISO_PIN = PA14 // MISO: SERCOM1/PAD[2]
)

// SPI on the Feather M4.
var (
	SPI0 = SPI{Bus: sam.SERCOM1_SPIM,
		SCK:         SPI0_SCK_PIN,
		MOSI:        SPI0_MOSI_PIN,
		MISO:        SPI0_MISO_PIN,
		DOpad:       spiTXPad3SCK1,
		DIpad:       sercomRXPad2,
		MISOPinMode: PinSERCOM,
		MOSIPinMode: PinSERCOM,
		SCKPinMode:  PinSERCOM,
	}
)
