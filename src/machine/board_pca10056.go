// +build pca10056

package machine

const HasLowFrequencyCrystal = true

// LEDs on the pca10056
const (
	LED  = LED1
	LED1 = 13
	LED2 = 14
	LED3 = 15
	LED4 = 16
)

// Buttons on the pca10056
const (
	BUTTON  = BUTTON1
	BUTTON1 = 11
	BUTTON2 = 12
	BUTTON3 = 24
	BUTTON4 = 25
)

// UART pins
const (
	UART_TX_PIN = 6
	UART_RX_PIN = 8
)

// I2C pins
const (
	SDA_PIN = 26 // P0.26
	SCL_PIN = 27 // P0.27
)

// SPI pins
const (
	SPI0_SCK_PIN  = 47 // P1.15
	SPI0_MOSI_PIN = 45 // P1.13
	SPI0_MISO_PIN = 46 // P1.14
)
