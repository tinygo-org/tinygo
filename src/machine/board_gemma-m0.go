//go:build sam && atsamd21 && gemma_m0

package machine

// Used to reset into bootloader.
const resetMagicValue = 0xf01669ef

// GPIO Pins.
const (
	D0  = PA04 // SERCOM0/PAD[0]
	D1  = PA02
	D2  = PA05 // SERCOM0/PAD[1]
	D3  = PA00 // DotStar LED: SERCOM1/PAD[0]: APA102/MOSI
	D4  = PA01 // DotStar LED: SERCOM1/PAD[1]: APA102/SCK
	D11 = PA30 // Flash Access: SERCOM1/PAD[2]
	D12 = PA31 // Flash Access: SERCOM1/PAD[3]
	D13 = PA23 // LED: SERCOM3/PAD[1] SERCOM5/PAD[1]
)

// Analog pins.
const (
	A0 = D1
	A1 = D2
	A2 = D0
)

const (
	LED = PA23
)

// USBCDC pins.
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART0 pins.
const (
	UART_TX_PIN = PA04 // TX: SERCOM0/PAD[0]
	UART_RX_PIN = PA05 // RX: SERCOM0/PAD[1]
)

// UART0s on the Gemma M0.
var UART0 = &sercomUSART0

// SPI pins.
const (
	SPI0_SCK_PIN = PA05 // SCK: SERCOM0/PAD[1]
	SPI0_SDO_PIN = PA04 // MOSI: SERCOM0/PAD[0]
	SPI0_SDI_PIN = NoPin
	SPI0_CS_PIN  = NoPin
)

// SPI on the Gemma M0.
var SPI0 = sercomSPIM0

// SPI pins for DotStar LED (using APA102 software SPI) and Flash.
const (
	SPI1_SCK_PIN = PA01 // SCK: SERCOM1/PAD[0]
	SPI1_SDO_PIN = PA00 // MOSI: SERCOM1/PAD[1]
	SPI1_SDI_PIN = PA31 // MISO: SERCOM1/PAD[3]
	SPI1_CS_PIN  = PA30 // CS: SERCOM1/PAD[2]
)

// I2C pins.
const (
	SDA_PIN = PA04 // SDA: SERCOM0/PAD[0]
	SCL_PIN = PA05 // SCL: SERCOM0/PAD[1]
)

// I2C on the Gemma M0.
var (
	I2C0 = sercomI2CM0
)

// I2S (not connected, needed for atsamd21).
const (
	I2S_SCK_PIN = NoPin
	I2S_SD_PIN  = NoPin
	I2S_WS_PIN  = NoPin
)

// USB CDC identifiers.
const (
	usb_STRING_PRODUCT      = "Adafruit Gemma M0"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x801E
)

var (
	DefaultUART = UART0
)
