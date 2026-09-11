//go:build nucleoh723zg

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const xtalHz = 8_000_000
const hseBypass = true

const (
	// Arduino Pins
	A0 = PA3  // ADC12_INP15
	A1 = PC0  // ADC123_INP10
	A2 = PC3  // ADC12_INP13
	A3 = PB1  // ADC12_INP5
	A4 = PC2  // ADC123_INP12 || I2C1_SDA
	A5 = PF10 // ADC3_INP6 || I2C1_SCL
	A6 = PF4  // ADC3_INP9
	A7 = PF5  // ADC3_INP4
	A8 = PF6  // ADC3_INP8

	D0  = PB7 // LPUART1
	D1  = PB6 // LPUART1
	D2  = PG14
	D3  = PE13 // TIM3_CH3
	D4  = PE14
	D5  = PE11 // TIM1_CH2 / I2C1_SCL
	D6  = PE9  // TIM1_CH2
	D7  = PG12
	D8  = PF3
	D9  = PD15 // TIM4_CH4
	D10 = PD14 // SPI1_CS || TIM4_CH3
	D11 = PB5  // SPI1_MOSI || TIM3_CH2
	D12 = PA6  // SPI1_MISO
	D13 = PA5  // SPI1_SCK
	D14 = PB9
	D15 = PB8
)

const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PB0
	LED_YELLOW  = PE1
	LED_RED     = PB14
)

const (
	BUTTON = PC13
)

const (
	// UART pins
	// PA2 and PA3 are connected to the ST-Link Virtual Com Port (VCP)
	UART_TX_PIN = PD8
	UART_RX_PIN = PD9

	// SPI
	SPI1_SCK_PIN = PA5
	SPI1_SDI_PIN = PB5
	SPI1_SDO_PIN = PA6
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN

	// I2C pins
	I2C0_SCL_PIN = PF1 // I2C2
	I2C0_SDA_PIN = PF0 // I2C2
)

var (
	// USART3 is the hardware serial port connected to the
	// onboard ST-LINK debugger to be exposed as virtual COM
	// port over USB on Nucleo boards.
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART3,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 7,
	}
	DefaultUART = UART1

	// I2C2 is documented, alias to I2C0 as well
	I2C2 = &I2C{
		Bus:             stm32.I2C2,
		AltFuncSelector: 4,
	}
	I2C0 = I2C2
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART3, _UART1.handleInterrupt)
}
