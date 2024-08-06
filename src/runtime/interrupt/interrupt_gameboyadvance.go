//go:build gameboyadvance

package interrupt

// This is good documentation of the GBA: https://www.akkit.org/info/gbatek.htm

import (
	"device/gba"
)

// Enable enables this interrupt. Right after calling this function, the
// interrupt may be invoked if it was already pending.
func (irq Interrupt) Enable() {
	gba.INTERRUPT.IE.SetBits(1 << uint(irq.num))
}

var inInterrupt bool

//export handleInterrupt
func handleInterrupt() {
	inInterrupt = true
	flags := gba.INTERRUPT.IF.Get()
	for i := 0; i < 14; i++ {
		if flags&(1<<uint(i)) != 0 {
			gba.INTERRUPT.IF.Set(1 << uint(i)) // acknowledge interrupt
			callInterruptHandler(i)
		}
	}
	inInterrupt = false
}

// Pseudo function call that is replaced by the compiler with the actual
// functions registered through interrupt.New. If there are none, calls will be
// replaced with 'unreachablecalls will be replaced with 'unreachable'.
//
//go:linkname callHandlers runtime/interrupt.callHandlers
func callHandlers(num int)

func callInterruptHandler(id int) {
	switch id {
	case gba.IRQ_VBLANK:
		callHandlers(gba.IRQ_VBLANK)
	case gba.IRQ_HBLANK:
		callHandlers(gba.IRQ_HBLANK)
	case gba.IRQ_VCOUNT:
		callHandlers(gba.IRQ_VCOUNT)
	case gba.IRQ_TIMER0:
		callHandlers(gba.IRQ_TIMER0)
	case gba.IRQ_TIMER1:
		callHandlers(gba.IRQ_TIMER1)
	case gba.IRQ_TIMER2:
		callHandlers(gba.IRQ_TIMER2)
	case gba.IRQ_TIMER3:
		callHandlers(gba.IRQ_TIMER3)
	case gba.IRQ_COM:
		callHandlers(gba.IRQ_COM)
	case gba.IRQ_DMA0:
		callHandlers(gba.IRQ_DMA0)
	case gba.IRQ_DMA1:
		callHandlers(gba.IRQ_DMA1)
	case gba.IRQ_DMA2:
		callHandlers(gba.IRQ_DMA2)
	case gba.IRQ_DMA3:
		callHandlers(gba.IRQ_DMA3)
	case gba.IRQ_KEYPAD:
		callHandlers(gba.IRQ_KEYPAD)
	case gba.IRQ_GAMEPAK:
		callHandlers(gba.IRQ_GAMEPAK)
	}
}

// State represents the previous global interrupt state.
type State uint8

// Disable disables all interrupts and returns the previous interrupt state. It
// can be used in a critical section like this:
//
//	state := interrupt.Disable()
//	// critical section
//	interrupt.Restore(state)
//
// Critical sections can be nested. Make sure to call Restore in the same order
// as you called Disable (this happens naturally with the pattern above).
func Disable() (state State) {
	// Save the previous interrupt state.
	state = State(gba.INTERRUPT.PAUSE.Get())
	// Disable all interrupts.
	gba.INTERRUPT.PAUSE.Set(0)
	return
}

// Restore restores interrupts to what they were before. Give the previous state
// returned by Disable as a parameter. If interrupts were disabled before
// calling Disable, this will not re-enable interrupts, allowing for nested
// critical sections.
func Restore(state State) {
	// Restore interrupts to the previous state.
	gba.INTERRUPT.PAUSE.Set(uint16(state))
}

// In returns whether the system is currently in an interrupt.
func In() bool {
	return inInterrupt
}
