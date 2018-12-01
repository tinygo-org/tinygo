// +build nrf,pca10040

package machine

// The PCA10040 has a low-frequency (32kHz) crystal oscillator on board.
const HasLowFrequencyCrystal = true

// LEDs on the PCA10040 (nRF52832 dev board)
const (
	LED  = LED1
	LED1 = 17
	LED2 = 18
	LED3 = 19
	LED4 = 20
)

// Buttons on the PCA10040 (nRF52832 dev board)
const (
	BUTTON  = BUTTON1
	BUTTON1 = 13
	BUTTON2 = 14
	BUTTON3 = 15
	BUTTON4 = 16
)

// UART pins for NRF52840-DK
const (
	UART_TX_PIN = 6
	UART_RX_PIN = 8
)

// ADC pins
const (
	ADC0 = 3
	ADC1 = 4
	ADC2 = 28
	ADC3 = 29
	ADC4 = 30
	ADC5 = 31
)

// I2C pins
const (
	SDA_PIN = 26
	SCL_PIN = 27
)

// SPI pins
const (
	SPI0_SCK_PIN  = 25
	SPI0_MOSI_PIN = 23
	SPI0_MISO_PIN = 24

	SPI1_SCK_PIN  = 2
	SPI1_MOSI_PIN = 3
	SPI1_MISO_PIN = 4
)
