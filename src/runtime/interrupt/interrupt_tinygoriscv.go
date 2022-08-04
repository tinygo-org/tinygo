//go:build tinygo.riscv
// +build tinygo.riscv

package interrupt

import "tinygo.org/x/device/riscv"

// State represents the previous global interrupt state.
type State uintptr

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
	return State(riscv.DisableInterrupts())
}

// Restore restores interrupts to what they were before. Give the previous state
// returned by Disable as a parameter. If interrupts were disabled before
// calling Disable, this will not re-enable interrupts, allowing for nested
// cricital sections.
func Restore(state State) {
	riscv.EnableInterrupts(uintptr(state))
}
