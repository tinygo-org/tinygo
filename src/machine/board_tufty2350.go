//go:build tufty2350

// This contains the pin mappings for the Pimoroni Tufty 2350 board.
//
// For more information, see: https://shop.pimoroni.com/products/tufty-2350
// Pinout source: https://github.com/pimoroni/tufty2350/blob/main/board/pins.csv
package machine

// Case LEDs (4-zone mono illumination), named CL0..CL3 in the docs.
const (
	CL0 Pin = GPIO0
	CL1 Pin = GPIO1
	CL2 Pin = GPIO2
	CL3 Pin = GPIO3

	// Convention: first zone as default LED.
	LED = CL0
)

const (
	BUTTON_DOWN  Pin = GPIO6
	BUTTON_A     Pin = GPIO7
	BUTTON_B     Pin = GPIO9
	BUTTON_C     Pin = GPIO10
	BUTTON_UP    Pin = GPIO11
	BUTTON_HOME  Pin = GPIO22
	BUTTON_RESET Pin = GPIO14
	BUTTON_INT   Pin = GPIO15

	VBUS_DETECT Pin = GPIO12
	RTC_ALARM   Pin = GPIO13

	LCD_BACKLIGHT Pin = GPIO26
	LCD_CS        Pin = GPIO27
	LCD_DC        Pin = GPIO28
	LCD_WR        Pin = GPIO30
	LCD_RD        Pin = GPIO31
	LCD_DB0       Pin = GPIO32
	LCD_DB1       Pin = GPIO33
	LCD_DB2       Pin = GPIO34
	LCD_DB3       Pin = GPIO35
	LCD_DB4       Pin = GPIO36
	LCD_DB5       Pin = GPIO37
	LCD_DB6       Pin = GPIO38
	LCD_DB7       Pin = GPIO39

	VBAT_SENSE  Pin = GPIO40
	POWER_EN    Pin = GPIO41
	SENSE_1V1   Pin = GPIO42
	LIGHT_SENSE Pin = GPIO43
)

// CYW43439 wireless chip control pins.
const (
	WL_REG_ON Pin = GPIO23
	WL_DATA   Pin = GPIO24
	WL_CLOCK  Pin = GPIO29
	WL_CS     Pin = GPIO25
)

// I2C pins. I2C0 carries the onboard PCF85063 real time clock.
const (
	I2C0_SDA_PIN Pin = GPIO4
	I2C0_SCL_PIN Pin = GPIO5

	I2C1_SDA_PIN Pin = NoPin
	I2C1_SCL_PIN Pin = NoPin
)

// SPI pins. No SPI peripheral is wired to a header on this board.
const (
	SPI0_SCK_PIN Pin = NoPin
	SPI0_SDO_PIN Pin = NoPin
	SPI0_SDI_PIN Pin = NoPin

	SPI1_SCK_PIN Pin = NoPin
	SPI1_SDO_PIN Pin = NoPin
	SPI1_SDI_PIN Pin = NoPin
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Tufty 2350"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x1101
)

// UART pins.
// Note: GPIO0/GPIO1 are also CL0/CL1 (case LEDs).
// Do not use UART0 and the case LEDs at the same time.
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0
