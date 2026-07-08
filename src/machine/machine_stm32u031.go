//go:build stm32u031

package machine

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

var deviceIDAddr = []uintptr{0x1FFF3E50, 0x1FFF3E54, 0x1FFF3E58}

// Pin constants for all stm32u0 package sizes
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
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_GPIOAEN)
	case 1:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_GPIOBEN)
	case 2:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_GPIOCEN)
	case 3:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_GPIODEN)
	case 4:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_GPIOEEN)
	case 5:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_GPIOFEN)
	default:
		panic("machine: unknown port")
	}
}

func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.I2C1):
		stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_I2C1EN)
	case unsafe.Pointer(stm32.I2C2):
		stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_I2C2EN)
	case unsafe.Pointer(stm32.I2C3):
		stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_I2C3EN)
	case unsafe.Pointer(stm32.USART1):
		stm32.RCC.APBENR2.SetBits(stm32.RCC_APBENR2_USART1EN)
	case unsafe.Pointer(stm32.USART2):
		stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_USART2EN)
	case unsafe.Pointer(stm32.USART3):
		stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_USART3EN)
	case unsafe.Pointer(stm32.USART4):
		stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_USART4EN)
	}
}

var (
	// TODO: implement output channels

	TIM1 = TIM{
		EnableRegister: &stm32.RCC.APBENR2,
		EnableFlag:     stm32.RCC_APBENR2_TIM1EN,
		Device:         stm32.TIM1,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: 4e6,
	}

	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APBENR1,
		EnableFlag:     stm32.RCC_APBENR1_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: 4e6,
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APBENR1,
		EnableFlag:     stm32.RCC_APBENR1_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: 4e6,
	}

	TIM6 = TIM{
		EnableRegister: &stm32.RCC.APBENR1,
		EnableFlag:     stm32.RCC_APBENR1_TIM6EN,
		Device:         stm32.TIM6,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: 4e6,
	}

	TIM7 = TIM{
		EnableRegister: &stm32.RCC.APBENR1,
		EnableFlag:     stm32.RCC_APBENR1_TIM7EN,
		Device:         stm32.TIM7,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: 4e6,
	}

	TIM15 = TIM{
		EnableRegister: &stm32.RCC.APBENR2,
		EnableFlag:     stm32.RCC_APBENR2_TIM15EN,
		Device:         stm32.TIM15,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: 4e6,
	}

	TIM16 = TIM{
		EnableRegister: &stm32.RCC.APBENR2,
		EnableFlag:     stm32.RCC_APBENR2_TIM16EN,
		Device:         stm32.TIM16,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: 4e6,
	}
)

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_BRK_UP_TRG_COM, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleUPInterrupt)
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6_DAC_LPTIM1, TIM6.handleUPInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7_LPTIM2, TIM7.handleUPInterrupt)
	case &TIM15:
		return interrupt.New(stm32.IRQ_TIM15_LPTIM3, TIM15.handleUPInterrupt)
	case &TIM16:
		return interrupt.New(stm32.IRQ_TIM16, TIM16.handleUPInterrupt)
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
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6_DAC_LPTIM1, TIM6.handleOCInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7_LPTIM2, TIM7.handleOCInterrupt)
	case &TIM15:
		return interrupt.New(stm32.IRQ_TIM15_LPTIM3, TIM15.handleOCInterrupt)
	case &TIM16:
		return interrupt.New(stm32.IRQ_TIM16, TIM16.handleOCInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) enableMainOutput() {
	t.Device.BDTR.SetBits(stm32.TIM_BDTR_MOE)
}

type arrtype = uint16
type psctype = uint16
type arrRegType = volatile.Register16

const (
	ARR_MAX = 0x10000
	PSC_MAX = 0x10000
)

const (
	UART_TX_PIN = NoPin
	UART_RX_PIN = NoPin
)

func (uart *UART) configurePins(config UARTConfig) {
	panic("unimplemented: UART")
}

func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	panic("unimplemented: UART")
}

func (uart *UART) setRegisters() uint32 {
	panic("unimplemented: UART")
}

const (
	I2C0_SCL_PIN = NoPin
	I2C0_SDA_PIN = NoPin
)

func (i2c I2C) getFreqRange(br uint32) uint32 {
	panic("unimplemented: I2C")
}
