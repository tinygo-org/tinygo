//go:build esp32_coreboard_v2
// +build esp32_coreboard_v2

package machine

// Built-in LED on some ESP32 boards.
const LED = IO2

const (
	CLK  Pin = 6
	CMD  Pin = 11
	IO0  Pin = 0
	IO1  Pin = 1
	IO10 Pin = 10
	IO16 Pin = 16
	IO17 Pin = 17
	IO18 Pin = 18
	IO19 Pin = 19
	IO2  Pin = 2
	IO21 Pin = 21
	IO22 Pin = 22
	IO23 Pin = 23
	IO25 Pin = 25
	IO26 Pin = 26
	IO27 Pin = 27
	IO3  Pin = 3
	IO32 Pin = 32
	IO33 Pin = 33
	IO34 Pin = 34
	IO35 Pin = 35
	IO36 Pin = 36
	IO39 Pin = 39
	IO4  Pin = 4
	IO5  Pin = 5
	IO9  Pin = 9
	RXD  Pin = 3
	SD0  Pin = 7
	SD1  Pin = 8
	SD2  Pin = 9
	SD3  Pin = 10
	SVN  Pin = 39
	SVP  Pin = 36
	TCK  Pin = 13
	TD0  Pin = 15
	TDI  Pin = 12
	TMS  Pin = 14
	TXD  Pin = 1
)

// SPI pins
const (
	SPI0_SCK_PIN = IO18
	SPI0_SDO_PIN = IO23
	SPI0_SDI_PIN = IO19
	SPI0_CS0_PIN = IO5
)

// I2C pins
const (
	SDA_PIN = IO21
	SCL_PIN = IO22
)

// ADC pins
const (
	ADC0 Pin = IO34
	ADC1 Pin = IO35
	ADC2 Pin = IO36
	ADC3 Pin = IO39
)

// UART0 pins
const (
	UART_TX_PIN = IO1
	UART_RX_PIN = IO3
)

// UART1 pins
const (
	UART1_TX_PIN = IO9
	UART1_RX_PIN = IO10
)

// PWM pins
const (
	PWM0_PIN Pin = IO2
	PWM1_PIN Pin = IO0
	PWM2_PIN Pin = IO4
)
