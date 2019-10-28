// +build sam,atsamd51,itsybitsy_m4

package machine

import "device/sam"

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0  = PA16 // UART0 RX/PWM available
	D1  = PA17 // UART0 TX/PWM available
	D2  = PA07
	D3  = PB22
	D4  = PA14 // PWM available
	D5  = PA15 // PWM available
	D6  = PB02 // dotStar clock
	D7  = PA18 // PWM available
	D8  = PB03 // dotStar data
	D9  = PA19 // PWM available
	D10 = PA20 // can be used for PWM or UART1 TX
	D11 = PA21 // can be used for PWM or UART1 RX
	D12 = PA23 // PWM available
	D13 = PA22 // PWM available
)

// Analog pins
const (
	A0 = PA02 // ADC/AIN[0]
	A1 = PA05 // ADC/AIN[2]
	A2 = PB08 // ADC/AIN[3]
	A3 = PB09 // ADC/AIN[4]
	A4 = PA04 // ADC/AIN[5]
	A5 = PA06 // ADC/AIN[10]
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
	SDA_PIN = PA12 // SDA: SERCOM3/PAD[0]
	SCL_PIN = PA13 // SCL: SERCOM3/PAD[1]
)

// I2C on the ItsyBitsy M4.
var (
	I2C0 = I2C{Bus: sam.SERCOM2_I2CM,
		SDA:     SDA_PIN,
		SCL:     SCL_PIN,
		PinMode: PinSERCOM}
)

// SPI pins
const (
	SPI0_SCK_PIN  = PA01 // SCK: SERCOM1/PAD[1]
	SPI0_MOSI_PIN = PA00 // MOSI: SERCOM1/PAD[0]
	SPI0_MISO_PIN = PB23 // MISO: SERCOM1/PAD[3]
)

// SPI on the ItsyBitsy M4.
var (
	SPI0 = SPI{Bus: sam.SERCOM1_SPIM,
		SCK:         SPI0_SCK_PIN,
		MOSI:        SPI0_MOSI_PIN,
		MISO:        SPI0_MISO_PIN,
		DOpad:       spiTXPad2SCK3,
		DIpad:       sercomRXPad3,
		MISOPinMode: PinSERCOM,
	}
)
