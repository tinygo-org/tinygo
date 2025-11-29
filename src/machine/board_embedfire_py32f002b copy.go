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
