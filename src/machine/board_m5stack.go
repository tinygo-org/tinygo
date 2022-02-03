//go:build m5stack
// +build m5stack

package machine

const (
	//     GND     | ADC     G35
	//     GND     | ADC     G36
	//     GND     | RST     EN
	// G23 MOSI    | DAC/SPK G25
	// G19 MISO    | DAC     G26
	// G18 SCK     | 3.3V
	// G3  RXD1    | TXD1    G1
	// G16 RXD2    | TXD2    G17
	// G21 SDA     | DCL     G22
	// G2  GPIO    | GPIO    G5
	// G12 IIS_SK  | IIS_WS  G13
	// G15 IIS_OUT | IIS_MK  G0
	//     HPWR    | IIS_IN  G34
	//     HPWR    | 5V
	//     HPWR    | BATTERY

	IO0  Pin = 0
	IO1  Pin = 1
	IO2  Pin = 2
	IO3  Pin = 3
	IO4  Pin = 4
	IO5  Pin = 5
	IO6  Pin = 6
	IO7  Pin = 7
	IO8  Pin = 8
	IO9  Pin = 9
	IO10 Pin = 10
	IO11 Pin = 11
	IO12 Pin = 12
	IO13 Pin = 13
	IO14 Pin = 14
	IO15 Pin = 15
	IO16 Pin = 16
	IO17 Pin = 17
	IO18 Pin = 18
	IO19 Pin = 19
	IO20 Pin = 20
	IO21 Pin = 21
	IO22 Pin = 22
	IO23 Pin = 23
	IO24 Pin = 24
	IO25 Pin = 25
	IO26 Pin = 26
	IO27 Pin = 27
	IO28 Pin = 28
	IO29 Pin = 29
	IO30 Pin = 30
	IO31 Pin = 31
	IO32 Pin = 32
	IO33 Pin = 33
	IO34 Pin = 34
	IO35 Pin = 35
	IO36 Pin = 36
	IO37 Pin = 37
	IO38 Pin = 38
	IO39 Pin = 39
)

const (
	// Buttons
	BUTTON_A = IO39
	BUTTON_B = IO38
	BUTTON_C = IO37
	BUTTON   = BUTTON_A

	// Speaker
	SPEAKER_PIN = IO25
)

// SPI pins
const (
	SPI0_SCK_PIN = IO18
	SPI0_SDO_PIN = IO23
	SPI0_SDI_PIN = IO19
	SPI0_CS0_PIN = IO14

	// LCD (ILI9342C)
	LCD_SCK_PIN = SPI0_SCK_PIN
	LCD_SDO_PIN = SPI0_SDO_PIN
	LCD_SDI_PIN = SPI0_SDI_PIN // NoPin ?
	LCD_SS_PIN  = SPI0_CS0_PIN
	LCD_DC_PIN  = IO27
	LCD_RST_PIN = IO33
	LCD_BL_PIN  = IO32

	// SD CARD
	SDCARD_SCK_PIN = SPI0_SCK_PIN
	SDCARD_SDO_PIN = SPI0_SDO_PIN
	SDCARD_SDI_PIN = SPI0_SDI_PIN
	SDCARD_SS_PIN  = IO4
)

// I2C pins
const (
	SDA0_PIN = IO21
	SCL0_PIN = IO22

	SDA_PIN = SDA0_PIN
	SCL_PIN = SCL0_PIN
)

// ADC pins
const (
	ADC1 Pin = IO35
	ADC2 Pin = IO36
)

// DAC pins
const (
	DAC1 Pin = IO25
	DAC2 Pin = IO26
)

// UART pins
const (
	// UART0 (CP2104)
	UART0_TX_PIN = IO1
	UART0_RX_PIN = IO3

	UART1_TX_PIN = IO17
	UART1_RX_PIN = IO16

	UART_TX_PIN = UART0_TX_PIN
	UART_RX_PIN = UART0_RX_PIN
)
