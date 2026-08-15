//go:build stm32f4 && !stm32f401

package machine

// This file contains stm32f4 peripheral implementations for chips that have the
// full set of F4 peripherals (F405, F407, F469, etc.). Chips with fewer
// peripherals (e.g. F401) use machine_stm32f401_periph.go instead.

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
	case 5:
		return stm32.GPIOF
	case 6:
		return stm32.GPIOG
	case 7:
		return stm32.GPIOH
	case 8:
		return stm32.GPIOI
	case 9:
		return stm32.GPIOJ
	case 10:
		return stm32.GPIOK
	default:
		panic("machine: unknown port")
	}
}

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.DAC):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_DACEN)
	case unsafe.Pointer(stm32.PWR):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	case unsafe.Pointer(stm32.CAN2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_CAN2EN)
	case unsafe.Pointer(stm32.CAN1):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_CAN1EN)
	case unsafe.Pointer(stm32.I2C3):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C3EN)
	case unsafe.Pointer(stm32.I2C2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C2EN)
	case unsafe.Pointer(stm32.I2C1):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)
	case unsafe.Pointer(stm32.UART5):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_UART5EN)
	case unsafe.Pointer(stm32.UART4):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_UART4EN)
	case unsafe.Pointer(stm32.USART3):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART3EN)
	case unsafe.Pointer(stm32.USART2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)
	case unsafe.Pointer(stm32.SPI3):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_SPI3EN)
	case unsafe.Pointer(stm32.SPI2):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_SPI2EN)
	case unsafe.Pointer(stm32.WWDG):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_WWDGEN)
	case unsafe.Pointer(stm32.TIM14):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM14EN)
	case unsafe.Pointer(stm32.TIM13):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM13EN)
	case unsafe.Pointer(stm32.TIM12):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM12EN)
	case unsafe.Pointer(stm32.TIM7):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM7EN)
	case unsafe.Pointer(stm32.TIM6):
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM6EN)
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
	case unsafe.Pointer(stm32.SPI1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	case unsafe.Pointer(stm32.SDIO):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SDIOEN)
	case unsafe.Pointer(stm32.ADC3):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC3EN)
	case unsafe.Pointer(stm32.ADC2):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC2EN)
	case unsafe.Pointer(stm32.ADC1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC1EN)
	case unsafe.Pointer(stm32.USART6):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART6EN)
	case unsafe.Pointer(stm32.USART1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	case unsafe.Pointer(stm32.TIM8):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM8EN)
	case unsafe.Pointer(stm32.TIM1):
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM1EN)
	}
}

var (
	TIM6 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM6EN,
		Device:         stm32.TIM6,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM7 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM7EN,
		Device:         stm32.TIM7,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM8 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM8EN,
		Device:         stm32.TIM8,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PC6, AF3_TIM8_9_10_11}, {PI5, AF3_TIM8_9_10_11}}},
			TimerChannel{Pins: []PinFunction{{PC7, AF3_TIM8_9_10_11}, {PI6, AF3_TIM8_9_10_11}}},
			TimerChannel{Pins: []PinFunction{{PC8, AF3_TIM8_9_10_11}, {PI7, AF3_TIM8_9_10_11}}},
			TimerChannel{Pins: []PinFunction{{PC9, AF3_TIM8_9_10_11}, {PI2, AF3_TIM8_9_10_11}}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM12 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM12EN,
		Device:         stm32.TIM12,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PB14, AF9_CAN1_CAN2_TIM12_13_14}, {PH6, AF9_CAN1_CAN2_TIM12_13_14}}},
			TimerChannel{Pins: []PinFunction{{PB15, AF9_CAN1_CAN2_TIM12_13_14}, {PH9, AF9_CAN1_CAN2_TIM12_13_14}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM13 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM13EN,
		Device:         stm32.TIM13,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA6, AF9_CAN1_CAN2_TIM12_13_14}, {PF8, AF9_CAN1_CAN2_TIM12_13_14}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM14 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM14EN,
		Device:         stm32.TIM14,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA7, AF9_CAN1_CAN2_TIM12_13_14}, {PF9, AF9_CAN1_CAN2_TIM12_13_14}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}
)

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
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6_DAC, TIM6.handleUPInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleUPInterrupt)
	case &TIM8:
		return interrupt.New(stm32.IRQ_TIM8_UP_TIM13, TIM8.handleUPInterrupt)
	case &TIM9:
		return interrupt.New(stm32.IRQ_TIM1_BRK_TIM9, TIM9.handleUPInterrupt)
	case &TIM10:
		return interrupt.New(stm32.IRQ_TIM1_UP_TIM10, TIM10.handleUPInterrupt)
	case &TIM11:
		return interrupt.New(stm32.IRQ_TIM1_TRG_COM_TIM11, TIM11.handleUPInterrupt)
	case &TIM12:
		return interrupt.New(stm32.IRQ_TIM8_BRK_TIM12, TIM12.handleUPInterrupt)
	case &TIM13:
		return interrupt.New(stm32.IRQ_TIM8_UP_TIM13, TIM13.handleUPInterrupt)
	case &TIM14:
		return interrupt.New(stm32.IRQ_TIM8_TRG_COM_TIM14, TIM14.handleUPInterrupt)
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
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6_DAC, TIM6.handleOCInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleOCInterrupt)
	case &TIM8:
		return interrupt.New(stm32.IRQ_TIM8_UP_TIM13, TIM8.handleOCInterrupt)
	case &TIM9:
		return interrupt.New(stm32.IRQ_TIM1_BRK_TIM9, TIM9.handleOCInterrupt)
	case &TIM10:
		return interrupt.New(stm32.IRQ_TIM1_UP_TIM10, TIM10.handleOCInterrupt)
	case &TIM11:
		return interrupt.New(stm32.IRQ_TIM1_TRG_COM_TIM11, TIM11.handleOCInterrupt)
	case &TIM12:
		return interrupt.New(stm32.IRQ_TIM8_BRK_TIM12, TIM12.handleOCInterrupt)
	case &TIM13:
		return interrupt.New(stm32.IRQ_TIM8_UP_TIM13, TIM13.handleOCInterrupt)
	case &TIM14:
		return interrupt.New(stm32.IRQ_TIM8_TRG_COM_TIM14, TIM14.handleOCInterrupt)
	}

	return interrupt.Interrupt{}
}

func initRNG() {
	stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_RNGEN)
	stm32.RNG.CR.SetBits(stm32.RNG_CR_RNGEN)
}

func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var clock uint32
	switch uart.Bus {
	case stm32.USART1, stm32.USART6:
		clock = CPUFrequency() / 2 // APB2 Frequency
	case stm32.USART2, stm32.USART3, stm32.UART4, stm32.UART5:
		clock = CPUFrequency() / 4 // APB1 Frequency
	}
	return clock / baudRate
}
