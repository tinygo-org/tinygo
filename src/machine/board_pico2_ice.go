//go:build pico2_ice

// Most of the info is from
// https://pico2-ice.tinyvision.ai/md_pinout.html although
// (2025-09-07) RP4 appears twice in that pinout - the schematic is
// more clear. Consistent with other RPi boards, we use GPn instead of
// RPn to reference the RPi connected pins.

package machine

// GPIO pins
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
	GP23 = GPIO23
	GP24 = GPIO24
	GP25 = GPIO25
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

	// RPi pins shared with ICE. The ICE number is what appears on
	// the board silkscreen.
	ICE9  = GP28
	ICE11 = GP29
	ICE14 = GP7
	ICE15 = GP6
	ICE16 = GP5
	ICE17 = GP4
	ICE18 = GP27
	ICE19 = GP23
	ICE20 = GP22
	ICE21 = GP26
	ICE23 = GP25
	ICE25 = GP30
	ICE26 = GP24
	ICE27 = GP20

	// FPGA Clock pin.
	ICE35_G0 = GP21

	// Silkscreen & Pinout names
	ICE_SSN = ICE16
	ICE_SO  = ICE14
	ICE_SI  = ICE17
	ICE_CK  = ICE15
	SD      = GP2
	SC      = GP3

	FPGA_RSTN = GP31
	A3        = GP32
	A1        = GP33
	A4        = GP34
	A2        = GP35
	B3        = GP36
	B1        = GP37
	B4        = GP38
	B2        = GP39
	N0        = GP40 // On the board these are labeled "~0"
	N1        = GP41
	N2        = GP42
	N3        = GP43
	N4        = GP44
	N5        = GP45
	N6        = GP46

	// Functions from Schematic.
	ICE_DONE = GP40
	USB_BOOT = GP42

	// Button
	SW1     = GP42
	BOOTSEL = GP42

	// Tricolor LEDs
	LED_RED   = GP1
	LED_GREEN = GP0
	LED_BLUE  = GP9

	// Onboard LED
	LED = LED_GREEN

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// This board does not define default i2c pins.
const (
	I2C0_SDA_PIN = NoPin
	I2C0_SCL_PIN = NoPin
	I2C1_SDA_PIN = NoPin
	I2C1_SCL_PIN = NoPin
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
	SPI1_SCK_PIN = GPIO10
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO11 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO12 // Rx
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
	usb_STRING_PRODUCT      = "Pico2"
	usb_STRING_MANUFACTURER = "Raspberry Pi"
)

var (
	usb_VID uint16 = 0x2E8A
	usb_PID uint16 = 0x000A
)
