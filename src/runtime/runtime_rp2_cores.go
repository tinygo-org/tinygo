//go:build (rp2040 || rp2350) && scheduler.cores

package runtime

import (
	"device/arm"
	"runtime/volatile"
)

// Set to 1 by schedulerEndWait to end a wait in schedulerUnlockAndWait.
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

// End schedulerUnlockAndWait on the given core.
func schedulerEndWait(core uint32) {
	volatile.StoreUint32(&schedulerWakePending[core], 1)
	arm.Asm("sev")
}
