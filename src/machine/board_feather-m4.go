// +build sam,atsamd51,feather_m4

package machine

import "device/sam"

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0  = PB17 // UART0 RX/PWM available
	D1  = PB16 // UART0 TX/PWM available
	D4  = PA14 // PWM available
	D5  = PA16 // PWM available
	D6  = PA18 // PWM available
	D8  = PB03 // built-in neopixel
	D9  = PA19 // PWM available
	D10 = PA20 // can be used for PWM or UART1 TX
	D11 = PA21 // can be used for PWM or UART1 RX
	D12 = PA22 // PWM available
	D13 = PA23 // PWM available
	D21 = PA13 // PWM available
	D22 = PA12 // PWM available
	D23 = PB22 // PWM available
	D24 = PB23 // PWM available
	D25 = PA17 // PWM available
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
	SDA_PIN = D22 // SDA: SERCOM2/PAD[0]
	SCL_PIN = D21 // SCL: SERCOM2/PAD[1]
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
	SPI0_SCK_PIN  = D25 // SCK: SERCOM1/PAD[1]
	SPI0_MOSI_PIN = D24 // MOSI: SERCOM1/PAD[3]
	SPI0_MISO_PIN = D23 // MISO: SERCOM1/PAD[2]
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
