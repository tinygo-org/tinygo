//go:build stm32f4
// +build stm32f4

package machine

// Peripheral abstraction layer for the stm32f4

import (
	"math/bits"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"

	"tinygo.org/x/device/stm32"
)

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

	PF0  = portF + 0
	PF1  = portF + 1
	PF2  = portF + 2
	PF3  = portF + 3
	PF4  = portF + 4
	PF5  = portF + 5
	PF6  = portF + 6
	PF7  = portF + 7
	PF8  = portF + 8
	PF9  = portF + 9
	PF10 = portF + 10
	PF11 = portF + 11
	PF12 = portF + 12
	PF13 = portF + 13
	PF14 = portF + 14
	PF15 = portF + 15

	PG0  = portG + 0
	PG1  = portG + 1
	PG2  = portG + 2
	PG3  = portG + 3
	PG4  = portG + 4
	PG5  = portG + 5
	PG6  = portG + 6
	PG7  = portG + 7
	PG8  = portG + 8
	PG9  = portG + 9
	PG10 = portG + 10
	PG11 = portG + 11
	PG12 = portG + 12
	PG13 = portG + 13
	PG14 = portG + 14
	PG15 = portG + 15

	PH0  = portH + 0
	PH1  = portH + 1
	PH2  = portH + 2
	PH3  = portH + 3
	PH4  = portH + 4
	PH5  = portH + 5
	PH6  = portH + 6
	PH7  = portH + 7
	PH8  = portH + 8
	PH9  = portH + 9
	PH10 = portH + 10
	PH11 = portH + 11
	PH12 = portH + 12
	PH13 = portH + 13
	PH14 = portH + 14
	PH15 = portH + 15

	PI0  = portI + 0
	PI1  = portI + 1
	PI2  = portI + 2
	PI3  = portI + 3
	PI4  = portI + 4
	PI5  = portI + 5
	PI6  = portI + 6
	PI7  = portI + 7
	PI8  = portI + 8
	PI9  = portI + 9
	PI10 = portI + 10
	PI11 = portI + 11
	PI12 = portI + 12
	PI13 = portI + 13
	PI14 = portI + 14
	PI15 = portI + 15

	PK0  = portK + 0
	PK1  = portK + 1
	PK2  = portK + 2
	PK3  = portK + 3
	PK4  = portK + 4
	PK5  = portK + 5
	PK6  = portK + 6
	PK7  = portK + 7
	PK8  = portK + 8
	PK9  = portK + 9
	PK10 = portK + 10
	PK11 = portK + 11
	PK12 = portK + 12
	PK13 = portK + 13
	PK14 = portK + 14
	PK15 = portK + 15
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

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	bit := p / 16
	if 0 <= bit && bit <= 10 {
		stm32.RCC.AHB1ENR.SetBits(0b1 << bit)
	} else {
		panic("machine: unknown port")
	}
}

func (p Pin) registerInterrupt() interrupt.Interrupt {
	pin := uint8(p) % 16

	switch pin {
	case 0:
		return interrupt.New(stm32.IRQ_EXTI0, func(interrupt.Interrupt) { handlePinInterrupt(0) })
	case 1:
		return interrupt.New(stm32.IRQ_EXTI1, func(interrupt.Interrupt) { handlePinInterrupt(1) })
	case 2:
		return interrupt.New(stm32.IRQ_EXTI2, func(interrupt.Interrupt) { handlePinInterrupt(2) })
	case 3:
		return interrupt.New(stm32.IRQ_EXTI3, func(interrupt.Interrupt) { handlePinInterrupt(3) })
	case 4:
		return interrupt.New(stm32.IRQ_EXTI4, func(interrupt.Interrupt) { handlePinInterrupt(4) })
	case 5:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(5) })
	case 6:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(6) })
	case 7:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(7) })
	case 8:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(8) })
	case 9:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(9) })
	case 10:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(10) })
	case 11:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(11) })
	case 12:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(12) })
	case 13:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(13) })
	case 14:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(14) })
	case 15:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(15) })
	}

	return interrupt.Interrupt{}
}

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.DAC): // DAC interface clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_DACEN)
	case unsafe.Pointer(stm32.PWR): // Power interface clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_PWREN)
	case unsafe.Pointer(stm32.CAN2): // CAN 2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_CAN2EN)
	case unsafe.Pointer(stm32.CAN1): // CAN 1 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_CAN1EN)
	case unsafe.Pointer(stm32.I2C3): // I2C3 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C3EN)
	case unsafe.Pointer(stm32.I2C2): // I2C2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C2EN)
	case unsafe.Pointer(stm32.I2C1): // I2C1 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)
	case unsafe.Pointer(stm32.UART5): // UART5 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_UART5EN)
	case unsafe.Pointer(stm32.UART4): // UART4 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_UART4EN)
	case unsafe.Pointer(stm32.USART3): // USART3 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART3EN)
	case unsafe.Pointer(stm32.USART2): // USART2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)
	case unsafe.Pointer(stm32.SPI3): // SPI3 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_SPI3EN)
	case unsafe.Pointer(stm32.SPI2): // SPI2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_SPI2EN)
	case unsafe.Pointer(stm32.WWDG): // Window watchdog clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_WWDGEN)
	case unsafe.Pointer(stm32.TIM14): // TIM14 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM14EN)
	case unsafe.Pointer(stm32.TIM13): // TIM13 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM13EN)
	case unsafe.Pointer(stm32.TIM12): // TIM12 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM12EN)
	case unsafe.Pointer(stm32.TIM7): // TIM7 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM7EN)
	case unsafe.Pointer(stm32.TIM6): // TIM6 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM6EN)
	case unsafe.Pointer(stm32.TIM5): // TIM5 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM5EN)
	case unsafe.Pointer(stm32.TIM4): // TIM4 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM4EN)
	case unsafe.Pointer(stm32.TIM3): // TIM3 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM3EN)
	case unsafe.Pointer(stm32.TIM2): // TIM2 clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_TIM2EN)
	case unsafe.Pointer(stm32.TIM11): // TIM11 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM11EN)
	case unsafe.Pointer(stm32.TIM10): // TIM10 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM10EN)
	case unsafe.Pointer(stm32.TIM9): // TIM9 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM9EN)
	case unsafe.Pointer(stm32.SYSCFG): // System configuration controller clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SYSCFGEN)
	case unsafe.Pointer(stm32.SPI1): // SPI1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	case unsafe.Pointer(stm32.SDIO): // SDIO clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SDIOEN)
	case unsafe.Pointer(stm32.ADC3): // ADC3 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC3EN)
	case unsafe.Pointer(stm32.ADC2): // ADC2 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC2EN)
	case unsafe.Pointer(stm32.ADC1): // ADC1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADC1EN)
	case unsafe.Pointer(stm32.USART6): // USART6 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART6EN)
	case unsafe.Pointer(stm32.USART1): // USART1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	case unsafe.Pointer(stm32.TIM8): // TIM8 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM8EN)
	case unsafe.Pointer(stm32.TIM1): // TIM1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM1EN)
	}
}

//---------- Timer related code

var (
	TIM1 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM1EN,
		Device:         stm32.TIM1,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA8, AF1_TIM1_2}, {PE9, AF1_TIM1_2}}},
			TimerChannel{Pins: []PinFunction{{PA9, AF1_TIM1_2}, {PE11, AF1_TIM1_2}}},
			TimerChannel{Pins: []PinFunction{{PA10, AF1_TIM1_2}, {PE13, AF1_TIM1_2}}},
			TimerChannel{Pins: []PinFunction{{PA11, AF1_TIM1_2}, {PE14, AF1_TIM1_2}}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA0, AF1_TIM1_2}, {PA5, AF1_TIM1_2}, {PA15, AF1_TIM1_2}}},
			TimerChannel{Pins: []PinFunction{{PA1, AF1_TIM1_2}, {PB3, AF1_TIM1_2}}},
			TimerChannel{Pins: []PinFunction{{PA2, AF1_TIM1_2}, {PB10, AF1_TIM1_2}}},
			TimerChannel{Pins: []PinFunction{{PA3, AF1_TIM1_2}, {PB11, AF1_TIM1_2}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA6, AF2_TIM3_4_5}, {PB4, AF2_TIM3_4_5}, {PC6, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PA7, AF2_TIM3_4_5}, {PB5, AF2_TIM3_4_5}, {PC7, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PB0, AF2_TIM3_4_5}, {PC8, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PB1, AF2_TIM3_4_5}, {PC9, AF2_TIM3_4_5}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM4 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM4EN,
		Device:         stm32.TIM4,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PB6, AF2_TIM3_4_5}, {PD12, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PB7, AF2_TIM3_4_5}, {PD13, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PB8, AF2_TIM3_4_5}, {PD14, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PB9, AF2_TIM3_4_5}, {PD15, AF2_TIM3_4_5}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM5 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM5EN,
		Device:         stm32.TIM5,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PH10, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PH11, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PH12, AF2_TIM3_4_5}}},
			TimerChannel{Pins: []PinFunction{{PI0, AF2_TIM3_4_5}}},
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

	TIM9 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM9EN,
		Device:         stm32.TIM9,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA2, AF3_TIM8_9_10_11}, {PE5, AF3_TIM8_9_10_11}}},
			TimerChannel{Pins: []PinFunction{{PA3, AF3_TIM8_9_10_11}, {PE6, AF3_TIM8_9_10_11}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM10 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM10EN,
		Device:         stm32.TIM10,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PB8, AF3_TIM8_9_10_11}, {PF6, AF3_TIM8_9_10_11}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM11 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM11EN,
		Device:         stm32.TIM11,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PB9, AF3_TIM8_9_10_11}, {PF7, AF3_TIM8_9_10_11}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
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

func (t *TIM) enableMainOutput() {
	t.Device.BDTR.SetBits(stm32.TIM_BDTR_MOE)
}

type arrtype = uint32
type arrRegType = volatile.Register32

const (
	ARR_MAX = 0x10000
	PSC_MAX = 0x10000
)

func initRNG() {
	stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_RNGEN)
	stm32.RNG.CR.SetBits(stm32.RNG_CR_RNGEN)
}

// Alternative peripheral pin functions
const (
	AF0_SYSTEM                = 0
	AF1_TIM1_2                = 1
	AF2_TIM3_4_5              = 2
	AF3_TIM8_9_10_11          = 3
	AF4_I2C1_2_3              = 4
	AF5_SPI1_SPI2             = 5
	AF6_SPI3                  = 6
	AF7_USART1_2_3            = 7
	AF8_USART4_5_6            = 8
	AF9_CAN1_CAN2_TIM12_13_14 = 9
	AF10_OTG_FS_OTG_HS        = 10
	AF11_ETH                  = 11
	AF12_FSMC_SDIO_OTG_HS_1   = 12
	AF13_DCMI                 = 13
	AF14                      = 14
	AF15_EVENTOUT             = 15
)

// -- UART ---------------------------------------------------------------------

func (uart *UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
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

func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.DR
	uart.txReg = &uart.Bus.DR
	uart.statusReg = &uart.Bus.SR
	uart.txEmptyFlag = stm32.USART_SR_TXE
}

// -- SPI ----------------------------------------------------------------------

type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

func (spi SPI) config8Bits() {
	// no-op on this series
}

func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var clock uint32
	switch spi.Bus {
	case stm32.SPI1:
		clock = CPUFrequency() / 2
	case stm32.SPI2, stm32.SPI3:
		clock = CPUFrequency() / 4
	}

	// limit requested frequency to bus frequency and min frequency (DIV256)
	freq := config.Frequency
	if min := clock / 256; freq < min {
		freq = min
	} else if freq > clock {
		freq = clock
	}

	// calculate the exact clock divisor (freq=clock/div -> div=clock/freq).
	// truncation is fine, since it produces a less-than-or-equal divisor, and
	// thus a greater-than-or-equal frequency.
	// divisors only come in consecutive powers of 2, so we can use log2 (or,
	// equivalently, bits.Len - 1) to convert to respective enum value.
	div := bits.Len32(clock/freq) - 1

	// but DIV1 (2^0) is not permitted, as the least divisor is DIV2 (2^1), so
	// subtract 1 from the log2 value, keeping a lower bound of 0
	if div < 0 {
		div = 0
	} else if div > 0 {
		div--
	}

	// finally, shift the enumerated value into position for SPI CR1
	return uint32(div) << stm32.SPI_CR1_BR_Pos
}

// -- I2C ----------------------------------------------------------------------

type I2C struct {
	Bus             *stm32.I2C_Type
	AltFuncSelector uint8
}

func (i2c *I2C) configurePins(config I2CConfig) {
	config.SCL.ConfigureAltFunc(PinConfig{Mode: PinModeI2CSCL}, i2c.AltFuncSelector)
	config.SDA.ConfigureAltFunc(PinConfig{Mode: PinModeI2CSDA}, i2c.AltFuncSelector)
}

func (i2c *I2C) getFreqRange(config I2CConfig) uint32 {
	// all I2C interfaces are on APB1
	clock := CPUFrequency() / 4
	// convert to MHz
	clock /= 1000000
	// must be between 2 MHz (or 4 MHz for fast mode (Fm)) and 50 MHz, inclusive
	var min, max uint32 = 2, 50
	if config.Frequency > 100000 {
		min = 4 // fast mode (Fm)
	}
	if clock < min {
		clock = min
	} else if clock > max {
		clock = max
	}
	return clock << stm32.I2C_CR2_FREQ_Pos
}

func (i2c *I2C) getRiseTime(config I2CConfig) uint32 {
	// These bits must be programmed with the maximum SCL rise time given in the
	// I2C bus specification, incremented by 1.
	// For instance: in Sm mode, the maximum allowed SCL rise time is 1000 ns.
	// If, in the I2C_CR2 register, the value of FREQ[5:0] bits is equal to 0x08
	// and PCLK1 = 125 ns, therefore the TRISE[5:0] bits must be programmed with
	// 09h (1000 ns / 125 ns = 8 + 1)
	freqRange := i2c.getFreqRange(config)
	if config.Frequency > 100000 {
		// fast mode (Fm) adjustment
		freqRange *= 300
		freqRange /= 1000
	}
	return (freqRange + 1) << stm32.I2C_TRISE_TRISE_Pos
}

func (i2c *I2C) getSpeed(config I2CConfig) uint32 {
	ccr := func(pclk uint32, freq uint32, coeff uint32) uint32 {
		return (((pclk - 1) / (freq * coeff)) + 1) & stm32.I2C_CCR_CCR_Msk
	}
	sm := func(pclk uint32, freq uint32) uint32 { // standard mode (Sm)
		if s := ccr(pclk, freq, 2); s < 4 {
			return 4
		} else {
			return s
		}
	}
	fm := func(pclk uint32, freq uint32, duty uint8) uint32 { // fast mode (Fm)
		if duty == DutyCycle2 {
			return ccr(pclk, freq, 3)
		} else {
			return ccr(pclk, freq, 25) | stm32.I2C_CCR_DUTY
		}
	}
	// all I2C interfaces are on APB1
	clock := CPUFrequency() / 4
	if config.Frequency <= 100000 {
		return sm(clock, config.Frequency)
	} else {
		s := fm(clock, config.Frequency, config.DutyCycle)
		if (s & stm32.I2C_CCR_CCR_Msk) == 0 {
			return 1
		} else {
			return s | stm32.I2C_CCR_F_S
		}
	}
}
