//go:build stm32f4disco
// +build stm32f4disco

package machine

import (
	"runtime/interrupt"
	"tinygo.org/x/device/stm32"
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

const (
	BUTTON = PA0
)

// Analog Pins
const (
	ADC0  = PA0
	ADC1  = PA1
	ADC2  = PA2
	ADC3  = PA3
	ADC4  = PA4
	ADC5  = PA5
	ADC6  = PA6
	ADC7  = PA7
	ADC8  = PB0
	ADC9  = PB1
	ADC10 = PC0
	ADC11 = PC1
	ADC12 = PC2
	ADC13 = PC3
	ADC14 = PC4
	ADC15 = PC5
)

// UART pins
const (
	UART_TX_PIN = PA2
	UART_RX_PIN = PA3
)

var (
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: AF7_USART1_2_3,
		RxAltFuncSelector: AF7_USART1_2_3,
	}
	DefaultUART = UART1
)

// set up RX IRQ handler. Follow similar pattern for other UARTx instances
func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART1.handleInterrupt)
}

// SPI pins
const (
	SPI1_SCK_PIN = PA5
	SPI1_SDI_PIN = PA6
	SPI1_SDO_PIN = PA7
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN
)

// MEMs accelerometer
const (
	MEMS_ACCEL_CS   = PE3
	MEMS_ACCEL_INT1 = PE0
	MEMS_ACCEL_INT2 = PE1
)

// Since the first interface is named SPI1, both SPI0 and SPI1 refer to SPI1.
// TODO: implement SPI2 and SPI3.
var (
	SPI0 = SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF5_SPI1_SPI2,
	}
	SPI1 = &SPI0
)

const (
	I2C0_SCL_PIN = PB6
	I2C0_SDA_PIN = PB9
)

var (
	I2C0 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: AF4_I2C1_2_3,
	}
)
