//go:build scheduler.none

package runtime

//go:linkname sleep time.Sleep
func sleep(duration int64) {
	if duration <= 0 {
		return
	}

	sleepTicks(nanosecondsToTicks(duration))
}

// getSystemStackPointer returns the current stack pointer of the system stack.
// This is always the current stack pointer.
func getSystemStackPointer() uintptr {
	return getCurrentStackPointer()
}

// run is called by the program entry point to execute the go program.
// With the "none" scheduler, init and the main function are invoked directly.
func run() {
	if !runtimeInitialized {
		initHeap()
		initAll()
		runtimeInitialized = true
	}
	callMain()
}

const hasScheduler = false
