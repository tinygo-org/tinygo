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

func configureDefaultUARTPins() {
	ConfigureUARTPin(PA6, 1) // TX
	ConfigureUARTPin(PA7, 3) // RX
}
