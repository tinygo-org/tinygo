// +build nrf52840_mdk

package machine

const HasLowFrequencyCrystal = true

// LEDs on the nrf52840-mdk (nRF52840 dev board)
const (
	LED       Pin = LED_GREEN
	LED_GREEN Pin = 22
	LED_RED   Pin = 23
	LED_BLUE  Pin = 24
)

// UART pins
const (
	UART_TX_PIN Pin = 20
	UART_RX_PIN Pin = 19
)

// UART0 is the USB device and UART1 is the NRF UART
var (
	UART0 = USB
	UART1 = NRF_UART0
)

// I2C pins (unused)
const (
	SDA_PIN = NoPin
	SCL_PIN = NoPin
)

// SPI pins (unused)
const (
	SPI0_SCK_PIN  = NoPin
	SPI0_MOSI_PIN = NoPin
	SPI0_MISO_PIN = NoPin
)
