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

// UART
const (
	DEFAULT_UART_TX_PIN    = PA7
	DEFAULT_UART_RX_PIN    = PA8
	DEFAULT_UART_TX_PIN_AF = 8
	DEFAULT_UART_RX_PIN_AF = 8
)
