// +build sam,atsamd21,itsybitsy_m0

package machine

import "device/sam"

// GPIO Pins
const (
	D0  = PA11 // UART0 RX
	D1  = PA10 // UART0 TX
	D2  = PA14
	D3  = PA09 // PWM available
	D4  = PA08 // PWM available
	D5  = PA15 // PWM available
	D6  = PA20 // PWM available
	D7  = PA21 // PWM available
	D8  = PA06 // PWM available
	D9  = PA07 // PWM available
	D10 = PA18 // can be used for PWM or UART1 TX
	D11 = PA16 // can be used for PWM or UART1 RX
	D12 = PA19 // PWM available
	D13 = PA17 // PWM available
)

// Analog pins
const (
	A0 = PA02 // ADC/AIN[0]
	A1 = PB08 // ADC/AIN[2]
	A2 = PB09 // ADC/AIN[3]
	A3 = PA04 // ADC/AIN[4]
	A4 = PA05 // ADC/AIN[5]
	A5 = PB02 // ADC/AIN[10]
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
	UART_TX_PIN = D10
	UART_RX_PIN = D11
)

// I2C pins
const (
	SDA_PIN = PA22 // SDA: SERCOM3/PAD[0]
	SCL_PIN = PA23 // SCL: SERCOM3/PAD[1]
)

// I2C on the ItsyBitsy M0.
var (
	I2C0 = I2C{Bus: sam.SERCOM3_I2CM,
		SDA:     SDA_PIN,
		SCL:     SCL_PIN,
		PinMode: GPIO_SERCOM}
)

// SPI pins
const (
	SPI0_SCK_PIN  = PB11 // SCK: SERCOM4/PAD[3]
	SPI0_MOSI_PIN = PB10 // MOSI: SERCOM4/PAD[2]
	SPI0_MISO_PIN = PA12 // MISO: SERCOM4/PAD[0]
)

// SPI on the ItsyBitsy M0.
var (
	SPI0 = SPI{Bus: sam.SERCOM4_SPI}
)

// I2S pins
const (
	I2S_SCK_PIN = PA10
	I2S_SD_PIN  = PA08
	I2S_WS_PIN  = 0xff // TODO: figure out what this is on ItsyBitsy M0.
)

// I2S on the ItsyBitsy M0.
var (
	I2S0 = I2S{Bus: sam.I2S}
)
