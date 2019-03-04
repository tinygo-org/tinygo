// +build sam,atsamd21,circuitplay_express

package machine

import "device/sam"

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
	D11 = 0xff // does not seem to exist
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

// USBCDC pins
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART0 pins
const (
	UART_TX_PIN = PB08 // PORTB
	UART_RX_PIN = PB09 // PORTB
)

// I2C pins
const (
	SDA_PIN = PA00 // SDA: SERCOM3/PAD[0]
	SCL_PIN = PA01 // SCL: SERCOM3/PAD[1]
)

// I2C on the Circuit Playground Express.
var (
	I2C0 = I2C{Bus: sam.SERCOM5_I2CM} // external device
	I2C1 = I2C{Bus: sam.SERCOM1_I2CM} // internal device
)
