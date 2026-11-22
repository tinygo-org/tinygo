//go:build (rp2040 || rp2350) && scheduler.cores

package runtime

import (
	"device/arm"
	"math/bits"
	"runtime/volatile"
)

// Bitset of cores waiting in schedulerUnlockAndWait.
// Must only be accessed with the scheduler lock held.
var waitingCores uint8

// Set to 1 by schedulerWakeCore to end a wait in schedulerUnlockAndWait.
var schedulerWakePending [2]uint32

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	core := currentCPU()
	waitingCores |= 1 << core
	schedulerWakePending[core] = 0
	schedulerLock.Unlock()

	// The sev in Unlock also sets the local event register, so one wfe returns at once.
	// See the Arm Cortex-M33 Devices Generic User Guide (100235) section 4.13.11.
	for volatile.LoadUint32(&schedulerWakePending[core]) == 0 {
		arm.Asm("wfe")
	}
	schedulerLock.Lock()
}

// Wake another core, if one is waiting or sleeping. Must be called with the
// scheduler lock held.
func schedulerWake() {
	if waitingCores != 0 {
		schedulerWakeCore(uint32(bits.TrailingZeros8(waitingCores)))
	} else if hasSleepingCore() {
		interruptSleepTicksMulticore(0)
	}
}

// Wake the given core if it is waiting or sleeping. Must be called with the
// scheduler lock held.
func schedulerWakeCore(core uint32) {
	if waitingCores&(1<<core) != 0 {
		waitingCores &^= 1 << core
		volatile.StoreUint32(&schedulerWakePending[core], 1)
		arm.Asm("sev")
	} else if uint32(sleepingCore) == core {
		interruptSleepTicksMulticore(0)
	}
}
