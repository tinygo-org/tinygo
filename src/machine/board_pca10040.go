//go:build pca10040
// +build pca10040

package machine

// The PCA10040 has a low-frequency (32kHz) crystal oscillator on board.
const HasLowFrequencyCrystal = true

// LEDs on the PCA10040 (nRF52832 dev board)
const (
	LED1 Pin = 17
	LED2 Pin = 18
	LED3 Pin = 19
	LED4 Pin = 20
	LED  Pin = LED1
)

// Buttons on the PCA10040 (nRF52832 dev board)
const (
	BUTTON1 Pin = 13
	BUTTON2 Pin = 14
	BUTTON3 Pin = 15
	BUTTON4 Pin = 16
	BUTTON  Pin = BUTTON1
)

var DefaultUART = UART0

// UART pins for NRF52840-DK
const (
	UART_TX_PIN Pin = 6
	UART_RX_PIN Pin = 8
)

// ADC pins
const (
	ADC0 Pin = 3
	ADC1 Pin = 4
	ADC2 Pin = 28
	ADC3 Pin = 29
	ADC4 Pin = 30
	ADC5 Pin = 31
)

// I2C pins
const (
	SDA_PIN Pin = 26
	SCL_PIN Pin = 27
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = 25
	SPI0_SDO_PIN Pin = 23
	SPI0_SDI_PIN Pin = 24
)
