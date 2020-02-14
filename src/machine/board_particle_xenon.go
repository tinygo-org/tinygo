// +build particle_xenon

package machine

const HasLowFrequencyCrystal = true

// LEDs on the XENON
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

// Other periferals
const (
	MODE_BUTTON   = 11
	CHARGE_STATUS = 41
	LIPO_VOLTAGE  = 5
	PCB_ANTENNA   = 24
	EXTERNAL_UFL  = 25
	NFC1          = 9
	NFC2          = 10
)