//go:build pca10031
// +build pca10031

// pca10031 is a nrf51 based dongle, intended for use in wireless applications.
//
// https://infocenter.nordicsemi.com/pdf/nRF51_Dongle_UG_v1.0.pdf
package machine

// The pca10031 has a 32kHz crystal on board.
const HasLowFrequencyCrystal = true

// LED on the pca10031
const (
	LED       Pin = LED_RED
	LED1      Pin = LED_RED
	LED2      Pin = LED_GREEN
	LED3      Pin = LED_BLUE
	LED_RED   Pin = 21
	LED_GREEN Pin = 22
	LED_BLUE  Pin = 23
)

var DefaultUART = UART0

// UART pins
const (
	UART_TX_PIN Pin = 9
	UART_RX_PIN Pin = 11
)

// I2C pins (disabled)
const (
	SDA_PIN = NoPin
	SCL_PIN = NoPin
)

// SPI pins (unused)
const (
	SPI0_SCK_PIN = NoPin
	SPI0_SDO_PIN = NoPin
	SPI0_SDI_PIN = NoPin
)
