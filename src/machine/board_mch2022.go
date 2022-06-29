//go:build mch2022
// +build mch2022

package machine

// See: https://badge.team/docs/badges/mch2022/pinout/

const (
	UART_TX_PIN Pin = 1
	UART_RX_PIN Pin = 3

	WS2812 Pin = 5

	PowerOn Pin = 19 // Set high to enable power to LEDs and SD card

	// I2C pins
	SDA_PIN Pin = 22
	SCL_PIN Pin = 21

	// SPI and related pins (ICE40 and LCD).
	LCD_RESET         Pin = 25
	LCD_MODE          Pin = 26
	LCD_DC            Pin = 33
	SPI0_SCK_PIN      Pin = 18
	SPI0_SDO_PIN      Pin = 23
	SPI0_SDI_PIN      Pin = 35 // connected to ICE40
	SPI0_CS_ICE40_PIN Pin = 27
	SPI0_CS_LCD_PIN   Pin = 32
)
