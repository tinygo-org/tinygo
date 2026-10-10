//go:build feather_esp32s3_tft

// Silkscreen for the Adafruit ESP32-S3 TFT Feather
// https://learn.adafruit.com/adafruit-esp32-s3-tft-feather/pinouts

package machine

const (
	SCL_PIN = GPIO41
	SDA_PIN = GPIO42

	SPI1_SCK_PIN  = GPIO36 // SCK
	SPI1_MOSI_PIN = GPIO35 // SDO (MOSI)
	SPI1_MISO_PIN = GPIO37 // SDI (MISO)
	SPI1_CS_PIN   = NoPin  // CS

	SPI2_SCK_PIN  = NoPin // SCK
	SPI2_MOSI_PIN = NoPin // SDO (MOSI)
	SPI2_MISO_PIN = NoPin // SDI (MISO)
	SPI2_CS_PIN   = NoPin // CS
)

// TFT pins
const (
	TFT_I2C_POWER = GPIO21
	TFT_CS        = GPIO7
	TFT_DC        = GPIO39
	TFT_RESET     = GPIO40
	TFT_BACKLIGHT = GPIO45
)

// Neopixel pins
const (
	NEOPIXEL       = GPIO33
	NEOPIXEL_POWER = GPIO34
)

// UART pins
const (
	RX = GPIO2
	TX = GPIO1
)

const (
	BUTTON = GPIO0
)

// Analog pins
const (
	A0 = GPIO18
	A1 = GPIO17
	A2 = GPIO16
	A3 = GPIO15
	A4 = GPIO14
	A5 = GPIO8
)

// Digital pins
const (
	D5  = GPIO5
	D6  = GPIO6
	D9  = GPIO9
	D10 = GPIO10
	D11 = GPIO11
	D12 = GPIO12
	D13 = GPIO13
)
