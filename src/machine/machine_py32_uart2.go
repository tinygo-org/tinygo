//go:build py32f003xx || py32f030xx || py32f040xx

package machine

import (
	"device/py32"
	"runtime/interrupt"
)

// UART2 is the second USART (USART2), available on PY32 parts that provide it.
// It shares the generic UART driver; its clock gate and interrupt are wired up
// by setupUSART2.
var UART2 = &UART{Bus: py32.USART2, Buffer: NewRingBuffer(), setup: setupUSART2}

// setupUSART2 enables the USART2 peripheral clock and installs its RX interrupt
// handler.
func setupUSART2(uart *UART) {
	usart2RX = uart
	py32.RCC.APBENR1.SetBits(py32.RCC_APBENR1_USART2EN)
	uart.irq = interrupt.New(py32.IRQ_USART2, handleUSART2Interrupt)
	uart.irq.SetPriority(0xc0)
	uart.irq.Enable()
}

func handleUSART2Interrupt(interrupt.Interrupt) {
	usart2RX.Receive(uint8(usart2RX.Bus.DR.Get()))
}

// usart2RX is the UART serviced by handleUSART2Interrupt, set by setupUSART2.
var usart2RX *UART
