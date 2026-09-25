//go:build esp32h2_devkitm_1

// This file contains the pin mappings for the Espressif ESP32-H2-DevKitM-1 board.
//
// - https://docs.espressif.com/projects/esp-dev-kits/en/latest/esp32h2/esp32-h2-devkitm-1/user_guide.html

package machine

// Data pin of the onboard addressable RGB LED. The board has no plain LED.
const (
	WS2812 = GPIO8
)

// BOOT button
const (
	BUTTON = GPIO9
)

// UART pins
const (
	UART_TX_PIN = GPIO24
	UART_RX_PIN = GPIO23
)
