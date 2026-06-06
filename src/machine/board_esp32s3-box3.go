//go:build esp32s3_box3

// This file contains the pin mappings for the ESP32-S3-BOX-3 board.
//
// ESP32-S3-BOX-3 is an AI voice development kit with 2.4" display.
// - https://github.com/espressif/esp-box
//
// Based on ESP-BOX-3 BSP from Espressif.

package machine

// CPUFrequency returns the current CPU frequency for ESP32-S3-BOX-3.
// Returns 240MHz (max speed for ESP32-S3 with PSRAM)
func CPUFrequency() uint32 {
	return 240_000_000 // 240MHz
}

const (
	// GPIO pins available on ESP32-S3-BOX-3
	// Based on ESP-BOX-3 BSP pin definitions
	IO0  = GPIO0 // Button Config
	IO1  = GPIO1 // Button Mute / Mute Status
	IO2  = GPIO2 // I2S MCLK
	IO3  = GPIO3 // LCD Touch Interrupt
	IO4  = GPIO4 // LCD DC
	IO5  = GPIO5 // LCD CS
	IO6  = GPIO6 // LCD DATA0 (MOSI)
	IO7  = GPIO7 // LCD PCLK (SCK)
	IO8  = GPIO8 // I2C SDA
	IO9  = GPIO9 // SD Card D0
	IO10 = GPIO10
	IO11 = GPIO11 // SD Card CLK
	IO12 = GPIO12 // SD Card D3
	IO13 = GPIO13 // SD Card D1
	IO14 = GPIO14 // SD Card CMD
	IO15 = GPIO15 // I2S DOUT (Speaker)
	IO16 = GPIO16 // I2S DSIN (Microphone)
	IO17 = GPIO17 // I2S SCLK
	IO18 = GPIO18 // I2C SCL
	IO19 = GPIO19 // USB_NEG
	IO20 = GPIO20 // USB_POS
	IO21 = GPIO21
	IO38 = GPIO38
	IO39 = GPIO39
	IO40 = GPIO40 // I2C Dock SCL
	IO41 = GPIO41 // I2C Dock SDA
	IO42 = GPIO42 // SD Card D2
	IO43 = GPIO43 // SD Card Power
	IO44 = GPIO44 // (Unused)
	IO45 = GPIO45 // I2S LCLK (LRCK)
	IO46 = GPIO46 // Power Amp Control
	IO47 = GPIO47 // LCD Backlight
	IO48 = GPIO48 // LCD Reset
)

// SPI pins
const (
	// SPI1 - used for LCD
	SPI1_SCK_PIN  = IO7   // LCD_PCLK
	SPI1_MOSI_PIN = IO6   // LCD_DATA0 (MOSI)
	SPI1_MISO_PIN = NoPin // Not used for LCD
	SPI1_CS_PIN   = IO5   // LCD_CS

	// SPI2 (not used on this board)
	SPI2_SCK_PIN  = NoPin
	SPI2_MOSI_PIN = NoPin
	SPI2_MISO_PIN = NoPin
	SPI2_CS_PIN   = NoPin
)

// LCD pins (ST7789/ILI9341)
const (
	LCD_SCK_PIN = SPI1_SCK_PIN
	LCD_SDO_PIN = SPI1_MOSI_PIN
	LCD_SDI_PIN = NoPin
	LCD_SS_PIN  = SPI1_CS_PIN
	LCD_DC_PIN  = IO4
	LCD_RST_PIN = IO48
	LCD_BL_PIN  = IO47
)

// I2C pins (Internal I2C for sensors)
const (
	SDA0_PIN = IO8
	SCL0_PIN = IO18

	SDA_PIN = SDA0_PIN
	SCL_PIN = SCL0_PIN
)

// Dock I2C pins (for expansion dock)
const (
	SDA1_PIN = IO41
	SCL1_PIN = IO40
)

// UART pins (USB CDC - native USB on ESP32-S3)
// Note: Native USB doesn't require GPIO pins, these are kept for compatibility
const (
	UART_TX_PIN = NoPin
	UART_RX_PIN = NoPin
)

// Built-in LED
// Using LCD backlight as LED indicator
const LED = GPIO47

// Buttons
const (
	BUTTON_CONFIG = GPIO0
	BUTTON_MUTE   = GPIO1
)

// Audio pins (I2S)
const (
	I2S_SCLK_PIN = IO17 // Bit Clock
	I2S_MCLK_PIN = IO2  // Master Clock
	I2S_LRCK_PIN = IO45 // Left/Right Clock (Frame Sync)
	I2S_DOUT_PIN = IO15 // Data Out to Speaker (ES8311)
	I2S_DSIN_PIN = IO16 // Data In from Microphone (ES7210)
)

// SD Card pins (SD/MMC mode)
const (
	SDCARD_CMD_PIN = IO14
	SDCARD_CLK_PIN = IO11
	SDCARD_D0_PIN  = IO9
	SDCARD_D1_PIN  = IO13
	SDCARD_D2_PIN  = IO42
	SDCARD_D3_PIN  = IO12
	SDCARD_PWR_PIN = IO43 // SD Card Power
)

// USB pins (native USB)
const (
	USB_DPPIN = IO20 // USB D+
	USB_DMPIN = IO19 // USB D-
)

// Touch interrupt
const TOUCH_INT_PIN = IO3

// Power amplifier control
const POWER_AMP_PIN = IO46
