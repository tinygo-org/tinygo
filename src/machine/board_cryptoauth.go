package machine

const (
	LED Pin = PA02
)

const (
	SPI0_SCK_PIN = PA19
	SPI0_SDO_PIN = PA16
	SPI0_SDI_PIN = PA18
)

const (
	UART_TX_PIN = PA22
	UART_RX_PIN = PA23
)

var (
	DefaultUART = &sercomUSART0
)

var (
	I2C0 = sercomI2CM1
)

const (
	SDA_PIN         = PA08
	SCL_PIN         = PA09
	I2S_SCK_PIN     = PA13
	I2S_WS_PIN      = NoPin
	I2S_SD_PIN      = PA11
	resetMagicValue = 0xf01669ef
)

const (
	USBCDC_DM_PIN           = PA24
	USBCDC_DP_PIN           = PA25
	usb_STRING_PRODUCT      = "CryptoAuth Trust Platform"
	usb_STRING_MANUFACTURER = "Microchip"
)

var (
	usb_VID uint16 = 0x03eb
	usb_PID uint16 = 0x2175
)
