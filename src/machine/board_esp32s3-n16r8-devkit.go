//go:build esp32s3_n16r8_devkit

// This file contains the pin mappings for ESP32-S3-DevKitC-1 style boards
// with an ESP32-S3-WROOM-1-N16R8 module.
//
// - https://docs.espressif.com/projects/esp-dev-kits/en/latest/esp32s3/esp32-s3-devkitc-1/user_guide_v1.0.html

package machine

// Header pins. GPIO35 to GPIO37 connect to the octal PSRAM in the module.
const (
	IO0  = GPIO0
	IO1  = GPIO1
	IO2  = GPIO2
	IO3  = GPIO3
	IO4  = GPIO4
	IO5  = GPIO5
	IO6  = GPIO6
	IO7  = GPIO7
	IO8  = GPIO8
	IO9  = GPIO9
	IO10 = GPIO10
	IO11 = GPIO11
	IO12 = GPIO12
	IO13 = GPIO13
	IO14 = GPIO14
	IO15 = GPIO15
	IO16 = GPIO16
	IO17 = GPIO17
	IO18 = GPIO18
	IO19 = GPIO19
	IO20 = GPIO20
	IO21 = GPIO21
	IO38 = GPIO38
	IO39 = GPIO39
	IO40 = GPIO40
	IO41 = GPIO41
	IO42 = GPIO42
	IO43 = GPIO43
	IO44 = GPIO44
	IO45 = GPIO45
	IO46 = GPIO46
	IO47 = GPIO47
	IO48 = GPIO48
)

// Boot button
const BUTTON = GPIO0

// RGB LED, connected when the RGB solder jumper is bridged.
const WS2812 = GPIO48

// UART0 pins, connected to the USB to UART bridge.
const (
	UART_TX_PIN = GPIO43
	UART_RX_PIN = GPIO44
)

const (
	SCL_PIN = NoPin
	SDA_PIN = NoPin

	SPI1_SCK_PIN  = NoPin
	SPI1_MOSI_PIN = NoPin
	SPI1_MISO_PIN = NoPin
	SPI1_CS_PIN   = NoPin

	SPI2_SCK_PIN  = NoPin
	SPI2_MOSI_PIN = NoPin
	SPI2_MISO_PIN = NoPin
	SPI2_CS_PIN   = NoPin
)
