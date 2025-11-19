//go:build (rp2040 || rp2350) && scheduler.cores

package machine

const numCPU = 2 // RP2040 and RP2350 both have 2 cores

// LockCore sets the affinity for the current goroutine to the specified core.
// This does not immediately migrate the goroutine; migration occurs at the next
// scheduling point. See machine_rp2.go for full documentation.
// Important: LockCore sets the affinity but does not immediately migrate the
// goroutine to the target core. The actual migration happens at the next
// scheduling point (e.g., channel operation, time.Sleep, or Gosched). After
// that point, the goroutine will wait in the target core's queue if that core
// is busy running another goroutine.
//
// To avoid potential blocking on a busy core, consider calling LockCore in an
// init function before any other goroutines have started. This guarantees the
// target core is available.
//
// This is useful for:
//   - Isolating time-critical operations to a dedicated core
//   - Improving cache locality for performance-sensitive code
//   - Exclusive access to core-local resources
//
// Warning: Pinning goroutines can lead to load imbalance. The goroutine will
// wait in the specified core's queue even if other cores are idle. If a
// long-running goroutine occupies the target core, LockCore may appear to
// block indefinitely (until the next scheduling point on the target core).
//
// Valid core values are 0 and 1. Panics if core is out of range.
//
// Only available on RP2040 and RP2350 with the "cores" scheduler.
func LockCore(core int) {
	if core < 0 || core >= numCPU {
		panic("machine: core out of range")
	}
	machineLockCore(core)
}

// UnlockCore unpins the calling goroutine, allowing it to run on any available core.
// This undoes a previous call to LockCore.
//
// After calling UnlockCore, the scheduler is free to schedule the goroutine on
// any core for automatic load balancing.
//
// Only available on RP2040 and RP2350 with the "cores" scheduler.
func UnlockCore() {
	machineUnlockCore()
}

// Internal functions implemented in runtime/scheduler_cores.go
//
//go:linkname machineLockCore runtime.machineLockCore
func machineLockCore(core int)

//go:linkname machineUnlockCore runtime.machineUnlockCore
func machineUnlockCore()
