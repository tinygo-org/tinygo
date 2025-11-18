//go:build (rp2040 || rp2350) && scheduler.cores

package machine

const numCPU = 2 // RP2040 and RP2350 both have 2 cores

// LockCore implementation for the cores scheduler.
func LockCore(core int) {
	if core < 0 || core >= numCPU {
		panic("machine: core out of range")
	}
	machineLockCore(core)
}

// UnlockCore implementation for the cores scheduler.
func UnlockCore() {
	machineUnlockCore()
}

// Internal functions implemented in runtime/scheduler_cores.go
//
//go:linkname machineLockCore runtime.machineLockCore
func machineLockCore(core int)

//go:linkname machineUnlockCore runtime.machineUnlockCore
func machineUnlockCore()
