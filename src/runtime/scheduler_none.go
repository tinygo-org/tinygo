// +build scheduler.none

package runtime

//go:linkname sleep time.Sleep
func sleep(duration int64) {
	sleepTicks(timeUnit(duration / tickMicros))
}

// getSystemStackPointer returns the current stack pointer of the system stack.
// This is always the current stack pointer.
func getSystemStackPointer() uintptr {
	return getCurrentStackPointer()
}
