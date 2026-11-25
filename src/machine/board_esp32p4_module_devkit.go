//go:build esp32p4_module_devkit

package machine

// Pins for the Waveshare ESP32-P4-Module-DEV-KIT.
// https://docs.waveshare.com/ESP32-P4-Module-DEV-KIT
const (
	UART0_TX_PIN = GPIO37
	UART0_RX_PIN = GPIO38
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN

	SDA_PIN = GPIO7
	SCL_PIN = GPIO8
)
