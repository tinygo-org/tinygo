// +build stm32
// +build stm32f405 stm32f407

package machine

// Peripheral abstraction layer for the stm32f40x

import (
	"device/stm32"
	"runtime/interrupt"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 168_000_000
}

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
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOAEN)
	case 1:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOBEN)
	case 2:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOCEN)
	case 3:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIODEN)
	case 4:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOEEN)
	case 5:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOFEN)
	case 6:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOGEN)
	case 7:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOHEN)
	case 8:
		stm32.RCC.AHB1ENR.SetBits(stm32.RCC_AHB1ENR_GPIOIEN)
	default:
		panic("machine: unknown port")
	}
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

//---------- UART related types and code

// UART representation
type UART struct {
	Buffer          *RingBuffer
	Bus             *stm32.USART_Type
	Interrupt       interrupt.Interrupt
	AltFuncSelector stm32.AltFunc
}

// Configure the UART.
func (uart UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.AltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.AltFuncSelector)
}

// UART baudrate calc based on the bus and clockspeed
// NOTE: keep this in sync with the runtime/runtime_stm32f40x.go clock init code
func (uart UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var clock uint32
	switch uart.Bus {
	case stm32.USART1, stm32.USART6:
		clock = CPUFrequency() / 2 // APB2 Frequency
	case stm32.USART2, stm32.USART3, stm32.UART4, stm32.UART5:
		clock = CPUFrequency() / 4 // APB1 Frequency
	}
	return clock / baudRate
}

//---------- SPI related types and code

// SPI on the STM32Fxxx using MODER / alternate function pins
type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector stm32.AltFunc
}

// Set baud rate for SPI
func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	localFrequency := config.Frequency
	if spi.Bus != stm32.SPI1 {
		// Assume it's SPI2 or SPI3 on APB1 at 1/2 the clock frequency of APB2, so
		//  we want to pretend to request 2x the baudrate asked for
		localFrequency = localFrequency * 2
	}

	// set frequency dependent on PCLK prescaler. Since these are rather weird
	// speeds due to the CPU freqency, pick a range up to that frquency for
	// clients to use more human-understandable numbers, e.g. nearest 100KHz

	// These are based on APB2 clock frquency (84MHz on all stm32f40x boards)
	// TODO: also include the MCU/APB clock setting in the equation
	switch true {
	case localFrequency < 328125:
		conf = stm32.SPI_PCLK_256
	case localFrequency < 656250:
		conf = stm32.SPI_PCLK_128
	case localFrequency < 1312500:
		conf = stm32.SPI_PCLK_64
	case localFrequency < 2625000:
		conf = stm32.SPI_PCLK_32
	case localFrequency < 5250000:
		conf = stm32.SPI_PCLK_16
	case localFrequency < 10500000:
		conf = stm32.SPI_PCLK_8
		// NOTE: many SPI components won't operate reliably (or at all) above 10MHz
		// Check the datasheet of the part
	case localFrequency < 21000000:
		conf = stm32.SPI_PCLK_4
	case localFrequency < 42000000:
		conf = stm32.SPI_PCLK_2
	default:
		// None of the specific baudrates were selected; choose the lowest speed
		conf = stm32.SPI_PCLK_256
	}

	return conf << stm32.SPI_CR1_BR_Pos
}

// Configure SPI pins for input output and clock
func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}
