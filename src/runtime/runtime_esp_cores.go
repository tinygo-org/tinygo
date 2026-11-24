//go:build (esp32 || esp32s3 || esp32p4) && scheduler.cores

package runtime

import "sync/atomic"

var (
	// Set to 1 by interruptSleepTicksMulticore to end a sleep in sleepTicks.
	sleepWake [numCPU]atomic.Uint32

	// Set to 1 while a core waits or sleeps with no goroutine to run.
	coreIdle [numCPU]atomic.Uint32
)

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
