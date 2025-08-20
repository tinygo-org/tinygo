//go:build pga2350

package machine

// PGA2350 pin definitions.
const (
	GP0  = GPIO0
	GP1  = GPIO1
	GP2  = GPIO2
	GP3  = GPIO3
	GP4  = GPIO4
	GP5  = GPIO5
	GP6  = GPIO6
	GP7  = GPIO7
	GP8  = GPIO8
	GP9  = GPIO9
	GP10 = GPIO10
	GP11 = GPIO11
	GP12 = GPIO12
	GP13 = GPIO13
	GP14 = GPIO14
	GP15 = GPIO15
	GP16 = GPIO16
	GP17 = GPIO17
	GP18 = GPIO18
	GP19 = GPIO19
	GP20 = GPIO20
	GP21 = GPIO21
	GP22 = GPIO22
	GP26 = GPIO26
	GP27 = GPIO27
	GP28 = GPIO28
	GP29 = GPIO29
	GP30 = GPIO30
	GP31 = GPIO31
	GP32 = GPIO32
	GP33 = GPIO33
	GP34 = GPIO34
	GP35 = GPIO35
	GP36 = GPIO36
	GP37 = GPIO37
	GP38 = GPIO38
	GP39 = GPIO39
	GP40 = GPIO40
	GP41 = GPIO41
	GP42 = GPIO42
	GP43 = GPIO43
	GP44 = GPIO44
	GP45 = GPIO45
	GP46 = GPIO46
	GP47 = GPIO47
)

var DefaultUART = UART0

// Peripheral defaults.
const (
	xoscFreq = 12 // MHz

	I2C0_SDA_PIN = GP4
	I2C0_SCL_PIN = GP5

	I2C1_SDA_PIN = GP2
	I2C1_SCL_PIN = GP3

	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO18
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO19 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO16 // Rx

	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO10
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO11 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO12 // Rx

	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART1_TX_PIN = GPIO8
	UART1_RX_PIN = GPIO9
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

// USB identifiers
const (
	usb_STRING_PRODUCT      = "PGA2350"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2E8A
	usb_PID uint16 = 0x000A
)
