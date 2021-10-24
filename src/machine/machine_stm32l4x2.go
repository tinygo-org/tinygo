// +build stm32l4x2

package machine

// Peripheral abstraction layer for the stm32l4x2

import (
	"device/stm32"
)

func CPUFrequency() uint32 {
	return 80000000
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 80e6 // 80MHz
const APB2_TIM_FREQ = 80e6 // 80MHz

//---------- UART related code

// Configure the UART.
func (uart *UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

// UART baudrate calc based on the bus and clockspeed
// NOTE: keep this in sync with the runtime/runtime_stm32l5x2.go clock init code
func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	return (CPUFrequency() / baudRate)
}

// Register names vary by ST processor, these are for STM L5
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR.Register32
	uart.txReg = &uart.Bus.TDR.Register32
	uart.statusReg = &uart.Bus.ISR.Register32
	uart.txEmptyFlag = stm32.USART_ISR_TXE
}

//---------- I2C related code

// Gets the value for TIMINGR register
func (i2c *I2C) getFreqRange() uint32 {
	// This is a 'magic' value calculated by STM32CubeMX
	// for 80MHz PCLK1.
	// TODO: Do calculations based on PCLK1
	return 0x10909CEC
}
