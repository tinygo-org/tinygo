//go:build waveshare_esp32c6_lcd_1_47

// This file contains the pin mappings for the Waveshare ESP32-C6-LCD-1.47 board.
//
// The board has an ESP32-C6FH8 with 8 MB flash, a 1.47 inch 172x320 ST7789
// LCD, a microSD card slot that shares the LCD SPI bus, and a WS2812B LED.
//
// - https://docs.waveshare.com/ESP32-C6-LCD-1.47
// - https://files.waveshare.com/wiki/ESP32-C6-LCD-1.47/ESP32-C6-LCD-1.47_schemetics.pdf

package machine

// Header pins
const (
	D0  = GPIO0
	D1  = GPIO1
	D2  = GPIO2
	D3  = GPIO3
	D4  = GPIO4
	D5  = GPIO5
	D9  = GPIO9
	D16 = GPIO16
	D17 = GPIO17
	D18 = GPIO18
	D19 = GPIO19
	D20 = GPIO20
	D23 = GPIO23
)

// Analog pins
const (
	A0 = GPIO0
	A1 = GPIO1
	A2 = GPIO2
	A3 = GPIO3
	A4 = GPIO4
	A5 = GPIO5
)

// Button pin
const (
	BUTTON      = BUTTON_BOOT
	BUTTON_BOOT = GPIO9
)

// Onboard LEDs
const (
	WS2812   = GPIO8
	NEOPIXEL = WS2812
)

// UART pins
const (
	UART0_TX_PIN = GPIO16
	UART0_RX_PIN = GPIO17
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

// I2C pins
const (
	SDA_PIN = GPIO19
	SCL_PIN = GPIO20
)

// SPI pins
const (
	SPI_SCK_PIN = GPIO7
	SPI_SDO_PIN = GPIO6
	SPI_SDI_PIN = GPIO5
)

// LCD pins
const (
	LCD_SCK_PIN = SPI_SCK_PIN
	LCD_SDO_PIN = SPI_SDO_PIN
	LCD_CS_PIN  = GPIO14
	LCD_DC_PIN  = GPIO15
	LCD_RST_PIN = GPIO21
	LCD_BL_PIN  = GPIO22
)

// MicroSD pins
const (
	SD_SCK_PIN = SPI_SCK_PIN
	SD_SDO_PIN = SPI_SDO_PIN
	SD_SDI_PIN = SPI_SDI_PIN
	SD_CS_PIN  = GPIO4
)
