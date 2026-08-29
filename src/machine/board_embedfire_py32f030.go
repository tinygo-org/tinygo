//go:build embedfire_py32f030

// Pin mappings for the Embedfire PY32F030 board.

package machine

// LEDs
const (
	LED1 = PA2
	LED2 = PA3
	LED3 = PA4
	LED  = LED2
)

// Buttons
const (
	KEY1 = PA5
	KEY2 = PA6
)

const (
	UART_TX_PIN = PA7
	UART_RX_PIN = PA8
)

func defaultUARTPins() (Pin, Pin) {
	return UART_TX_PIN, UART_RX_PIN
}
