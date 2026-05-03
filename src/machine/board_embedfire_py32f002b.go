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
	PA6.Configure(PinConfig{Mode: PinAlternate})
	PA6.SetAltFunc(1)
	PA7.Configure(PinConfig{Mode: PinAlternate})
	PA7.SetAltFunc(3)
}
