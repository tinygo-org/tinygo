//go:build blinky2350

// Chip: RP2350A (QFN-60, 30 GPIO) -> target inherits from "rp2350".
// Pin source: https://github.com/pimoroni/blinky2350/blob/main/board/pins.csv
//
// NOTES (verify before first flash):
//   - POWER_EN (GPIO27) likely switches peripheral power.
//     Verify polarity against the schematic; pull HIGH early if needed,
//     otherwise the display will remain blank.
//   - CHARGE_STAT is connected to EXT_GPIO2 (I2C IO expander) according to pins.csv,
//     NOT the RP2350 -> intentionally not defined as a pin here.
//   - The dot matrix is NOT driven via SPI (Latch/Blank/
//     Row-Clock are not SPI signals) -> use PIO or bit-banging.
//     SPI0 and SPI1 are therefore set to NoPin (required by the machine package).

package machine

// Crystal oscillator frequency.
const xoscFreq = 12 // MHz

// User buttons.
const (
	BUTTON_A    Pin = GPIO7
	BUTTON_B    Pin = GPIO9
	BUTTON_C    Pin = GPIO10
	BUTTON_UP   Pin = GPIO11
	BUTTON_DOWN Pin = GPIO6
	BUTTON_HOME Pin = GPIO22
	BUTTON_BOOT Pin = BUTTON_HOME

	BUTTON_RESET Pin = GPIO14
	BUTTON_INT   Pin = GPIO15
)

// Case LEDs (4-zone mono illumination), named CL0..CL3 in the docs.
const (
	CL0 Pin = GPIO0
	CL1 Pin = GPIO1
	CL2 Pin = GPIO2
	CL3 Pin = GPIO3

	LED = CL0
)

// LED dot matrix: column shift register + row driver.
// NOT SPI -> drive via PIO or bit-banging.
const (
	DISPLAY_COL_SCLK  Pin = GPIO16
	DISPLAY_COL_DATA  Pin = GPIO17
	DISPLAY_COL_LATCH Pin = GPIO18
	DISPLAY_COL_BLANK Pin = GPIO19
	DISPLAY_ROW_DATA  Pin = GPIO20
	DISPLAY_ROW_CLK   Pin = GPIO21
)

// I2C0 (Qwiic/STEMMA QT + RTC).
const (
	I2C0_SDA_PIN = GPIO4
	I2C0_SCL_PIN = GPIO5
)

// I2C1 is not routed to any header; NoPin satisfies the machine package.
const (
	I2C1_SDA_PIN Pin = NoPin
	I2C1_SCL_PIN Pin = NoPin
)

// SPI pins - the LED matrix is driven via PIO/bit-banging, not SPI.
// NoPin satisfies the machine package requirements.
const (
	SPI0_SCK_PIN Pin = NoPin
	SPI0_SDO_PIN Pin = NoPin
	SPI0_SDI_PIN Pin = NoPin

	SPI1_SCK_PIN Pin = NoPin
	SPI1_SDO_PIN Pin = NoPin
	SPI1_SDI_PIN Pin = NoPin
)

// UART pins - default RP2350 UART0 mapping.
// Note: GPIO0/GPIO1 are also CL0/CL1 (case LEDs)
// Do not use UART0 and LEDs simultaneously.
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

// Power / sensing.
const (
	VBUS_DETECT Pin = GPIO12
	RTC_ALARM   Pin = GPIO13
	VBAT_SENSE  Pin = GPIO26
	POWER_EN    Pin = GPIO27
	SENSE_1V1   Pin = GPIO28

	BATTERY = VBAT_SENSE
)

var DefaultUART = UART0

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Blinky 2350"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2E8A
	usb_PID uint16 = 0x0005
)
