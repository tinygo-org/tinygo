// +build reelboard

package machine

const HasLowFrequencyCrystal = true

// LEDs on the reel board
const (
	LED          Pin = LED1
	LED1         Pin = LED_YELLOW
	LED2         Pin = LED_RED
	LED3         Pin = LED_GREEN
	LED4         Pin = LED_BLUE
	LED_RED      Pin = 11
	LED_GREEN    Pin = 12
	LED_BLUE     Pin = 41
	LED_YELLOW   Pin = 13
	EPDBUSY_PIN  Pin = 14
	EPDRESET_PIN Pin = 15
	EPDDC_PIN    Pin = 16
	EPDCS_PIN    Pin = 17
	EPDSCK_PIN   Pin = 19
	EPDMOSI_PIN  Pin = 20
	RESET_PIN    Pin = 32
)

// User "a" button on the reel board
const (
	BUTTON Pin = 7
)

// UART pins
const (
	UART_TX_PIN Pin = 6
	UART_RX_PIN Pin = 8
)

// I2C pins
const (
	SDA_PIN Pin = 26
	SCL_PIN Pin = 27
)

// SPI pins
const (
	SPI0_SCK_PIN  Pin = 47
	SPI0_MOSI_PIN Pin = 45
	SPI0_MISO_PIN Pin = 46
)
