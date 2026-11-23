//go:build scheduler.cores && !virt

package runtime

import "math/bits"

// Bitset of cores waiting in schedulerUnlockAndWait.
// Must only be accessed with the scheduler lock held.
var waitingCores uint8

// Wake another core, if one is waiting or sleeping. Must be called with the
// scheduler lock held.
func schedulerWake() {
	if core := bits.TrailingZeros8(waitingCores); core < numCPU {
		schedulerWakeCore(uint32(core))
	} else if hasSleepingCore() {
		interruptSleepTicksMulticore(0)
	}
}

// Wake the given core if it is waiting or sleeping. Must be called with the
// scheduler lock held.
func schedulerWakeCore(core uint32) {
	if waitingCores&(1<<core) != 0 {
		waitingCores &^= 1 << core
		schedulerEndWait(core)
	} else if uint32(sleepingCore) == core {
		interruptSleepTicksMulticore(0)
	}
}
