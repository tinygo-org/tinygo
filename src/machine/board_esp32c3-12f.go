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
	IO0  = GPIO0
	IO1  = GPIO1
	IO2  = GPIO2
	IO3  = GPIO3
	IO4  = GPIO4
	IO5  = GPIO5
	IO6  = GPIO6
	IO7  = GPIO7
	IO8  = GPIO8
	IO9  = GPIO9
	IO10 = GPIO10
	IO18 = GPIO18
	IO19 = GPIO19
	RXD  = GPIO20
	TXD  = GPIO21
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
