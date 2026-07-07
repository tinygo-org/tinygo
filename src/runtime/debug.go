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

// Stub for SetBlockProfileRate, does not do anything. TinyGo has no block
// profiler.
func SetBlockProfileRate(rate int) {
}

// Stub for SetMutexProfileFraction, does not do anything and always reports the
// previous rate as 0. TinyGo has no mutex profiler.
func SetMutexProfileFraction(rate int) int {
	return 0
}
