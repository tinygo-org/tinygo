//go:build gameboyadvance
// +build gameboyadvance

package interrupt

import (
	"runtime/volatile"
	"unsafe"
)

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
	regInterruptEnable       = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000200)))
	regInterruptRequestFlags = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000202)))
	regGlobalInterruptEnable = (*volatile.Register16)(unsafe.Pointer(uintptr(0x4000208)))
)

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
	regInterruptEnable.SetBits(1 << uint(irq.num))
}

//export handleInterrupt
func handleInterrupt() {
	flags := regInterruptRequestFlags.Get()
	for i := 0; i < 14; i++ {
		if flags&(1<<uint(i)) != 0 {
			regInterruptRequestFlags.Set(1 << uint(i)) // acknowledge interrupt
			callInterruptHandler(i)
		}
	}
}

// Pseudo function call that is replaced by the compiler with the actual
// functions registered through interrupt.New. If there are none, calls will be
// replaced with 'unreachablecalls will be replaced with 'unreachable'.
//go:linkname callHandlers runtime/interrupt.callHandlers
func callHandlers(num int)

func callInterruptHandler(id int) {
	switch id {
	case IRQ_VBLANK:
		callHandlers(IRQ_VBLANK)
	case IRQ_HBLANK:
		callHandlers(IRQ_HBLANK)
	case IRQ_VCOUNT:
		callHandlers(IRQ_VCOUNT)
	case IRQ_TIMER0:
		callHandlers(IRQ_TIMER0)
	case IRQ_TIMER1:
		callHandlers(IRQ_TIMER1)
	case IRQ_TIMER2:
		callHandlers(IRQ_TIMER2)
	case IRQ_TIMER3:
		callHandlers(IRQ_TIMER3)
	case IRQ_COM:
		callHandlers(IRQ_COM)
	case IRQ_DMA0:
		callHandlers(IRQ_DMA0)
	case IRQ_DMA1:
		callHandlers(IRQ_DMA1)
	case IRQ_DMA2:
		callHandlers(IRQ_DMA2)
	case IRQ_DMA3:
		callHandlers(IRQ_DMA3)
	case IRQ_KEYPAD:
		callHandlers(IRQ_KEYPAD)
	case IRQ_GAMEPAK:
		callHandlers(IRQ_GAMEPAK)
	}
}

// State represents the previous global interrupt state.
type State uint8

// Disable disables all interrupts and returns the previous interrupt state. It
// can be used in a critical section like this:
//
//     state := interrupt.Disable()
//     // critical section
//     interrupt.Restore(state)
//
// Critical sections can be nested. Make sure to call Restore in the same order
// as you called Disable (this happens naturally with the pattern above).
func Disable() (state State) {
	// Save the previous interrupt state.
	state = State(regGlobalInterruptEnable.Get())
	// Disable all interrupts.
	regGlobalInterruptEnable.Set(0)
	return
}

// Restore restores interrupts to what they were before. Give the previous state
// returned by Disable as a parameter. If interrupts were disabled before
// calling Disable, this will not re-enable interrupts, allowing for nested
// cricital sections.
func Restore(state State) {
	// Restore interrupts to the previous state.
	regGlobalInterruptEnable.Set(uint16(state))
}
