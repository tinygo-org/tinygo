//go:build stm32l0x1
// +build stm32l0x1

package machine

// Peripheral abstraction layer for the stm32l0

import (
	"runtime/interrupt"
	"runtime/volatile"
	"tinygo.org/x/device/stm32"
	"unsafe"
)

const (
	AF0_SYSTEM_SPI1_USART2_LPTIM_TIM21 = 0
	AF1_SPI1_I2C1_LPTIM                = 1
	AF2_LPTIM_TIM2                     = 2
	AF3_I2C1                           = 3
	AF4_I2C1_USART2_LPUART1_TIM22      = 4
	AF5_TIM2_21_22                     = 5
	AF6_LPUART1                        = 6
	AF7_COMP1_2                        = 7
)

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.PWR): // Power interface clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	case unsafe.Pointer(stm32.I2C3): // I2C3 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C3EN)
	case unsafe.Pointer(stm32.I2C2): // I2C2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C2EN)
	case unsafe.Pointer(stm32.I2C1): // I2C1 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)
	case unsafe.Pointer(stm32.USART5): // UART5 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART5EN)
	case unsafe.Pointer(stm32.USART4): // UART4 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART4EN)
	case unsafe.Pointer(stm32.USART2): // USART2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)
	case unsafe.Pointer(stm32.SPI2): // SPI2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_SPI2EN)
	case unsafe.Pointer(stm32.LPUART1): // LPUART1 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_LPUART1EN)
	case unsafe.Pointer(stm32.WWDG): // Window watchdog clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_WWDGEN)
	case unsafe.Pointer(stm32.TIM7): // TIM7 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM7EN)
	case unsafe.Pointer(stm32.TIM6): // TIM6 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM6EN)
	case unsafe.Pointer(stm32.TIM3): // TIM3 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM3EN)
	case unsafe.Pointer(stm32.TIM2): // TIM2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM2EN)
	case unsafe.Pointer(stm32.SYSCFG): // System configuration controller clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SYSCFGEN)
	case unsafe.Pointer(stm32.SPI1): // SPI1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	case unsafe.Pointer(stm32.ADC): // ADC clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADCEN)
	case unsafe.Pointer(stm32.USART1): // USART1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	}
}

//---------- Timer related code

var (
	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA0, AF2_LPTIM_TIM2}, {PA5, AF5_TIM2_21_22}, {PA8, AF5_TIM2_21_22}, {PA15, AF5_TIM2_21_22}}},
			TimerChannel{Pins: []PinFunction{{PA1, AF2_LPTIM_TIM2}, {PB3, AF2_LPTIM_TIM2}}},
			TimerChannel{Pins: []PinFunction{{PA2, AF2_LPTIM_TIM2}, {PB0, AF5_TIM2_21_22}, {PB10, AF2_LPTIM_TIM2}}},
			TimerChannel{Pins: []PinFunction{{PA3, AF2_LPTIM_TIM2}, {PB1, AF5_TIM2_21_22}, {PB11, AF2_LPTIM_TIM2}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

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

	TIM21 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM21EN,
		Device:         stm32.TIM21,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM22 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM22EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}
)

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleUPInterrupt)
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6, TIM6.handleUPInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleUPInterrupt)
	case &TIM21:
		return interrupt.New(stm32.IRQ_TIM21, TIM21.handleUPInterrupt)
	case &TIM22:
		return interrupt.New(stm32.IRQ_TIM22, TIM22.handleUPInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) registerOCInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleOCInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleOCInterrupt)
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6, TIM6.handleOCInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleOCInterrupt)
	case &TIM21:
		return interrupt.New(stm32.IRQ_TIM21, TIM21.handleOCInterrupt)
	case &TIM22:
		return interrupt.New(stm32.IRQ_TIM22, TIM22.handleOCInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) enableMainOutput() {
	// nothing to do - no BDTR register
}

type arrtype = uint16
type arrRegType = volatile.Register16

const (
	ARR_MAX = 0x10000
	PSC_MAX = 0x10000
)
