//go:build reelboard

package machine

const HasLowFrequencyCrystal = true

// Pins on the reel board
const (
	LED_RED          Pin = 11
	LED_GREEN        Pin = 12
	LED_BLUE         Pin = 41
	LED_YELLOW       Pin = 13
	LED1             Pin = LED_YELLOW
	LED2             Pin = LED_RED
	LED3             Pin = LED_GREEN
	LED4             Pin = LED_BLUE
	LED              Pin = LED1
	EPD_BUSY_PIN     Pin = 14
	EPD_RESET_PIN    Pin = 15
	EPD_DC_PIN       Pin = 16
	EPD_CS_PIN       Pin = 17
	EPD_SCK_PIN      Pin = 19
	EPD_SDO_PIN      Pin = 20
	POWER_SUPPLY_PIN Pin = 32
)

// User "a" button on the reel board
const (
	BUTTON Pin = 7
)

var DefaultUART = UART0

// UART pins
const (
	UART_TX_PIN Pin = 6
	UART_RX_PIN Pin = 8
)

// I2C pins
const (
	SDA_PIN Pin = 26
	SCL_PIN Pin = 27
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = 47
	SPI0_SDO_PIN Pin = 45
	SPI0_SDI_PIN Pin = 46
)

// PowerSupplyActive enables the supply voltages for nRF52840 and peripherals (true) or only for nRF52840 (false)
// This controls the TPS610981 boost converter. You must turn the power supply active in order to use the EPD and
// other onboard peripherals.
func PowerSupplyActive(active bool) {
	POWER_SUPPLY_PIN.Configure(PinConfig{Mode: PinOutput})
	if active {
		POWER_SUPPLY_PIN.High()
	} else {
		POWER_SUPPLY_PIN.Low()
	}
}

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "PHYTEC reelboard"
	usb_STRING_MANUFACTURER = "PHYTEC"
)

var (
	usb_VID uint16 = 0x2FE3
	usb_PID uint16 = 0x100
)
