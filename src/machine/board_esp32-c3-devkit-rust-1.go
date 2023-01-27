//go:build esp32_c3_devkit_rust_1

// This file contains the pin mappings for the Espressif ESP32-C3 Development Board for Rust.
//
// The Espressif ESP32-C3-DevKit-RUST-1 development board is powered
// by the Espressif ESP32-C3 SoC featuring an open-source RISC-V architecture.
//
// Specifications:
//   SoC: ESP32-C3-MINI-1, 4MB Flash, RISCV-32bit, 160MHz, 400KB SRAM
//   Wireless: WiFi & Bluetooth 5.0 (BLE)
//   ICM-42670-P 6-Axis IMU  (I2C Addr 0x68)
//   SHTC3 Humidity and Temperature Sensor (I2C Addr 0x70)
//   WS2812B LED

// GitHub:    https://github.com/esp-rs/esp-rust-board
// Schematic: https://github.com/esp-rs/esp-rust-board/blob/master/hardware/esp-rust-board/schematic/esp-rust-board.pdf
// Datasheet: https://www.espressif.com/sites/default/files/documentation/esp32-c3_datasheet_en.pdf

package machine

// Digital pins
const (
	//    Pin    // Function
	//    -----  // ---------------
	D0  = GPIO0  //
	D1  = GPIO1  //
	D2  = GPIO2  // WS2812
	D3  = GPIO3  //
	D4  = GPIO4  // MTMS
	D5  = GPIO5  // MTDI
	D6  = GPIO6  // MTCK
	D7  = GPIO7  // Red LED / MTDO
	D8  = GPIO8  // I2C SCL
	D9  = GPIO9  // Boot Button
	D10 = GPIO10 // I2C SDA
	D18 = GPIO18 // USB DM
	D19 = GPIO19 // USB DP
	D20 = GPIO20 // UART RX
	D21 = GPIO21 // UART TX
)

// Analog pins
const (
	A0 = GPIO0
	A1 = GPIO1
	A2 = GPIO2
	A3 = GPIO3
	A4 = GPIO4
	A5 = GPIO5
)

// Button pin
const (
	BUTTON      = BUTTON_BOOT
	BUTTON_BOOT = D9
)

// LED pins
const (
	LED         = LED_BUILTIN
	WS2812      = D2
	LED_BUILTIN = D7
)

// I2C pins
const (
	I2C_SCL_PIN = D8
	I2C_SDA_PIN = D10
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
