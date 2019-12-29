// +build circuitplay_bluefruit

package machine

const HasLowFrequencyCrystal = true

// Hardware pins
const (
	P0_00 Pin = 0
	P0_01 Pin = 1
	P0_02 Pin = 2
	P0_03 Pin = 3
	P0_04 Pin = 4
	P0_05 Pin = 5
	P0_06 Pin = 6
	P0_07 Pin = 7
	P0_08 Pin = 8
	P0_09 Pin = 9
	P0_10 Pin = 10
	P0_11 Pin = 11
	P0_12 Pin = 12
	P0_13 Pin = 13
	P0_14 Pin = 14
	P0_15 Pin = 15
	P0_16 Pin = 16
	P0_17 Pin = 17
	P0_18 Pin = 18
	P0_19 Pin = 19
	P0_20 Pin = 20
	P0_21 Pin = 21
	P0_22 Pin = 22
	P0_23 Pin = 23
	P0_24 Pin = 24
	P0_25 Pin = 25
	P0_26 Pin = 26
	P0_27 Pin = 27
	P0_28 Pin = 28
	P0_29 Pin = 29
	P0_30 Pin = 30
	P0_31 Pin = 31
	P1_00 Pin = 32
	P1_01 Pin = 33
	P1_02 Pin = 34
	P1_03 Pin = 35
	P1_04 Pin = 36
	P1_05 Pin = 37
	P1_06 Pin = 38
	P1_07 Pin = 39
	P1_08 Pin = 40
	P1_09 Pin = 41
	P1_10 Pin = 42
	P1_11 Pin = 43
	P1_12 Pin = 44
	P1_13 Pin = 45
	P1_14 Pin = 46
	P1_15 Pin = 47
)

// GPIO Pins
const (
	D0  = P0_30
	D1  = P0_14
	D2  = P0_05
	D3  = P0_04
	D4  = P1_02
	D5  = P1_15
	D6  = P0_02
	D7  = P1_06
	D8  = P0_13
	D9  = P0_29
	D10 = P0_03
	D11 = P1_04
	D12 = P0_26
	D13 = P1_14
)

// Analog Pins
const (
	A1 = P0_02
	A2 = P0_29
	A3 = P0_03
	A4 = P0_04
	A5 = P0_05
	A6 = P0_30
	A7 = P0_14
	A8 = P0_28
	A9 = P0_31
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
)

// UART0 pins (logical UART1)
const (
	UART_TX_PIN = P0_14 // PORTB
	UART_RX_PIN = P0_30 // PORTB
)

// I2C pins
const (
	SDA_PIN = P0_05 // I2C0 external
	SCL_PIN = P0_04 // I2C0 external

	SDA1_PIN = P0_00 // I2C1 internal
	SCL1_PIN = P0_01 // I2C1 internal
)

// SPI pins (internal flash)
const (
	SPI0_SCK_PIN  = P0_19 // SCK
	SPI0_MOSI_PIN = P0_21 // MOSI
	SPI0_MISO_PIN = P0_23 // MISO
)
