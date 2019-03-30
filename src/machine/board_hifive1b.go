// +build hifive1b

package machine

const (
	LED       = LED1
	LED1      = LED_RED
	LED2      = LED_GREEN
	LED3      = LED_BLUE
	LED_RED   = P22
	LED_GREEN = P19
	LED_BLUE  = P21
)

const (
	// TODO: figure out the pin numbers for these.
	UART_TX_PIN = NoPin
	UART_RX_PIN = NoPin
)
