//go:build scheduler.cores

package machine

// AssignCore assigns the calling goroutine to core until it exits, and returns
// once it runs there. It panics if the goroutine is already assigned.
// Like any blocking call, it must not be called with interrupts disabled.
//
// Experimental: this API may change or be removed in a future release.
func AssignCore(core int) {
	assignCore(core)
}

func assignCore(core int)
