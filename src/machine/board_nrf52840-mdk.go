// +build nrf52840_mdk

package machine

const HasLowFrequencyCrystal = true

// LEDs on the nrf52840-mdk (nRF52840 dev board)
const (
	LED       = LED_GREEN
	LED_GREEN = 22
	LED_RED   = 23
	LED_BLUE  = 24
)

// UART pins
const (
	UART_TX_PIN = 20
	UART_RX_PIN = 19
)

// I2C pins (unused)
const (
	SDA_PIN = 0xff
	SCL_PIN = 0xff
)
