//go:build arduino_due

// Pin definitions are derived from ArduinoCore-sam 1.6.12
// variants/arduino_due_x/variant.cpp.
//
// Flashing requires the Arduino fork of bossac, found in PATH as "bossac".
// See https://github.com/arduino/BOSSA

package machine

const (
	D0  Pin = PA8
	D1  Pin = PA9
	D2  Pin = PB25
	D3  Pin = PC28
	D4  Pin = PC26
	D5  Pin = PC25
	D6  Pin = PC24
	D7  Pin = PC23
	D8  Pin = PC22
	D9  Pin = PC21
	D10 Pin = PC29
	D11 Pin = PD7
	D12 Pin = PD8
	D13 Pin = PB27
	D14 Pin = PD4
	D15 Pin = PD5
	D16 Pin = PA13
	D17 Pin = PA12
	D18 Pin = PA11
	D19 Pin = PA10
	D20 Pin = PB12
	D21 Pin = PB13
	D22 Pin = PB26
	D23 Pin = PA14
	D24 Pin = PA15
	D25 Pin = PD0
	D26 Pin = PD1
	D27 Pin = PD2
	D28 Pin = PD3
	D29 Pin = PD6
	D30 Pin = PD9
	D31 Pin = PA7
	D32 Pin = PD10
	D33 Pin = PC1
	D34 Pin = PC2
	D35 Pin = PC3
	D36 Pin = PC4
	D37 Pin = PC5
	D38 Pin = PC6
	D39 Pin = PC7
	D40 Pin = PC8
	D41 Pin = PC9
	D42 Pin = PA19
	D43 Pin = PA20
	D44 Pin = PC19
	D45 Pin = PC18
	D46 Pin = PC17
	D47 Pin = PC16
	D48 Pin = PC15
	D49 Pin = PC14
	D50 Pin = PC13
	D51 Pin = PC12
	D52 Pin = PB21
	D53 Pin = PB14
)

const (
	A0  Pin = PA16
	A1  Pin = PA24
	A2  Pin = PA23
	A3  Pin = PA22
	A4  Pin = PA6
	A5  Pin = PA4
	A6  Pin = PA3
	A7  Pin = PA2
	A8  Pin = PB17
	A9  Pin = PB18
	A10 Pin = PB19
	A11 Pin = PB20

	DAC0 Pin = PB15
	DAC1 Pin = PB16
)

const (
	ADC0  = A0
	ADC1  = A1
	ADC2  = A2
	ADC3  = A3
	ADC4  = A4
	ADC5  = A5
	ADC6  = A6
	ADC7  = A7
	ADC8  = A8
	ADC9  = A9
	ADC10 = A10
	ADC11 = A11
)

const (
	LED    = D13
	LED_RX = PC30
	LED_TX = PA21
)

const (
	UART_TX_PIN = D1
	UART_RX_PIN = D0
)

const (
	SDA_PIN  = D20
	SCL_PIN  = D21
	SDA1_PIN = PA17
	SCL1_PIN = PA18
)

const (
	SPI0_SDI_PIN = PA25
	SPI0_SDO_PIN = PA26
	SPI0_SCK_PIN = PA27
)

const (
	CAN0_RX_PIN = PA1
	CAN0_TX_PIN = PA0
)
