//go:build metro_rp2350

package machine

// GPIO pins
const (
	GP0  Pin = GPIO0
	GP1  Pin = GPIO1
	GP2  Pin = GPIO2
	GP3  Pin = GPIO3
	GP4  Pin = GPIO4
	GP5  Pin = GPIO5
	GP6  Pin = GPIO6
	GP7  Pin = GPIO7
	GP8  Pin = GPIO8
	GP9  Pin = GPIO9
	GP10 Pin = GPIO10
	GP11 Pin = GPIO11
	GP12 Pin = GPIO12
	GP13 Pin = GPIO13
	GP14 Pin = GPIO14
	GP15 Pin = GPIO15
	GP16 Pin = GPIO16
	GP17 Pin = GPIO17
	GP18 Pin = GPIO18
	GP19 Pin = GPIO19
	GP20 Pin = GPIO20
	GP21 Pin = GPIO21
	GP22 Pin = GPIO22
	GP23 Pin = GPIO23
	GP24 Pin = GPIO24
	GP25 Pin = GPIO25
	GP26 Pin = GPIO26
	GP27 Pin = GPIO27
	GP28 Pin = GPIO28
	GP29 Pin = GPIO29
	GP30 Pin = GPIO30
	GP31 Pin = GPIO31
	GP32 Pin = GPIO32
	GP33 Pin = GPIO33
	GP34 Pin = GPIO34
	GP35 Pin = GPIO35
	GP36 Pin = GPIO36
	GP37 Pin = GPIO37
	GP38 Pin = GPIO38
	GP39 Pin = GPIO39
	GP40 Pin = GPIO40
	GP41 Pin = GPIO41
	GP42 Pin = GPIO42
	GP43 Pin = GPIO43
	GP44 Pin = GPIO44
	GP45 Pin = GPIO45
	GP46 Pin = GPIO46

	// Boot button
	BUTTON Pin = GPIO24

	// Onboard LED
	LED Pin = GPIO23

	// Onboard NeoPixel
	NEOPIXEL Pin = GPIO25
	WS2812   Pin = GPIO25

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// Arduino-header digital pins
const (
	RX  Pin = GPIO1
	TX  Pin = GPIO0
	D2  Pin = GPIO2
	D3  Pin = GPIO3
	D4  Pin = GPIO4
	D5  Pin = GPIO5
	D6  Pin = GPIO6
	D7  Pin = GPIO7
	D8  Pin = GPIO8
	D9  Pin = GPIO9
	D10 Pin = GPIO10
	D11 Pin = GPIO11
	D22 Pin = GPIO22
	D23 Pin = GPIO23
)

// Arduino-header analog pins
const (
	A0 Pin = GPIO41
	A1 Pin = GPIO42
	A2 Pin = GPIO43
	A3 Pin = GPIO44
	A4 Pin = GPIO45
	A5 Pin = GPIO46
)

// I2C Default pins on Raspberry Pico.
const (
	I2C0_SDA_PIN = GP20
	I2C0_SCL_PIN = GP21

	I2C1_SDA_PIN = GP2
	I2C1_SCL_PIN = GP3
)

// SPI default pins
const (
	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO18
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO19 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO16 // Rx

	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO30
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO31 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO28 // Rx

	// SPI header pins
	MOSI Pin = SPI1_SDO_PIN
	MISO Pin = SPI1_SDI_PIN
	SCK  Pin = SPI1_SCK_PIN
)

// SD card reader pins
const (
	SD_SCK         = GPIO34
	SD_MOSI        = GPIO35
	SD_MISO        = GPIO36
	SDIO_DATA1     = GPIO37
	SDIO_DATA2     = GPIO38
	SD_CS          = GPIO39
	SD_CARD_DETECT = GPIO40
)

// HSTX pins
const (
	CKN Pin = GPIO15
	CKP Pin = GPIO14
	D0N Pin = GPIO19
	D0P Pin = GPIO18
	D1N Pin = GPIO17
	D1P Pin = GPIO16
	D2N Pin = GPIO13
	D2P Pin = GPIO12
	D26 Pin = GPIO26
	D27 Pin = GPIO27
	SCL Pin = GPIO21
	SDA Pin = GPIO20
)

// USB host header pins
const (
	USB_HOST_DATA_PLUS  Pin = GPIO32
	USB_HOST_DATA_MINUS Pin = GPIO33
	USB_HOST_5V_POWER   Pin = GPIO29
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART1_TX_PIN = GPIO8
	UART1_RX_PIN = GPIO9
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Metro RP2350"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x814E
)
