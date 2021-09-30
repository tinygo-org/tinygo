// +build m5stack_core2

package machine

const (
	//     GND      | ADC       G35
	//     GND      | ADC       G36
	//     GND      | RST       EN
	// G23 MOSI     | DAC       G25
	// G38 MISO     | DAC       G26
	// G18 SCK      | 3.3V
	// G3  RXD0     | TXD0      G1
	// G13 RXD2     | TXD2      G14
	// G21 intSDA   | intSC     G22
	// G32 PA_SDA   | PA_SCL     G33
	// G27 GPIO     | GPIO      G19
	// G2  I2S_DOUT | I2S_LRCKC G0
	//     N/C      | PDM_DAT   G34
	//     N/C      | 5V
	//     N/C      | BAT

	IO0  Pin = 0
	IO1  Pin = 1 // U0TXD
	IO2  Pin = 2
	IO3  Pin = 3 // U0RXD
	IO4  Pin = 4
	IO5  Pin = 5
	IO6  Pin = 6  // SD_CLK
	IO7  Pin = 7  // SD_DATA0
	IO8  Pin = 8  // SD_DATA1
	IO9  Pin = 9  // SD_DATA2
	IO10 Pin = 10 // SD_DATA3
	IO11 Pin = 11 // SD_CMD
	IO12 Pin = 12
	IO13 Pin = 13 // U0RXD
	IO14 Pin = 14 // U1TXD
	IO15 Pin = 15
	IO16 Pin = 16
	IO17 Pin = 17
	IO18 Pin = 18 // SPI0_SCK
	IO19 Pin = 19
	IO21 Pin = 21 // SDA0
	IO22 Pin = 22 // SCL0
	IO23 Pin = 23 // SPI0_SDO
	IO25 Pin = 25
	IO26 Pin = 26
	IO27 Pin = 27
	IO32 Pin = 32 // SDA1
	IO33 Pin = 33 // SCL1
	IO34 Pin = 34
	IO35 Pin = 35 // ADC1
	IO36 Pin = 36 // ADC2
	IO38 Pin = 38 // SPI0_SDI
	IO39 Pin = 39
)

// SPI pins
const (
	SPI0_SCK_PIN = IO18
	SPI0_SDO_PIN = IO23
	SPI0_SDI_PIN = IO38
	SPI0_CS0_PIN = IO5

	// LCD (ILI9342C)
	LCD_SCK_PIN = SPI0_SCK_PIN
	LCD_SDO_PIN = SPI0_SDO_PIN
	LCD_SDI_PIN = SPI0_SDI_PIN
	LCD_SS_PIN  = SPI0_CS0_PIN
	LCD_DC_PIN  = IO15

	// SD CARD
	SDCARD_SCK_PIN = SPI0_SCK_PIN
	SDCARD_SDO_PIN = SPI0_SDO_PIN
	SDCARD_SDI_PIN = SPI0_SDI_PIN
	SDCARD_SS_PIN  = IO4
)

// I2C pins
const (
	// Internal I2C (AXP192 / FT6336U / BM8563 / MPU6886)
	SDA0_PIN = IO21
	SCL0_PIN = IO22

	// External I2C (PORT A)
	SDA1_PIN = IO32
	SCL1_PIN = IO33

	SDA_PIN = SDA1_PIN
	SCL_PIN = SCL1_PIN
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

	UART1_TX_PIN = IO14
	UART1_RX_PIN = IO13

	UART_TX_PIN = UART0_TX_PIN
	UART_RX_PIN = UART0_RX_PIN
)
