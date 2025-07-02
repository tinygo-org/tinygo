package runtime

// Stub for NumCgoCall, does not return the real value
func NumCgoCall() int {
	return 0
}

// Stub for NumGoroutine, does not return the real value
func NumGoroutine() int {
	return 1
}

// Stub for Breakpoint, does not do anything.
func Breakpoint() {
	panic("Breakpoint not supported")
}
