//go:build esp32s3_generic

package machine

const (
	SCL_PIN = NoPin
	SDA_PIN = NoPin

	SPI1_SCK_PIN  = NoPin // SCK
	SPI1_MOSI_PIN = NoPin // SDO (MOSI)
	SPI1_MISO_PIN = NoPin // SDI (MISO)
	SPI1_CS_PIN   = NoPin // CS

	SPI2_SCK_PIN  = NoPin // SCK
	SPI2_MOSI_PIN = NoPin // SDO (MOSI)
	SPI2_MISO_PIN = NoPin // SDI (MISO)
	SPI2_CS_PIN   = NoPin // CS
)
