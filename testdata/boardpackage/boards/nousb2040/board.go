package nousb2040

import (
	"device/rp"
	"machine"
	"runtime/interrupt"
	_ "unsafe"
)

const (
	LED     = machine.GPIO25
	UART0TX = machine.GPIO0
	UART0RX = machine.GPIO1
)

var (
	uart0Buf machine.RingBuffer
	uart0    machine.UART
	UART0    *machine.UART
)

//go:linkname boardXTAL machine.boardXTAL
func boardXTAL() uint32 {
	return 12_000_000
}

//go:linkname boardInit0 runtime.boardInit0
func boardInit0() {
	rp.SIO.GPIO_OE_SET.Set(1 << uint32(LED))
}

//go:linkname boardInit1 runtime.boardInit1
func boardInit1() {
	uart0 = machine.UART{
		Buffer:    &uart0Buf,
		Bus:       rp.UART0,
		Interrupt: interrupt.New(rp.IRQ_UART0_IRQ, uart0.HandleInterrupt),
	}
	UART0 = &uart0
	_ = UART0.Configure(machine.UARTConfig{
		TX: UART0TX,
		RX: UART0RX,
	})
	setBoardSerial(UART0)
}
