// +build gameboyadvance

package interrupt

import (
	"runtime/volatile"
	"unsafe"
)

var handlers = [14]func(Interrupt){}

const (
	IRQ_VBLANK  = 0
	IRQ_HBLANK  = 1
	IRQ_VCOUNT  = 2
	IRQ_TIMER0  = 3
	IRQ_TIMER1  = 4
	IRQ_TIMER2  = 5
	IRQ_TIMER3  = 6
	IRQ_COM     = 7
	IRQ_DMA0    = 8
	IRQ_DMA1    = 9
	IRQ_DMA2    = 10
	IRQ_DMA3    = 11
	IRQ_KEYPAD  = 12
	IRQ_GAMEPAK = 13
)

var (
	regInterruptEnable = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000200)))
	regInterruptRequestFlags = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000202)))
	regInterruptMasterEnable = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000208)))
)

// New creates a new Interrupt object. Do not call it multiple times. If you do,
// make sure the interrupt is disabled while you do so. The last call will set
// the active interrupt handler.
func New(id int, handler func(Interrupt)) Interrupt {
	handlers[id] = handler
	return Interrupt{id}
}

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
	regInterruptEnable.SetBits(1 << irq.num)
}

//export handleInterrupt
func handleInterrupt() {
	flags := regInterruptRequestFlags.Get()
	for i := range handlers {
		if flags & (1 << i) != 0 {
			irq := Interrupt{i}
			regInterruptRequestFlags.Set(1 << i) // acknowledge interrupt
			handlers[i](irq)
		}
	}
}
