//go:build !scheduler.cores

package machine

// AssignCore assigns the calling goroutine to core until it exits. Without the
// cores scheduler only core 0 is valid.
//
// Experimental: this API may change or be removed in a future release.
func AssignCore(core int) {
	if core != 0 {
		panic("machine: core out of range")
	}
}
