// +build stm32f407

package machine

// Peripheral abstraction layer for the stm32f407

import (
	"device/stm32"
)

func CPUFrequency() uint32 {
	return 168000000
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

//---------- UART related code

// Configure the UART.
func (uart *UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.AltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.AltFuncSelector)
}

// UART baudrate calc based on the bus and clockspeed
// NOTE: keep this in sync with the runtime/runtime_stm32f407.go clock init code
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

// Register names vary by ST processor, these are for STM F407
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.DR
	uart.txReg = &uart.Bus.DR
	uart.statusReg = &uart.Bus.SR
	uart.txEmptyFlag = stm32.USART_SR_TXE
}

//---------- SPI related types and code

// SPI on the STM32Fxxx using MODER / alternate function pins
type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
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

	// These are based on APB2 clock frquency (84MHz on the discovery board)
	// TODO: also include the MCU/APB clock setting in the equation
	switch true {
	case localFrequency < 328125:
		conf = stm32.SPI_CR1_BR_Div256
	case localFrequency < 656250:
		conf = stm32.SPI_CR1_BR_Div128
	case localFrequency < 1312500:
		conf = stm32.SPI_CR1_BR_Div64
	case localFrequency < 2625000:
		conf = stm32.SPI_CR1_BR_Div32
	case localFrequency < 5250000:
		conf = stm32.SPI_CR1_BR_Div16
	case localFrequency < 10500000:
		conf = stm32.SPI_CR1_BR_Div8
		// NOTE: many SPI components won't operate reliably (or at all) above 10MHz
		// Check the datasheet of the part
	case localFrequency < 21000000:
		conf = stm32.SPI_CR1_BR_Div4
	case localFrequency < 42000000:
		conf = stm32.SPI_CR1_BR_Div2
	default:
		// None of the specific baudrates were selected; choose the lowest speed
		conf = stm32.SPI_CR1_BR_Div256
	}

	return conf << stm32.SPI_CR1_BR_Pos
}

// Configure SPI pins for input output and clock
func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}
