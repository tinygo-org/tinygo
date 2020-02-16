// +build particle_boron

package machine

const HasLowFrequencyCrystal = true

// More info: https://docs.particle.io/datasheets/cellular/boron-datasheet/
// Board diagram: https://docs.particle.io/assets/images/xenon/xenon-block-diagram.png

// LEDs
const (
	LED       Pin = 44
	LED_GREEN Pin = 14
	LED_RED   Pin = 13
	LED_BLUE  Pin = 15
)

// UART pins
const (
	UART_TX_PIN Pin = 6
	UART_RX_PIN Pin = 8
)

// I2C pins
const (
	SDA_PIN = 26
	SCL_PIN = 27

	// Internal I2C with MAX17043 and BQ24195 chips on it
	SDA1_PIN = 24
	SCL1_PIN = 41
	INT1_PIN = 5
)

// SPI pins
const (
	SPI0_SCK_PIN  = 47
	SPI0_MOSI_PIN = 45
	SPI0_MISO_PIN = 46
)

// Internal 4MB SPI Flash
const (
	SPI1_SCK_PIN  = 19
	SPI1_MOSI_PIN = 20
	SPI1_MISO_PIN = 21
	SPI1_CS_PIN   = 17
	SPI1_WP_PIN   = 22
	SPI1_HOLD_PIN = 23
)

// u-blox coprocessor
const (
	UBLOX_TXD_PIN       = 37
	UBLOX_RXD_PIN       = 36
	UBLOX_CTS_PIN       = 38
	UBLOX_RTS_PIN       = 39
	UBLOX_RESET_PIN     = 16
	UBLOX_POWER_ON_PIN  = 24
	UBLOX_BUFF_EN_PIN   = 25
	UBLOX_VINT_PIN      = 2
)

// Other periferals
const (
	MODE_BUTTON_PIN   = 11
	CHARGE_STATUS_PIN = 41
	LIPO_VOLTAGE_PIN  = 5
	PCB_ANTENNA_PIN   = 2
	EXTERNAL_UFL_PIN  = 25
	NFC1_PIN          = 9
	NFC2_PIN          = 10
)