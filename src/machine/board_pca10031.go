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
	LED       = LED_RED
	LED1      = LED_RED
	LED2      = LED_GREEN
	LED3      = LED_BLUE
	LED_RED   = P0_21
	LED_GREEN = P0_22
	LED_BLUE  = P0_23
)

var DefaultUART = UART0

// UART pins
const (
	UART_TX_PIN = P0_09
	UART_RX_PIN = P0_11
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
