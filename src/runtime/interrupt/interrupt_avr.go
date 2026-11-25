//go:build avr

package interrupt

import "device/avr"

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
	return State(avr.DisableInterrupts())
}

// Restore restores interrupts to what they were before. Give the previous state
// returned by Disable as a parameter. If interrupts were disabled before
// calling Disable, this will not re-enable interrupts, allowing for nested
// critical sections.
func Restore(state State) {
	avr.EnableInterrupts(uint8(state))
}

// In returns whether the system is currently in an interrupt.
//
// Warning: this always returns false on AVR, as there does not appear to be a
// reliable way to determine whether we're currently running inside an interrupt
// handler.
func In() bool {
	return false
}
