// +build reelboard

package machine

const HasLowFrequencyCrystal = true

// LEDs on the reel board
const (
	LED        = LED1
	LED1       = LED_YELLOW
	LED2       = LED_RED
	LED3       = LED_GREEN
	LED4       = LED_BLUE
	LED_RED    = 11
	LED_GREEN  = 12
	LED_BLUE   = 41
	LED_YELLOW = 13
)

// User "a" button on the reel board
const (
	BUTTON = 7
)

// UART pins
const (
	UART_TX_PIN = 6
	UART_RX_PIN = 8
)

// I2C pins
const (
	SDA_PIN = 26
	SCL_PIN = 27
)

// SPI pins
const (
	SPI0_SCK_PIN  = 47
	SPI0_MOSI_PIN = 45
	SPI0_MISO_PIN = 46
)
