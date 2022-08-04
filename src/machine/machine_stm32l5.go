//go:build stm32l5
// +build stm32l5

package machine

// Peripheral abstraction layer for the stm32l5

import (
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"

	"tinygo.org/x/device/stm32"
)

const (
	AF0_SYSTEM                                = 0
	AF1_TIM1_2_5_8_LPTIM1                     = 1
	AF2_TIM1_2_3_4_5_LPTIM3                   = 2
	AF3_SPI2_SAI1_I2C4_USART2_TIM1_8_OCTOSPI1 = 3
	AF4_I2C1_2_3_4                            = 4
	AF5_SPI1_2_3_I2C4_DFSDM1_OCTOSPI1         = 5
	AF6_SPI3_I2C3_DFSDM1_COMP1                = 6
	AF7_USART1_2_3                            = 7
	AF8_UART4_5_LPUART1_SDMMC1                = 8
	AF9_FDCAN1_TSC                            = 9
	AF10_USB_OCTOSPI1                         = 10
	AF11_UCPD1                                = 11
	AF12_SDMMC1_COMP1_2_TIM1_8_FMC            = 12
	AF13_SAI1_2_TIM8                          = 13
	AF14_TIM2_8_15_16_17_LPTIM2               = 14
	AF15_EVENTOUT                             = 15
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

	PH0 = portH + 0
	PH1 = portH + 1
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
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOAEN)
	case 1:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOBEN)
	case 2:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOCEN)
	case 3:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIODEN)
	case 4:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOEEN)
	case 5:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOFEN)
	case 6:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOGEN)
	case 7:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOHEN)
	default:
		panic("machine: unknown port")
	}
}

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.DAC): // DAC interface clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_DAC1EN)
	case unsafe.Pointer(stm32.PWR): // Power interface clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_PWREN)
	case unsafe.Pointer(stm32.I2C3): // I2C3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C3EN)
	case unsafe.Pointer(stm32.I2C2): // I2C2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C2EN)
	case unsafe.Pointer(stm32.I2C1): // I2C1 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C1EN)
	case unsafe.Pointer(stm32.UART5): // UART5 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_UART5EN)
	case unsafe.Pointer(stm32.UART4): // UART4 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_UART4EN)
	case unsafe.Pointer(stm32.USART3): // USART3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_USART3EN)
	case unsafe.Pointer(stm32.USART2): // USART2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_USART2EN)
	case unsafe.Pointer(stm32.SPI3): // SPI3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_SP3EN)
	case unsafe.Pointer(stm32.SPI2): // SPI2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_SPI2EN)
	case unsafe.Pointer(stm32.WWDG): // Window watchdog clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_WWDGEN)
	case unsafe.Pointer(stm32.TIM7): // TIM7 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM7EN)
	case unsafe.Pointer(stm32.TIM6): // TIM6 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM6EN)
	case unsafe.Pointer(stm32.TIM5): // TIM5 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM5EN)
	case unsafe.Pointer(stm32.TIM4): // TIM4 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM4EN)
	case unsafe.Pointer(stm32.TIM3): // TIM3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM3EN)
	case unsafe.Pointer(stm32.TIM2): // TIM2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM2EN)
	case unsafe.Pointer(stm32.UCPD1): // UCPD1 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_UCPD1EN)
	case unsafe.Pointer(stm32.FDCAN1): // FDCAN1 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_FDCAN1EN)
	case unsafe.Pointer(stm32.LPTIM3): // LPTIM3 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_LPTIM3EN)
	case unsafe.Pointer(stm32.LPTIM2): // LPTIM2 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_LPTIM2EN)
	case unsafe.Pointer(stm32.I2C4): // I2C4 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_I2C4EN)
	case unsafe.Pointer(stm32.LPUART1): // LPUART1 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_LPUART1EN)
	case unsafe.Pointer(stm32.TIM17): // TIM17 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM17EN)
	case unsafe.Pointer(stm32.TIM16): // TIM16 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM16EN)
	case unsafe.Pointer(stm32.TIM15): // TIM15 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM15EN)
	case unsafe.Pointer(stm32.SYSCFG): // System configuration controller clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SYSCFGEN)
	case unsafe.Pointer(stm32.SPI1): // SPI1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	case unsafe.Pointer(stm32.USART1): // USART1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	case unsafe.Pointer(stm32.TIM8): // TIM8 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM8EN)
	case unsafe.Pointer(stm32.TIM1): // TIM1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM1EN)
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
		return interrupt.New(stm32.IRQ_EXTI5, func(interrupt.Interrupt) { handlePinInterrupt(5) })
	case 6:
		return interrupt.New(stm32.IRQ_EXTI6, func(interrupt.Interrupt) { handlePinInterrupt(6) })
	case 7:
		return interrupt.New(stm32.IRQ_EXTI7, func(interrupt.Interrupt) { handlePinInterrupt(7) })
	case 8:
		return interrupt.New(stm32.IRQ_EXTI8, func(interrupt.Interrupt) { handlePinInterrupt(8) })
	case 9:
		return interrupt.New(stm32.IRQ_EXTI9, func(interrupt.Interrupt) { handlePinInterrupt(9) })
	case 10:
		return interrupt.New(stm32.IRQ_EXTI10, func(interrupt.Interrupt) { handlePinInterrupt(10) })
	case 11:
		return interrupt.New(stm32.IRQ_EXTI11, func(interrupt.Interrupt) { handlePinInterrupt(11) })
	case 12:
		return interrupt.New(stm32.IRQ_EXTI12, func(interrupt.Interrupt) { handlePinInterrupt(12) })
	case 13:
		return interrupt.New(stm32.IRQ_EXTI13, func(interrupt.Interrupt) { handlePinInterrupt(13) })
	case 14:
		return interrupt.New(stm32.IRQ_EXTI14, func(interrupt.Interrupt) { handlePinInterrupt(14) })
	case 15:
		return interrupt.New(stm32.IRQ_EXTI15, func(interrupt.Interrupt) { handlePinInterrupt(15) })
	}

	return interrupt.Interrupt{}
}

func handlePinInterrupt(pin uint8) {
	// The pin abstraction doesn't differentiate pull-up
	// events from pull-down events, so combine them to
	// a single call here.

	if stm32.EXTI.RPR1.HasBits(1<<pin) || stm32.EXTI.FPR1.HasBits(1<<pin) {
		// Writing 1 to the pending register clears the
		// pending flag for that bit
		stm32.EXTI.RPR1.Set(1 << pin)
		stm32.EXTI.FPR1.Set(1 << pin)

		callback := pinCallbacks[pin]
		if callback != nil {
			callback(interruptPins[pin])
		}
	}
}

//---------- Timer related code

var (
	TIM1 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM1EN,
		Device:         stm32.TIM1,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA8, AF1_TIM1_2_5_8_LPTIM1}, {PE9, AF1_TIM1_2_5_8_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA9, AF1_TIM1_2_5_8_LPTIM1}, {PE11, AF1_TIM1_2_5_8_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA10, AF1_TIM1_2_5_8_LPTIM1}, {PE13, AF1_TIM1_2_5_8_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA11, AF1_TIM1_2_5_8_LPTIM1}, {PE14, AF1_TIM1_2_5_8_LPTIM1}}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA0, AF1_TIM1_2_5_8_LPTIM1}, {PA5, AF1_TIM1_2_5_8_LPTIM1}, {PA15, AF1_TIM1_2_5_8_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA1, AF1_TIM1_2_5_8_LPTIM1}, {PB3, AF1_TIM1_2_5_8_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA2, AF1_TIM1_2_5_8_LPTIM1}, {PB10, AF1_TIM1_2_5_8_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA3, AF1_TIM1_2_5_8_LPTIM1}, {PB11, AF1_TIM1_2_5_8_LPTIM1}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA6, AF2_TIM1_2_3_4_5_LPTIM3}, {PB4, AF2_TIM1_2_3_4_5_LPTIM3}, {PC6, AF2_TIM1_2_3_4_5_LPTIM3}, {PE3, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PA7, AF2_TIM1_2_3_4_5_LPTIM3}, {PB5, AF2_TIM1_2_3_4_5_LPTIM3}, {PC7, AF2_TIM1_2_3_4_5_LPTIM3}, {PE4, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PB0, AF2_TIM1_2_3_4_5_LPTIM3}, {PC8, AF2_TIM1_2_3_4_5_LPTIM3}, {PE5, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PB1, AF2_TIM1_2_3_4_5_LPTIM3}, {PC9, AF2_TIM1_2_3_4_5_LPTIM3}, {PE6, AF2_TIM1_2_3_4_5_LPTIM3}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM4 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM4EN,
		Device:         stm32.TIM4,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PB6, AF2_TIM1_2_3_4_5_LPTIM3}, {PD12, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PB7, AF2_TIM1_2_3_4_5_LPTIM3}, {PD13, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PB8, AF2_TIM1_2_3_4_5_LPTIM3}, {PD14, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PB9, AF2_TIM1_2_3_4_5_LPTIM3}, {PD15, AF2_TIM1_2_3_4_5_LPTIM3}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM5 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM5EN,
		Device:         stm32.TIM5,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA0, AF2_TIM1_2_3_4_5_LPTIM3}, {PF6, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PA1, AF2_TIM1_2_3_4_5_LPTIM3}, {PF7, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PA2, AF2_TIM1_2_3_4_5_LPTIM3}, {PF8, AF2_TIM1_2_3_4_5_LPTIM3}}},
			TimerChannel{Pins: []PinFunction{{PA3, AF2_TIM1_2_3_4_5_LPTIM3}, {PF9, AF2_TIM1_2_3_4_5_LPTIM3}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM6 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM6EN,
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
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM7EN,
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
			TimerChannel{Pins: []PinFunction{{PC6, AF3_SPI2_SAI1_I2C4_USART2_TIM1_8_OCTOSPI1}}},
			TimerChannel{Pins: []PinFunction{{PC7, AF3_SPI2_SAI1_I2C4_USART2_TIM1_8_OCTOSPI1}}},
			TimerChannel{Pins: []PinFunction{{PC8, AF3_SPI2_SAI1_I2C4_USART2_TIM1_8_OCTOSPI1}}},
			TimerChannel{Pins: []PinFunction{{PC9, AF3_SPI2_SAI1_I2C4_USART2_TIM1_8_OCTOSPI1}}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM15 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM15EN,
		Device:         stm32.TIM15,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA1, AF14_TIM2_8_15_16_17_LPTIM2}, {PB14, AF14_TIM2_8_15_16_17_LPTIM2}, {PF9, AF14_TIM2_8_15_16_17_LPTIM2}, {PG10, AF14_TIM2_8_15_16_17_LPTIM2}}},
			TimerChannel{Pins: []PinFunction{{PA2, AF14_TIM2_8_15_16_17_LPTIM2}, {PB15, AF14_TIM2_8_15_16_17_LPTIM2}, {PF10, AF14_TIM2_8_15_16_17_LPTIM2}, {PG11, AF14_TIM2_8_15_16_17_LPTIM2}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM16 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM16EN,
		Device:         stm32.TIM16,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA6, AF14_TIM2_8_15_16_17_LPTIM2}, {PB8, AF14_TIM2_8_15_16_17_LPTIM2}, {PE0, AF14_TIM2_8_15_16_17_LPTIM2}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM17 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM17EN,
		Device:         stm32.TIM17,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA7, AF14_TIM2_8_15_16_17_LPTIM2}, {PB9, AF14_TIM2_8_15_16_17_LPTIM2}, {PE1, AF14_TIM2_8_15_16_17_LPTIM2}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}
)

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_UP, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleUPInterrupt)
	case &TIM4:
		return interrupt.New(stm32.IRQ_TIM4, TIM4.handleUPInterrupt)
	case &TIM5:
		return interrupt.New(stm32.IRQ_TIM5, TIM5.handleUPInterrupt)
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6, TIM6.handleUPInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleUPInterrupt)
	case &TIM8:
		return interrupt.New(stm32.IRQ_TIM8_UP, TIM8.handleUPInterrupt)
	case &TIM15:
		return interrupt.New(stm32.IRQ_TIM15, TIM15.handleUPInterrupt)
	case &TIM16:
		return interrupt.New(stm32.IRQ_TIM16, TIM16.handleUPInterrupt)
	case &TIM17:
		return interrupt.New(stm32.IRQ_TIM17, TIM17.handleUPInterrupt)
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
		return interrupt.New(stm32.IRQ_TIM6, TIM6.handleOCInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleOCInterrupt)
	case &TIM8:
		return interrupt.New(stm32.IRQ_TIM8_CC, TIM8.handleOCInterrupt)
	case &TIM15:
		return interrupt.New(stm32.IRQ_TIM15, TIM15.handleOCInterrupt)
	case &TIM16:
		return interrupt.New(stm32.IRQ_TIM16, TIM16.handleOCInterrupt)
	case &TIM17:
		return interrupt.New(stm32.IRQ_TIM17, TIM17.handleOCInterrupt)
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
	stm32.RCC.CRRCR.SetBits(stm32.RCC_CRRCR_HSI48ON)
	for !stm32.RCC.CRRCR.HasBits(stm32.RCC_CRRCR_HSI48RDY) {
	}

	stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_RNGEN)
	stm32.RNG.CR.SetBits(stm32.RNG_CR_RNGEN)
}
