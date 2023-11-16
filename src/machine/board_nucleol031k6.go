//go:build nucleol031k6

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	// Arduino Pins
	A0 = PA0 // ADC_IN0
	A1 = PA1 // ADC_IN1
	A2 = PA3 // ADC_IN3
	A3 = PA4 // ADC_IN4
	A4 = PA5 // ADC_IN5 || I2C1_SDA
	A5 = PA6 // ADC_IN6 || I2C1_SCL
	A6 = PA7 // ADC_IN7
	A7 = PA2 // ADC_IN2

	D0  = PA10 // USART1_TX
	D1  = PA9  // USART1_RX
	D2  = PA12
	D3  = PB0 // TIM2_CH3
	D4  = PB7
	D5  = PB6  // TIM16_CH1N
	D6  = PB1  // TIM14_CH1
	D9  = PA8  // TIM1_CH1
	D10 = PA11 // SPI_CS || TIM1_CH4
	D11 = PB5  // SPI1_MOSI || TIM3_CH2
	D12 = PB4  // SPI1_MISO
	D13 = PB3  // SPI1_SCK
)

const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PB3
)

const (
	// This board does not have a user button, so
	// use first GPIO pin by default
	BUTTON = PA0
)

const (
	// UART pins
	// PA2 and PA15 are connected to the ST-Link Virtual Com Port (VCP)
	UART_TX_PIN = PA2
	UART_RX_PIN = PA15

	// SPI
	SPI1_SCK_PIN = PB3
	SPI1_SDI_PIN = PB5
	SPI1_SDO_PIN = PB4
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN

	// I2C pins
	// PB6 and PB7 are mapped to CN4 pin 7 and CN4 pin 8 respectively with the
	// default solder bridge settings
	I2C0_SCL_PIN  = PB7
	I2C0_SDA_PIN  = PB6
	I2C0_ALT_FUNC = 1
)

var (
	// USART2 is the hardware serial port connected to the onboard ST-LINK
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	UART1  = &_UART1
	_UART1 = UART{
		UARTCommon:        NewUARTCommon(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: 4,
		RxAltFuncSelector: 4,
	}
	DefaultUART = UART1

	// I2C1 is documented, alias to I2C0 as well
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: 1,
	}
	I2C0 = I2C1

	// SPI
	SPI0 = SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: 0,
	}
	SPI1 = &SPI0
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART1.handleInterrupt)
}
