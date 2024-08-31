//go:build rak4631

package machine

const HasLowFrequencyCrystal = true

// Digital Pins
const (
	D0 Pin = P0_28
	D1 Pin = P0_02
)

// Analog pins
const (
	A0 Pin = P0_17
	A1 Pin = P1_02
	A2 Pin = P0_21
)

// Onboard LEDs
const (
	LED  = LED2
	LED1 = P1_03
	LED2 = P1_04
)

// UART pins
const (
	// Default to UART1
	UART_RX_PIN = UART0_RX_PIN
	UART_TX_PIN = UART0_TX_PIN

	// UART1
	UART0_RX_PIN = P0_19
	UART0_TX_PIN = P0_20

	// UART2
	UART1_RX_PIN = P0_15
	UART1_TX_PIN = P0_16
)

// I2C pins
const (
	SDA_PIN = SDA1_PIN
	SCL_PIN = SCL1_PIN

	SDA1_PIN = P0_13
	SCL1_PIN = P0_14

	SDA2_PIN = P0_24
	SCL2_PIN = P0_25
)

// SPI pins
const (
	SPI0_SCK_PIN = P0_03
	SPI0_SDO_PIN = P0_29
	SPI0_SDI_PIN = P0_30
)

// Peripherals
const (
	LORA_NSS    = P1_10
	LORA_SCK    = P1_11
	LORA_MOSI   = P1_12
	LORA_MISO   = P1_13
	LORA_BUSY   = P1_14
	LORA_DIO1   = P1_15
	LORA_NRESET = P1_06
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "WisCore RAK4631 Board"
	usb_STRING_MANUFACTURER = "RAKwireless"
)

var (
	usb_VID uint16 = 0x239a
	usb_PID uint16 = 0x8029
)

var (
	DefaultUART = UART1
)
