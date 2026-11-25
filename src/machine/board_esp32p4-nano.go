//go:build esp32p4_nano

package machine

// Minimal board pins for the WaveShare ESP32-P4-NANO.

const (
	UART0_TX_PIN = GPIO37
	UART0_RX_PIN = GPIO38
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN

	SDA_PIN = GPIO7
	SCL_PIN = GPIO8
)
