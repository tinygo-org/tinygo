// +build stm32,stm32f407

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/stm32"
	"unsafe"
)

const CPU_FREQUENCY = 168000000

const (
	// Alternative peripheral pin functions
	AF0_SYSTEM                AltFunc = 0
	AF1_TIM1_2                        = 1
	AF2_TIM3_4_5                      = 2
	AF3_TIM8_9_10_11                  = 3
	AF4_I2C1_2_3                      = 4
	AF5_SPI1_SPI2                     = 5
	AF6_SPI3                          = 6
	AF7_USART1_2_3                    = 7
	AF8_USART4_5_6                    = 8
	AF9_CAN1_CAN2_TIM12_13_14         = 9
	AF10_OTG_FS_OTG_HS                = 10
	AF11_ETH                          = 11
	AF12_FSMC_SDIO_OTG_HS_1           = 12
	AF13_DCMI                         = 13
	AF14                              = 14
	AF15_EVENTOUT                     = 15
)

// SPI clock frequency dividers. Note that the actual clock frequency
// depends on which SPI is being used, as SPI2 and 3 are attached to
// a different clock source
// By default at startup on the 407 discovery board SPI_PCLK_2 for SPI1
// will equate to 8 Mhz, though this may change based on the stm chip in use
const (
	SPI_PCLK_2 = iota
	SPI_PCLK_4
	SPI_PCLK_8
	SPI_PCLK_16
	SPI_PCLK_32
	SPI_PCLK_64
	SPI_PCLK_128
	SPI_PCLK_256
)

// Peripheral clock frequencies as set in runtime_stm32f407.go
const (
	APB1_PCLK1_FREQ = 42000000
	APB2_PCLK2_FREQ = 84000000
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
	if bus == unsafe.Pointer(stm32.USART1) {
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	} else if bus == unsafe.Pointer(stm32.USART2) {
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)
	} else if bus == unsafe.Pointer(stm32.I2C1) {
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)
	} else if bus == unsafe.Pointer(stm32.SPI1) {
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	}
}

//-------- UART hooks

var (
	// Both UART0 and UART1 refer to USART2.
	UART0 = UART{
		Buffer: NewRingBuffer(),
		Bus:    stm32.USART2,
		IRQVal: stm32.IRQ_USART2,
	}
	UART1 = &UART0
)

// UART pin config
func (uart UART) configurePins(config UARTConfig) {
	config.TX.Configure(PinConfig{Mode: PinModeUartTX})
	config.RX.Configure(PinConfig{Mode: PinModeUartRX})
}

// UART baudrate calc
func (uart UART) getBaudRateDivisor(br uint32) uint32 {
	// TODO: derive this from the CPU/system clocks
	//  appropriate for the USART in use
	var clock uint32
	switch uart.Bus {
	case unsafe.Pointer(stm32.USART1):
		clock = APB1_PCLK1_FREQ
	case unsafe.Pointer(stm32.USART2):
		fallthrough
	case unsafe.Pointer(stm32.USART3):
		clock = APB2_PCLK2_FREQ
	}
	return clock / br
}

//go:export USART2_IRQHandler
func handleUSART2() {
	UART1.Receive(byte((UART1.Bus.DR.Get() & 0xFF)))
}

// Configure the SDA and SCL pins for I2C
func (i2c I2C) configurePins(config I2CConfig) {
	// enable clock for GPIO pins used for I2C
	if config.SCL == 0 {
		config.SCL = I2C0_SCL
	}
	if config.SDA == 0 {
		config.SDA = I2C0_SDA
	}

	config.SCL.Configure(PinConfig{Mode: PinModeI2CScl})
	config.SDA.Configure(PinConfig{Mode: PinModeI2CSda})
}

// Set baud rate for SPI
func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	// set frequency dependent on PCLK prescaler
	// TODO: also include the MCU/APB clock setting in the equation
	// These are based on APB2 clock frquency
	switch config.Frequency {
	case 328125:
		conf = SPI_PCLK_256
	case 656250:
		conf = SPI_PCLK_128
	case 1312500:
		conf = SPI_PCLK_64
	case 2625000:
		conf = SPI_PCLK_32
	case 5250000:
		conf = SPI_PCLK_16
	case 10500000:
		conf = SPI_PCLK_8
	case 21000000:
		conf = SPI_PCLK_4
	case 42000000:
		conf = SPI_PCLK_2
	default:
		conf = SPI_PCLK_256
	}
	return conf << stm32.SPI_CR1_BR_Pos
}

// Configure SPI pins for input output and clock
func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.Configure(PinConfig{Mode: PinModeSpiClk})
	config.MOSI.Configure(PinConfig{Mode: PinModeSpiMosi})
	config.MISO.Configure(PinConfig{Mode: PinModeSpiMiso})
}
