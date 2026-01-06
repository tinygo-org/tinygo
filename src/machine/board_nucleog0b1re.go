//go:build nucleog0b1re

// Schematic: https://www.st.com/resource/en/user_manual/um2324-stm32-nucleo64-boards-mb1360-stmicroelectronics.pdf
// Datasheet: https://www.st.com/resource/en/datasheet/stm32g0b1re.pdf

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	// Arduino Pins
	A0 = PA0
	A1 = PA1
	A2 = PA4
	A3 = PB1
	A4 = PA11
	A5 = PA12

	D0  = PB7
	D1  = PB6
	D2  = PA10
	D3  = PB3
	D4  = PB5
	D5  = PB4
	D6  = PB10
	D7  = PA8
	D8  = PA9
	D9  = PC7
	D10 = PB0
	D11 = PA7
	D12 = PA6
	D13 = PA5
	D14 = PB9
	D15 = PB8
)

// User LD4: the green LED is a user LED connected to ARDUINO signal D13 corresponding
// to STM32 I/O PA5.
const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PA5
)

// User B1: the user button is connected to PC13.
const (
	BUTTON = PC13
)

const (
	// UART pins
	// PA2 and PA3 are connected to the ST-Link Virtual Com Port (VCP)
	UART_TX_PIN = PA2
	UART_RX_PIN = PA3

	// I2C pins
	// PB8 is SCL (connected to Arduino connector D15)
	// PB9 is SDA (connected to Arduino connector D14)
	I2C0_SCL_PIN = PB8
	I2C0_SDA_PIN = PB9

	// SPI pins
	SPI1_SCK_PIN = PA5
	SPI1_SDI_PIN = PA6
	SPI1_SDO_PIN = PA7
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN

	// CAN pins (directly accessible on Nucleo-G0B1RE board)
	// FDCAN1: PA11 (TX) / PA12 (RX) using AF9
	// FDCAN2: PD12 (TX) / PD13 (RX) using AF3
	CAN1_TX_PIN = PA11
	CAN1_RX_PIN = PA12
	CAN2_TX_PIN = PD12
	CAN2_RX_PIN = PD13
)

var (
	// USART2 is the hardware serial port connected to the onboard ST-LINK
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: AF1_TIM1_TIM2_TIM3_LPTIM1,
		RxAltFuncSelector: AF1_TIM1_TIM2_TIM3_LPTIM1,
	}
	DefaultUART = UART1

	// I2C1 is documented, alias to I2C0 as well
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: AF6_SPI2_USART3_USART4_I2C1,
	}
	I2C0 = I2C1

	// SPI1 is documented, alias to SPI0 as well
	SPI1 = &SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF0_SYSTEM,
	}
	SPI0 = SPI1

	// FDCAN1 on PA11 (TX) / PA12 (RX)
	CAN1  = &_CAN1
	_CAN1 = FDCAN{
		Bus:             stm32.FDCAN1,
		TxAltFuncSelect: AF9_FDCAN1_FDCAN2,
		RxAltFuncSelect: AF9_FDCAN1_FDCAN2,
		instance:        0,
	}

	// FDCAN2 on PD12 (TX) / PD13 (RX)
	CAN2  = &_CAN2
	_CAN2 = FDCAN{
		Bus:             stm32.FDCAN2,
		TxAltFuncSelect: AF3_FDCAN1_FDCAN2,
		RxAltFuncSelect: AF3_FDCAN1_FDCAN2,
		instance:        1,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART2_LPUART2, _UART1.handleInterrupt)
	// Note: FDCAN interrupts share with USB (IRQ_UCPD1_UCPD2_USB = 8)
	// User should configure interrupts via SetInterrupt method if needed
}
