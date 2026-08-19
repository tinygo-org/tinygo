//go:build embedfire_py32f002b

// Pin mappings for the Embedfire PY32F002B board.

package machine

// LEDs
const (
	LED1 = PA1
	LED2 = PA5
	LED3 = PA4
	LED  = LED2
)

// Buttons
const (
	KEY1 = PA3
	KEY2 = PA0
)

const (
	UART_TX_PIN = PA6
	UART_RX_PIN = PA7
)

func defaultUARTPins() (Pin, Pin) {
	return UART_TX_PIN, UART_RX_PIN
}
