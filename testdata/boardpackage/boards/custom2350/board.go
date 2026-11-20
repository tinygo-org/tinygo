package custom2350

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
	uart0 = machine.UART{
		Buffer: machine.NewRingBuffer(),
		Bus:    rp.UART0,
	}
	UART0 = &uart0
)

//go:linkname boardXTAL machine.boardXTAL
func boardXTAL() uint32 {
	return 12_000_000
}

//go:linkname boardUSBIdentity machine.boardUSBIdentity
func boardUSBIdentity() (uint16, uint16, string, string) {
	return 0x1209, 0x4243, "CustomHW", "Custom2350"
}

//go:linkname boardInit0 runtime.boardInit0
func boardInit0() {
	rp.SIO.GPIO_OE_SET.Set(1 << uint32(LED))
}

//go:linkname boardInit1 runtime.boardInit1
func boardInit1() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, uart0.HandleInterrupt)
	_ = UART0.Configure(machine.UARTConfig{
		TX: UART0TX,
		RX: UART0RX,
	})
	setBoardSerial(UART0)
}
