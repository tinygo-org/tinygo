//go:build esp32c3_generic

package machine

const (
	SDA_PIN = NoPin
	SCL_PIN = NoPin
)

// UART0 pins. See ESP-IDF soc/esp32c3/include/soc/uart_pins.h.
const (
	UART0_TX_PIN = GPIO21
	UART0_RX_PIN = GPIO20
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)
