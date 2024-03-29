//go:build esp32_coreboard_v2

package machine

const (
	CLK  = GPIO6
	CMD  = GPIO11
	IO0  = GPIO0
	IO1  = GPIO1
	IO2  = GPIO2
	IO3  = GPIO3
	IO4  = GPIO4
	IO5  = GPIO5
	IO9  = GPIO9
	IO10 = GPIO10
	IO16 = GPIO16
	IO17 = GPIO17
	IO18 = GPIO18
	IO19 = GPIO19
	IO21 = GPIO21
	IO22 = GPIO22
	IO23 = GPIO23
	IO25 = GPIO25
	IO26 = GPIO26
	IO27 = GPIO27
	IO32 = GPIO32
	IO33 = GPIO33
	IO34 = GPIO34
	IO35 = GPIO35
	IO36 = GPIO36
	IO39 = GPIO39
	RXD  = GPIO3
	SD0  = GPIO7
	SD1  = GPIO8
	SD2  = GPIO9
	SD3  = GPIO10
	SVN  = GPIO39
	SVP  = GPIO36
	TCK  = GPIO13
	TD0  = GPIO15
	TDI  = GPIO12
	TMS  = GPIO14
	TXD  = GPIO1
)

// Built-in LED on some ESP32 boards.
const LED = IO2

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
