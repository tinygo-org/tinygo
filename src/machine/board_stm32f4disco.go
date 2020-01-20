// +build stm32f4disco

package machine

const (
	PA0  = portA + 0
	PA1  = portA + 1
	PA2  = portA + 2
	PA3  = portA + 3
	PA4  = portA + 4
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA8  = portA + 8
	PA9  = portA + 9
	PA10 = portA + 10
	PA11 = portA + 11
	PA12 = portA + 12
	PA13 = portA + 13
	PA14 = portA + 14
	PA15 = portA + 15

	PB0  = portB + 0
	PB1  = portB + 1
	PB2  = portB + 2
	PB3  = portB + 3
	PB4  = portB + 4
	PB5  = portB + 5
	PB6  = portB + 6
	PB7  = portB + 7
	PB8  = portB + 8
	PB9  = portB + 9
	PB10 = portB + 10
	PB11 = portB + 11
	PB12 = portB + 12
	PB13 = portB + 13
	PB14 = portB + 14
	PB15 = portB + 15

	PC0  = portC + 0
	PC1  = portC + 1
	PC2  = portC + 2
	PC3  = portC + 3
	PC4  = portC + 4
	PC5  = portC + 5
	PC6  = portC + 6
	PC7  = portC + 7
	PC8  = portC + 8
	PC9  = portC + 9
	PC10 = portC + 10
	PC11 = portC + 11
	PC12 = portC + 12
	PC13 = portC + 13
	PC14 = portC + 14
	PC15 = portC + 15

	PD0  = portD + 0
	PD1  = portD + 1
	PD2  = portD + 2
	PD3  = portD + 3
	PD4  = portD + 4
	PD5  = portD + 5
	PD6  = portD + 6
	PD7  = portD + 7
	PD8  = portD + 8
	PD9  = portD + 9
	PD10 = portD + 10
	PD11 = portD + 11
	PD12 = portD + 12
	PD13 = portD + 13
	PD14 = portD + 14
	PD15 = portD + 15

	PE0  = portE + 0
	PE1  = portE + 1
	PE2  = portE + 2
	PE3  = portE + 3
	PE4  = portE + 4
	PE5  = portE + 5
	PE6  = portE + 6
	PE7  = portE + 7
	PE8  = portE + 8
	PE9  = portE + 9
	PE10 = portE + 10
	PE11 = portE + 11
	PE12 = portE + 12
	PE13 = portE + 13
	PE14 = portE + 14
	PE15 = portE + 15

	PH0 = portH + 0
	PH1 = portH + 1
)

const (
	LED         = LED_BUILTIN
	LED1        = LED_GREEN
	LED2        = LED_ORANGE
	LED3        = LED_RED
	LED4        = LED_BLUE
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PD12
	LED_ORANGE  = PD13
	LED_RED     = PD14
	LED_BLUE    = PD15
)

// UART pins
const (
	UART_TX_PIN = PA2
	UART_RX_PIN = PA3
)

// I2C pins
const (
	I2C1_SCL = PB6
	I2C1_SDA = PB9
	I2C0_SCL = I2C1_SCL
	I2C0_SDA = I2C1_SDA
)

// SPI pins
const (
	SPI1_SCK_PIN  = PA5
	SPI1_MISO_PIN = PA6
	SPI1_MOSI_PIN = PA7
	SPI0_SCK_PIN  = SPI1_SCK_PIN
	SPI0_MISO_PIN = SPI1_MISO_PIN
	SPI0_MOSI_PIN = SPI1_MOSI_PIN
)

// MEMs accelerometer
const (
	MEMS_ACCEL_CS   = PE3
	MEMS_ACCEL_INT1 = PE0
	MEMS_ACCEL_INT2 = PE1
)

// MEMs microphone
const (
	MEMS_MICROPHONE_CLK  = PB10
	MEMS_MICROPHONE_DATA = PC3
)

// I2S audio I/O driver CS43L22
// I2C address 0x94
const (
	AUDIO_DRIVER_SCL = I2C1_SCL
	AUDIO_DRIVER_SDA = I2C1_SDA
)
