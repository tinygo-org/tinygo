//go:build scheduler.cores && !rp2040 && !rp2350

package runtime

import "sync/atomic"

var (
	// Set to 1 by interruptSleepTicksMulticore to end a sleep in sleepTicks.
	sleepWake [numCPU]atomic.Uint32

	// Set to 1 while a core waits or sleeps with no goroutine to run.
	coreIdle [numCPU]atomic.Uint32

	// Set to 1 by schedulerEndWait to end a wait in schedulerUnlockAndWait.
	schedulerWakePending [numCPU]atomic.Uint32
)

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	core := currentCPU()
	waitingCores |= 1 << core
	if !hasExternalWakeups && waitingCores == (1<<numCPU)-1 {
		runtimeFatal("all cores are sleeping - deadlock!")
	}

	// Interrupts stay masked from the unlock until the lock is taken again,
	// except where coreWait unmasks them.
	state := schedulerLock.unlockMasked()
	coreIdle[core].Store(1)
	for schedulerWakePending[core].Swap(0) == 0 {
		coreWait()
	}
	coreIdle[core].Store(0)
	schedulerLock.lockMasked(state)
}

// Almost identical to sleepTicks, except that it will unlock/lock the scheduler
// while sleeping and is interruptible by interruptSleepTicksMulticore.
// This may only be called with the scheduler lock held.
func sleepTicksMulticore(d timeUnit) {
	core := currentCPU()
	sleepingCore = uint8(core)
	sleepWake[core].Store(0)

	schedulerLock.Unlock()
	coreIdle[core].Store(1)
	sleepTicks(d)
	coreIdle[core].Store(0)
	schedulerLock.Lock()

	sleepingCore = 0xff
}

// Interrupt an ongoing call to sleepTicksMulticore on another core.
// Must be called with the scheduler lock held.
func interruptSleepTicksMulticore(wakeup timeUnit) {
	if hasSleepingCore() {
		sleepWake[sleepingCore].Store(1)
	}
}

func sleepWoken(core uint32) bool {
	return sleepWake[core].Load() != 0
}
