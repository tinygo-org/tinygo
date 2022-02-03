//go:build esp32c312f
// +build esp32c312f

package machine

// Built-in RGB LED
const (
	LED_RED   = IO3
	LED_GREEN = IO4
	LED_BLUE  = IO5
	LED       = LED_RED
)

const (
	IO0  Pin = 0
	IO1  Pin = 1
	IO10 Pin = 10
	IO18 Pin = 18
	IO19 Pin = 19
	IO2  Pin = 2
	IO3  Pin = 3
	IO4  Pin = 4
	IO5  Pin = 5
	IO6  Pin = 6
	IO7  Pin = 7
	IO8  Pin = 8
	IO9  Pin = 9
	RXD  Pin = 20
	TXD  Pin = 21
)

// ADC pins
const (
	ADC0 Pin = ADC1_0
	ADC1 Pin = ADC2_0

	ADC1_0 Pin = IO0
	ADC1_1 Pin = IO1
	ADC1_2 Pin = IO2
	ADC1_3 Pin = IO3
	ADC1_4 Pin = IO4
	ADC2_0 Pin = IO5
)

// UART0 pins
const (
	UART_TX_PIN = TXD
	UART_RX_PIN = RXD
)
