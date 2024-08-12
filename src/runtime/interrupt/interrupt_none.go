//go:build !baremetal

package interrupt

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
	return 0
}

// Restore restores interrupts to what they were before. Give the previous state
// returned by Disable as a parameter. If interrupts were disabled before
// calling Disable, this will not re-enable interrupts, allowing for nested
// critical sections.
func Restore(state State) {}

// In returns whether the system is currently in an interrupt.
func In() bool {
	// There are no interrupts, so it can't be in one.
	return false
}
