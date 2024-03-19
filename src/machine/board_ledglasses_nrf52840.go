//go:build ledglasses_nrf52840

package machine

// Adafruit LED Glasses Driver - nRF52840 Sensor Board
// https://www.adafruit.com/product/5217
// ----------------------------------------------------------------------------
// References:
// https://cdn-learn.adafruit.com/assets/assets/000/105/134/original/adafruit_products_EyeLightsD_sch.png?1633457726
// https://github.com/adafruit/Adafruit_nRF52_Arduino/blob/master/variants/ledglasses_nrf52840/variant.h
// https://github.com/adafruit/Adafruit-EyeLights-LED-Glasses-and-Driver-PCB/blob/main/Adafruit%20EyeLights%20LED%20Glasses%20Driver.pdf

const HasLowFrequencyCrystal = false

// GPIO Pins
const (
	D2  = P0_31 // LED
	D3  = P1_15 // NEOPIXEL
	D4  = P0_30 // SWITCH
	D5  = P1_08 // MICROPHONE_DATA
	D6  = P0_07 // MICROPHONE_CLOCK
	D8  = P0_08 // SCL
	D9  = P0_06 // SDA
	D20 = P0_04 // Battery
	D22 = P0_19 // QSPI CLK
	D23 = P0_20 // QSPI CS
	D24 = P1_00 // QSPI Data 0
	D25 = P0_21 // QSPI Data 1
	D26 = P0_22 // QSPI Data 2
	D27 = P0_23 // QSPI Data 3
)

// Analog Pins
const (
	A6 = D20 // Battery
)

const (
	LED      = D2
	LED1     = LED
	NEOPIXEL = D3
	WS2812   = NEOPIXEL
	SWITCH   = D4
	BUTTON   = SWITCH

	QSPI_SCK   = D22
	QSPI_CS    = D23
	QSPI_DATA0 = D24
	QSPI_DATA1 = D25
	QSPI_DATA2 = D26
	QSPI_DATA3 = D27
)

// I2C pins
const (
	SDA_PIN = D9 // I2C0 external
	SCL_PIN = D8 // I2C0 external
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "LED Glasses Driver nRF52840"
	usb_STRING_MANUFACTURER = "Adafruit Industries LLC"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x810E
)

// LED Glasses Driver does not have pins broken out for the peripherals below,
// however the machine_nrf*.go implementations of I2C/SPI/etc expect the pin
// constants to be defined, so we are defining them all as NoPin
const (
	UART_TX_PIN  = NoPin
	UART_RX_PIN  = NoPin
	SPI0_SCK_PIN = NoPin
	SPI0_SDO_PIN = NoPin
	SPI0_SDI_PIN = NoPin
)
