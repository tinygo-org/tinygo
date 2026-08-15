//go:build stm32f401

package machine

// STM32F401 peripheral implementations. The F401 has a reduced set of
// peripherals compared to F405/F407, so this file provides F401-specific
// versions of the functions in machine_stm32f4_extended.go.

import (
	"device/stm32"
	"runtime/interrupt"
	"unsafe"
)

func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		return stm32.GPIOC
	case 3:
		return stm32.GPIOD
	case 4:
		return stm32.GPIOE
	case 7:
		return stm32.GPIOH
	default:
		panic("machine: unknown port")
	}
}

// Enable peripheral clock for STM32F401 peripherals.
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.PWR):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	case unsafe.Pointer(stm32.I2C3):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C3EN)
	case unsafe.Pointer(stm32.I2C2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C2EN)
	case unsafe.Pointer(stm32.I2C1):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)
	case unsafe.Pointer(stm32.USART2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)
	case unsafe.Pointer(stm32.SPI3):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_SPI3EN)
	case unsafe.Pointer(stm32.SPI2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_SPI2EN)
	case unsafe.Pointer(stm32.WWDG):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_WWDGEN)
	case unsafe.Pointer(stm32.TIM5):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM5EN)
	case unsafe.Pointer(stm32.TIM4):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM4EN)
	case unsafe.Pointer(stm32.TIM3):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM3EN)
	case unsafe.Pointer(stm32.TIM2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM2EN)
	case unsafe.Pointer(stm32.TIM11):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM11EN)
	case unsafe.Pointer(stm32.TIM10):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM10EN)
	case unsafe.Pointer(stm32.TIM9):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM9EN)
	case unsafe.Pointer(stm32.SYSCFG):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SYSCFGEN)
	case unsafe.Pointer(stm32.SPI4):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI4EN)
	case unsafe.Pointer(stm32.SPI1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	case unsafe.Pointer(stm32.SDIO):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SDIOEN)
	case unsafe.Pointer(stm32.ADC1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC1EN)
	case unsafe.Pointer(stm32.USART6):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART6EN)
	case unsafe.Pointer(stm32.USART1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	case unsafe.Pointer(stm32.TIM1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM1EN)
	}
}

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_UP_TIM10, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleUPInterrupt)
	case &TIM4:
		return interrupt.New(stm32.IRQ_TIM4, TIM4.handleUPInterrupt)
	case &TIM5:
		return interrupt.New(stm32.IRQ_TIM5, TIM5.handleUPInterrupt)
	case &TIM9:
		return interrupt.New(stm32.IRQ_TIM1_BRK_TIM9, TIM9.handleUPInterrupt)
	case &TIM10:
		return interrupt.New(stm32.IRQ_TIM1_UP_TIM10, TIM10.handleUPInterrupt)
	case &TIM11:
		return interrupt.New(stm32.IRQ_TIM1_TRG_COM_TIM11, TIM11.handleUPInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) registerOCInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_CC, TIM1.handleOCInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleOCInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleOCInterrupt)
	case &TIM4:
		return interrupt.New(stm32.IRQ_TIM4, TIM4.handleOCInterrupt)
	case &TIM5:
		return interrupt.New(stm32.IRQ_TIM5, TIM5.handleOCInterrupt)
	case &TIM9:
		return interrupt.New(stm32.IRQ_TIM1_BRK_TIM9, TIM9.handleOCInterrupt)
	case &TIM10:
		return interrupt.New(stm32.IRQ_TIM1_UP_TIM10, TIM10.handleOCInterrupt)
	case &TIM11:
		return interrupt.New(stm32.IRQ_TIM1_TRG_COM_TIM11, TIM11.handleOCInterrupt)
	}

	return interrupt.Interrupt{}
}

// initRNG is a no-op on STM32F401 which has no hardware RNG peripheral.
func initRNG() {}

func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var clock uint32
	switch uart.Bus {
	case stm32.USART1, stm32.USART6:
		clock = CPUFrequency() // APB2 = HCLK (no prescaler)
	case stm32.USART2:
		clock = CPUFrequency() / 2 // APB1 = HCLK/2
	}
	return clock / baudRate
}
