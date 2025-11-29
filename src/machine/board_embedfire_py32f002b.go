//go:build embedfire_py32f030

// Pin mappings for the Embedfire PY32F030 board.
// Only LED and button aliases are provided.

package machine

// LEDs
const (
	LED2 = PA2
	LED3 = PA3
	LED4 = PA4
	LED  = LED2
)

// Buttons
const (
	KEY1 = PA5
	KEY2 = PA6
)
