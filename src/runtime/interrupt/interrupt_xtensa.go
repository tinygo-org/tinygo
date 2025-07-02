//go:build xtensa

package interrupt

import "device"

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
	return State(device.AsmFull("rsil {}, 15", nil))
}

// Restore restores interrupts to what they were before. Give the previous state
// returned by Disable as a parameter. If interrupts were disabled before
// calling Disable, this will not re-enable interrupts, allowing for nested
// critical sections.
func Restore(state State) {
	device.AsmFull("wsr {state}, PS", map[string]interface{}{
		"state": state,
	})
}

// In returns whether the system is currently in an interrupt.
//
// Warning: interrupts have not been implemented for Xtensa yet so this always
// returns false.
func In() bool {
	return false
}
