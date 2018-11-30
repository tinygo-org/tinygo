// +build nrf51,pca10031

// pca10031 is a nrf51 based dongle, intended for use in wireless applications.
//
// https://infocenter.nordicsemi.com/pdf/nRF51_Dongle_UG_v1.0.pdf
package machine

// The pca10031 has a 32kHz crystal on board.
const HasLowFrequencyCrystal = true

// LED on the pca10031
const (
	LED       = LED_RED
	LED1      = LED_RED
	LED2      = LED_GREEN
	LED3      = LED_BLUE
	LED_RED   = 21
	LED_GREEN = 22
	LED_BLUE  = 23
)

// UART pins
const (
	UART_TX_PIN = 9
	UART_RX_PIN = 11
)

// I2C pins (disabled)
const (
	SDA_PIN = 0xff
	SCL_PIN = 0xff
)

// SPI pins (unused)
const (
	SPI0_SCK_PIN  = 0
	SPI0_MOSI_PIN = 0
	SPI0_MISO_PIN = 0
)
