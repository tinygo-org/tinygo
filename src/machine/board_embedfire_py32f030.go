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

func configureDefaultUARTPins() {
	PA7.Configure(PinConfig{Mode: PinAlternate})
	PA7.SetAltFunc(8)
	PA8.Configure(PinConfig{Mode: PinAlternate})
	PA8.SetAltFunc(8)
}
