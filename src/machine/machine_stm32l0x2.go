//go:build stm32l0x2
// +build stm32l0x2

package machine

// Peripheral abstraction layer for the stm32l0

import (
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"

	"tinygo.org/x/device/stm32"
)

const (
	AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22 = 0
	AF1_SPI1_2_I2S2_I2C1_TIM2_21                                      = 1
	AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3                  = 2
	AF3_I2C1_TSC                                                      = 3
	AF4_I2C1_USART1_2_LPUART1_TIM3_22                                 = 4
	AF5_SPI2_I2S2_I2C2_USART1_TIM2_21_22                              = 5
	AF6_I2C1_2_LPUART1_USART4_5_TIM21                                 = 6
	AF7_I2C3_LPUART1_COMP1_2_TIM3                                     = 7
)

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.DAC): // DAC interface clock enable
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_DACEN)
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
			TimerChannel{Pins: []PinFunction{
				{PA0, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PA5, AF5_SPI2_I2S2_I2C2_USART1_TIM2_21_22},
				{PA15, AF5_SPI2_I2S2_I2C2_USART1_TIM2_21_22},
				{PE9, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA1, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PB3, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PE10, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA2, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PB10, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PE11, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA3, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PB11, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PE12, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
			}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{
				{PA6, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PB4, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PC6, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PE3, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA7, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PB5, AF4_I2C1_USART1_2_LPUART1_TIM3_22},
				{PC7, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PE4, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
			}},
			TimerChannel{Pins: []PinFunction{
				{PB0, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PC8, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PE5, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
			}},
			TimerChannel{Pins: []PinFunction{
				{PB1, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PC9, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
				{PE6, AF2_SPI1_2_I2S2_LPUART1_USART5_USB_LPTIM1_TIM2_3},
			}},
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
			TimerChannel{Pins: []PinFunction{
				{PA2, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
				{PB13, AF6_I2C1_2_LPUART1_USART4_5_TIM21},
				{PD0, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
				{PE5, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA3, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
				{PB14, AF6_I2C1_2_LPUART1_USART4_5_TIM21},
				{PD7, AF1_SPI1_2_I2S2_I2C1_TIM2_21},
				{PE6, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
			}},
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
			TimerChannel{Pins: []PinFunction{
				{PA6, AF5_SPI2_I2S2_I2C2_USART1_TIM2_21_22},
				{PB4, AF4_I2C1_USART1_2_LPUART1_TIM3_22},
				{PC6, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
				{PE3, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
			}},
			TimerChannel{Pins: []PinFunction{
				{PA7, AF5_SPI2_I2S2_I2C2_USART1_TIM2_21_22},
				{PB5, AF4_I2C1_USART1_2_LPUART1_TIM3_22},
				{PC7, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
				{PE4, AF0_SYSTEM_SPI1_2_I2S2_USART1_2_LPUART1_USB_LPTIM1_TSC_TIM2_21_22},
			}},
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
		return interrupt.New(stm32.IRQ_TIM6_DAC, TIM6.handleUPInterrupt)
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
		return interrupt.New(stm32.IRQ_TIM6_DAC, TIM6.handleOCInterrupt)
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

func initRNG() {
	stm32.RCC.AHBENR.SetBits(stm32.RCC_AHBENR_RNGEN)
	stm32.RNG.CR.SetBits(stm32.RNG_CR_RNGEN)
}
