//go:build arduino_uno_q

// Arduino UNO Q board with STM32U585 processor.

package machine

import (
	"device/stm32"
	"runtime/interrupt"
	"unsafe"
)

const (
	// Arduino Pins
	A0 = PA4
	A1 = PA5
	A2 = PA6
	A3 = PA7
	A4 = PC1
	A5 = PC0

	// ADC pin aliases
	ADC0 = A0
	ADC1 = A1
	ADC2 = A2
	ADC3 = A3
	ADC4 = A4
	ADC5 = A5

	D0  = PB7 // USART1 RX, PWM TIM4_CH2
	D1  = PB6 // USART1 TX, PWM TIM4_CH1
	D2  = PB3 // PWM TIM2_CH2
	D3  = PB0 // PWM TIM3_CH3
	D4  = PA12
	D5  = PA11 // PWM TIM1_CH4
	D6  = PB1  // PWM TIM3_CH4
	D7  = PB2
	D8  = PB4  // PWM TIM3_CH1
	D9  = PB8  // PWM TIM4_CH3 / TIM16_CH1
	D10 = PB9  // PWM TIM4_CH4 / TIM17_CH1
	D11 = PB15 // PWM TIM15_CH2
	D12 = PB14 // PWM TIM15_CH1
	D13 = PB13
	D18 = PC1
	D19 = PC0
	D20 = PB10 // I2C2 SCL, PWM TIM2_CH3
	D21 = PB11 // I2C2 SDA, PWM TIM2_CH4
)

const (
	LED    = LED3_R
	LED3_R = PH10
	LED3_G = PH11
	LED3_B = PH12
	LED4_R = PH13
	LED4_G = PH14
	LED4_B = PH15
)

const (
	// Default UART pins (LPUART1 active via ST-LINK virtual COM port)
	UART_TX_PIN = PG7
	UART_RX_PIN = PG8

	// USART1 pins (Arduino header D1/D0)
	UART1_TX_PIN = D1
	UART1_RX_PIN = D0

	// LPUART1 pins (active via ST-LINK virtual COM port)
	UART2_TX_PIN = PG7
	UART2_RX_PIN = PG8

	// I2C pins, also connected to I2C2
	I2C0_SCL_PIN = D20
	I2C0_SDA_PIN = D21
	// QWIIC connector pins, also connected to I2C4
	I2C1_SCL_PIN = PD12
	I2C1_SDA_PIN = PD13

	// SPI pins
	SPI1_SCK_PIN = PA5
	SPI1_SDI_PIN = PA6
	SPI1_SDO_PIN = PA7
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN
)

var (
	// USART1 on PB6/PB7 (Arduino header D1/D0).
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: AF7_USART1_2_3,
		RxAltFuncSelector: AF7_USART1_2_3,
	}

	// LPUART1 on PG7/PG8 (active via ST-LINK virtual COM port).
	UART2  = &_UART2
	_UART2 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               (*stm32.USART_Type)(unsafe.Pointer(stm32.LPUART1)),
		TxAltFuncSelector: AF8_UART4_5_LPUART1_SDMMC1,
		RxAltFuncSelector: AF8_UART4_5_LPUART1_SDMMC1,
	}
	DefaultUART = UART2

	// I2C2 is documented, alias to I2C0 as well
	I2C2 = &I2C{
		Bus:             stm32.I2C2,
		AltFuncSelector: AF4_I2C1_2_3_4,
	}
	I2C0 = I2C2

	// I2C4 is is connected to the QWIIC connector, alias to I2C1 as well
	I2C4 = &I2C{
		Bus:             stm32.I2C4,
		AltFuncSelector: AF4_I2C1_2_3_4,
	}
	I2C1 = I2C4

	// SPI1 is documented, alias to SPI0 as well
	SPI1 = &SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF5_SPI1_2_3_OCTOSPI1_OCTOSPI2,
	}
	SPI0 = SPI1
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(stm32.IRQ_LPUART1, _UART2.handleInterrupt)
}
