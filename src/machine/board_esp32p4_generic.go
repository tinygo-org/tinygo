//go:build esp32p4 && !esp32p4_nano && !crowpanel_esp32p4_10_1

package machine

// Default pins for the generic ESP32-P4 target. Board targets override these.
// UART0 uses the ESP32-P4 default pins; I2C uses the pins from the Waveshare
// ESP32-P4-Nano board file.
const (
	UART_TX_PIN = GPIO37
	UART_RX_PIN = GPIO38

	SDA_PIN = GPIO7
	SCL_PIN = GPIO8
)
