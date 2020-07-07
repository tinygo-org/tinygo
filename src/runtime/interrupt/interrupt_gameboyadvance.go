// +build gameboyadvance

package interrupt

import (
	"runtime/volatile"
	"unsafe"
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

// callInterruptHandler is a compiler-generated function that calls the
// appropriate interrupt handler for the given interrupt ID.
//go:linkname callInterruptHandler runtime.callInterruptHandler
func callInterruptHandler(id int)

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
