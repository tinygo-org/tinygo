//go:build xiao_ble
// +build xiao_ble

// This file contains the pin mappings for the Seeed XIAO BLE nRF52840 [Sense] boards.
//
// Seeed XIAO BLE is an ultra-small size, ultra-low power Bluetooth development board based on the Nordic nRF52840.
// It features an onboard Bluetooth antenna, onboard battery charging chip, and 21*17.5mm thumb size, which makes it ideal for IoT projects.
//
// Seeed XIAO BLE nRF52840 Sense is a tiny Bluetooth LE development board designed for IoT and AI applications.
// It features an onboard antenna, 6 Dof IMU, microphone, all of which make it an ideal board to run AI using TinyML and TensorFlow Lite.
//
// SoftDevice (s140v7) is pre-flashed on this board already.
// See https://github.com/tinygo-org/bluetooth
//
// - https://www.seeedstudio.com/Seeed-XIAO-BLE-nRF52840-p-5201.html
// - https://www.seeedstudio.com/Seeed-XIAO-BLE-Sense-nRF52840-p-5253.html
//
// - https://wiki.seeedstudio.com/XIAO_BLE/
// - https://github.com/Seeed-Studio/ArduinoCore-mbed/tree/master/variants/SEEED_XIAO_NRF52840_SENSE
//
package machine

const HasLowFrequencyCrystal = true

// Digital Pins
const (
	D0  Pin = P0_02
	D1  Pin = P0_03
	D2  Pin = P0_28
	D3  Pin = P0_29
	D4  Pin = P0_04
	D5  Pin = P0_05
	D6  Pin = P1_11
	D7  Pin = P1_12
	D8  Pin = P1_13
	D9  Pin = P1_14
	D10 Pin = P1_15
)

// Analog pins
const (
	A0 Pin = P0_02
	A1 Pin = P0_03
	A2 Pin = P0_28
	A3 Pin = P0_29
	A4 Pin = P0_04
	A5 Pin = P0_05
)

// Onboard LEDs
const (
	LED       = LED_CHG
	LED1      = LED_RED
	LED2      = LED_GREEN
	LED3      = LED_BLUE
	LED_CHG   = P0_17
	LED_RED   = P0_26
	LED_GREEN = P0_30
	LED_BLUE  = P0_06
)

// UART0 pins
const (
	UART_RX_PIN = P1_12
	UART_TX_PIN = P1_11
)

// I2C pins
const (
	// Defaults to internal
	SDA_PIN = SDA1_PIN
	SCL_PIN = SCL1_PIN

	// I2C0 (external) pins
	SDA0_PIN = P0_04
	SCL0_PIN = P0_05

	// I2C1 (internal) pins
	SDA1_PIN = P0_07
	SCL1_PIN = P0_27
)

// SPI pins
const (
	SPI0_SCK_PIN = P1_13
	SPI0_SDO_PIN = P1_14
	SPI0_SDI_PIN = P1_15
)

// Peripherals
const (
	LSM_PWR = P1_08 // IMU (LSM6DS3TR) power
	LSM_INT = P0_11 // IMU (LSM6DS3TR) interrupt

	MIC_PWR = P1_10 // Microphone (MSM261D3526H1CPM) power
	MIC_CLK = P1_00
	MIC_DIN = P0_16
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "XIAO nRF52840 Sense"
	usb_STRING_MANUFACTURER = "Seeed"
)

var (
	usb_VID uint16 = 0x2886
	usb_PID uint16 = 0x0045
)

var (
	DefaultUART = UART0
)
