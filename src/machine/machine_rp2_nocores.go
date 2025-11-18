//go:build (rp2040 || rp2350) && !scheduler.cores

package machine

// LockCore is not available without the cores scheduler.
// This is a stub that panics.
func LockCore(core int) {
	panic("machine.LockCore: not available without scheduler.cores")
}

// UnlockCore is not available without the cores scheduler.
// This is a stub that panics.
func UnlockCore() {
	panic("machine.UnlockCore: not available without scheduler.cores")
}
