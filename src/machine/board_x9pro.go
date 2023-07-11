//go:build x9pro

package machine

// https://hackaday.io/project/144350-hacking-wearables-for-mental-health-and-more/details
const (
	LED          Pin = 4 // HR LED pin
	UART_TX_PIN  Pin = NoPin
	UART_RX_PIN  Pin = NoPin
	SCL_PIN      Pin = NoPin
	SDA_PIN      Pin = NoPin
	SPI0_SCK_PIN Pin = 18
	SPI0_SDI_PIN Pin = 19
	SPI0_SDO_PIN Pin = 20
)

// LCD pins.
const (
	OLED_CS      Pin = 15 // chip select
	OLED_RES     Pin = 14 // reset pin
	OLED_DC      Pin = 13 // data/command
	OLED_SCK     Pin = 12 // SPI clock
	OLED_SDO     Pin = 11 // SPI SDO (chip-out, peripheral-in)
	OLED_LED_POW Pin = 16
	OLED_IC_POW  Pin = 17
)

const HasLowFrequencyCrystal = true

var DefaultUART = UART0
