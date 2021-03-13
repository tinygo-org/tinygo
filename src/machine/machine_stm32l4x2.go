// +build stm32l4x2

package machine

// Peripheral abstraction layer for the stm32l4x2

import (
	"device/stm32"
)

func CPUFrequency() uint32 {
	return 80000000
}

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
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXE
}
