//go:build pico2_ice

// Most of the info is from
// https://pico2-ice.tinyvision.ai/md_pinout.html although
// (2025-09-07) RP4 appears twice in that pinout - the schematic is
// more clear. Consistent with other RPi boards, we use GPn instead of
// RPn to reference the RPi connected pins.

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
	GP30 Pin = 30
	GP31 Pin = 31
	GP32 Pin = 32
	GP33 Pin = 33
	GP34 Pin = 34
	GP35 Pin = 35
	GP36 Pin = 36
	GP37 Pin = 37
	GP38 Pin = 38
	GP39 Pin = 39
	GP40 Pin = 40
	GP41 Pin = 41
	GP42 Pin = 42
	GP43 Pin = 43
	GP44 Pin = 44
	GP45 Pin = 45
	GP46 Pin = 46
	GP47 Pin = 47

	// RPi pins shared with ICE
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
	ICE_SSN   = ICE16
	ICE_SO    = ICE14
	ICE_SI    = ICE17
	ICE_CK    = ICE15
	FPGA_RSTN = GP31
	ICE_DONE  = GP40
	USB_BOOT  = GP42

	// Button
	SW1     = GP42
	BOOTSEL = GP42

	// Tricolor LEDs
	RED   Pin = GP1
	GREEN Pin = GP0
	BLUE  Pin = GP9

	// Onboard LED
	LED Pin = GREEN

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// This board does not define default i2c pins.
const (
	I2C0_SDA_PIN Pin = 0
	I2C0_SCL_PIN Pin = 0
	I2C1_SDA_PIN Pin = 0
	I2C1_SCL_PIN Pin = 0
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
