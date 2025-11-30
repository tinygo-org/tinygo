//go:build embedfire_py32f002b

// Pin mappings for the Embedfire PY32F002B board.
// Only LED and button aliases are provided.

package machine

// LEDs
const (
	LED2 = PA1
	LED3 = PA5
	LED4 = PA4
	LED  = LED2
)

// Buttons
const (
	KEY1 = PA3
	KEY2 = PA0
)

// UART
const (
	DEFAULT_UART_TX_PIN    = PA6
	DEFAULT_UART_RX_PIN    = PA7
	DEFAULT_UART_TX_PIN_AF = 1
	DEFAULT_UART_RX_PIN_AF = 3
)
