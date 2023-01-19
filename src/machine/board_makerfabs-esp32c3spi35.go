//go:build makerfabs_esp32c3spi35

// This file contains the pin mappings for the Makerfabs ESP32C3SPI35 board.
//
// The Makerfabs ESP32C3SPI35 is an LCD Touchscreen development board powered
// by the Espressif ESP32-C3 SoC featuring an open-source RISC-V architecture.
//
// Specifications:
//   SoC: ESP32-C3-MINI-1-N4, 4MB Flash, RISCV-32bit, 160MHz, 400KB SRAM
//   Wireless: WiFi & Bluetooth 5.0 (BLE)
//   LCD: 3.5inch TFT LCD (480x320)
//   LCD Driver: ILI9488 SPI
//   Touch Panel: Capacitive
//   Touch Panel Driver: FT6236
//   MicroSD Card Slot
//   Mabee Interface
//   Dual USB Type-C (one for USB-to-UART and one for native USB)
//
// Website:   https://www.makerfabs.com/ep32-c3-risc-v-spi-tft-touch.html
// Wiki:      https://wiki.makerfabs.com/ESP32_C3_SPI_3.5_TFT_with_Touch.html
// GitHub:    https://github.com/Makerfabs/Makerfabs-ESP32-C3-SPI-TFT-with-Touch
// Schematic: https://github.com/Makerfabs/Makerfabs-ESP32-C3-SPI-TFT-with-Touch/raw/main/Hardware/ESP32-C3%20TFT%20Touch%20v1.1(3.5''%20ili9488).PDF
// Datasheet: https://www.espressif.com/sites/default/files/documentation/esp32-c3-mini-1_datasheet_en.pdf

package machine

// Digital pins
const (
	//    Pin    // Function
	//    -----  // ---------------------
	D0  = GPIO0  // Touchscreen CS
	D1  = GPIO1  // MicroSD CS
	D2  = GPIO2  // I2C SDA
	D3  = GPIO3  // I2C SCL
	D4  = GPIO4  // SPI CS
	D5  = GPIO5  // SPI SCK
	D6  = GPIO6  // SPI SDO
	D7  = GPIO7  // SPI SDI
	D8  = GPIO8  // Touchscreen Backlight
	D9  = GPIO9  // Boot Button
	D10 = GPIO10 // TFT D/C
	D18 = GPIO18 // USB DM
	D19 = GPIO19 // USB DP
	D20 = GPIO20 // UART RX
	D21 = GPIO21 // UART TX
)

// Button pin
const (
	BUTTON      = BUTTON_BOOT
	BUTTON_BOOT = D9
)

// TFT pins
const (
	TFT_BL_PIN  = D8
	TFT_CS_PIN  = SPI_CS_PIN
	TFT_DC_PIN  = D10
	TFT_SCK_PIN = SPI_SCK_PIN
	TFT_SDI_PIN = SPI_SDI_PIN
	TFT_SDO_PIN = SPI_SDO_PIN
)

// Touchscreen pins
const (
	TS_CS_PIN  = D0
	TS_SDA_PIN = I2C_SDA_PIN
	TS_SCL_PIN = I2C_SCL_PIN
)

// MicroSD pins
const (
	SD_CS_PIN  = D1
	SD_SCK_PIN = SPI_SCK_PIN
	SD_SDI_PIN = SPI_SDI_PIN
	SD_SDO_PIN = SPI_SDO_PIN
)

// USBCDC pins
const (
	USBCDC_DM_PIN = D18
	USBCDC_DP_PIN = D19
)

// UART pins
const (
	UART_RX_PIN = D20
	UART_TX_PIN = D21
)

// I2C pins
const (
	I2C_SDA_PIN = D2
	I2C_SCL_PIN = D3
)

// SPI pins
const (
	SPI_CS_PIN  = D4
	SPI_SCK_PIN = D5
	SPI_SDI_PIN = D7
	SPI_SDO_PIN = D6
)
