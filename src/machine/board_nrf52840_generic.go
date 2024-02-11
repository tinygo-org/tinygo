//go:build nrf52840 && nrf52840_generic

package machine

var (
	LED          = NoPin
	SDA_PIN      = NoPin
	SCL_PIN      = NoPin
	UART_TX_PIN  = NoPin
	UART_RX_PIN  = NoPin
	SPI0_SCK_PIN = NoPin
	SPI0_SDO_PIN = NoPin
	SPI0_SDI_PIN = NoPin

	// https://pid.codes/org/TinyGo/
	usb_VID uint16 = 0x1209
	usb_PID uint16 = 0x9090

	usb_STRING_MANUFACTURER = "TinyGo"
	usb_STRING_PRODUCT      = "nRF52840 Generic board"
)

var (
	DefaultUART = UART0
)
