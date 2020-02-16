// +build particle_xenon

package machine

const HasLowFrequencyCrystal = true

// More info: https://docs.particle.io/datasheets/discontinued/xenon-datasheet/
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

// Other periferals
const (
	MODE_BUTTON_PIN   = 11
	CHARGE_STATUS_PIN = 41
	LIPO_VOLTAGE_PIN  = 5
	PCB_ANTENNA_PIN   = 24
	EXTERNAL_UFL_PIN  = 25
	NFC1_PIN          = 9
	NFC2_PIN          = 10
)
