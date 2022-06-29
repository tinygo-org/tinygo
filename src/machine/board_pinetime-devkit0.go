//go:build pinetime_devkit0
// +build pinetime_devkit0

package machine

// Board pins for the PineTime.
// Details: https://wiki.pine64.org/index.php/PineTime

// The PineTime has a low-frequency (32kHz) crystal oscillator on board.
const HasLowFrequencyCrystal = true

// LEDs simply expose the three brightness level LEDs on the PineTime. They can
// be useful for simple "hello world" style programs.
const (
	LED1 = LCD_BACKLIGHT_HIGH
	LED2 = LCD_BACKLIGHT_MID
	LED3 = LCD_BACKLIGHT_LOW
	LED  = LED1
)

var DefaultUART = UART0

// UART pins for PineTime. Note that RX is set to NoPin as RXD is not listed in
// the PineTime schematic 1.0:
// http://files.pine64.org/doc/PineTime/PineTime%20Port%20Assignment%20rev1.0.pdf
const (
	UART_TX_PIN Pin = 11 // TP29 (TXD)
	UART_RX_PIN Pin = NoPin
)

// SPI pins for the PineTime.
const (
	SPI0_SCK_PIN Pin = 2
	SPI0_SDO_PIN Pin = 3
	SPI0_SDI_PIN Pin = 4
)

// I2C pins for the PineTime.
const (
	SDA_PIN Pin = 6
	SCL_PIN Pin = 7
)

// Button pins. For some reason, there are two pins for the button.
const (
	BUTTON_IN  Pin = 13
	BUTTON_OUT Pin = 15
)

// Pin for the vibrator.
const VIBRATOR_PIN Pin = 16

// LCD pins, using the naming convention of the official docs:
// http://files.pine64.org/doc/PineTime/PineTime%20Port%20Assignment%20rev1.0.pdf
const (
	LCD_SCK                = SPI0_SCK_PIN
	LCD_SDI                = SPI0_SDO_PIN
	LCD_RS             Pin = 18
	LCD_CS             Pin = 25
	LCD_RESET          Pin = 26
	LCD_BACKLIGHT_LOW  Pin = 14
	LCD_BACKLIGHT_MID  Pin = 22
	LCD_BACKLIGHT_HIGH Pin = 23
)
