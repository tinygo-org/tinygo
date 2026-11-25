//go:build tinygo.riscv && virt && scheduler.cores

package runtime

import (
	"device/riscv"
	"sync/atomic"
)

// Spin until the time is up, or until another hart ends the sleep. esp32p4
// sleeps the same way.
func sleepTicks(d timeUnit) {
	core := currentCPU()
	target := ticks() + d
	for ticks() < target && !sleepWoken(core) {
	}
}

// State used to distinguish scheduler wakeups from other software interrupts.
var schedulerWakePending [numCPU]atomic.Uint32

// Put the scheduler to sleep, since there are no tasks to run.
// This will unlock the scheduler lock, and must be called with the scheduler
// lock held.
func schedulerUnlockAndWait() {
	hartID := currentCPU()

	// Mark the current hart as sleeping.
	waitingCores |= 1 << hartID

	// If this is the last core awake and is going to sleep, the scheduler is
	// deadlocked.
	// We can do this check since this is not baremetal: there won't be any
	// external interrupts that might unblock a goroutine.
	if waitingCores == (1<<numCPU)-1 {
		runtimeFatal("all cores are sleeping - deadlock!")
	}

	// wfi wakes on a pending interrupt even with MSTATUS.MIE clear, so a wake
	// sent after the unlock is not lost. Interrupts go on for a moment after
	// each wake, so a GC pause request can run.
	state := schedulerLock.unlockMasked()
	for schedulerWakePending[hartID].Swap(0) == 0 {
		riscv.Asm("wfi")
		riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
		riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)
	}
	schedulerLock.lockMasked(state)
}

// End schedulerUnlockAndWait on the given hart.
func schedulerEndWait(hart uint32) {
	schedulerWakePending[hart].Store(1)
	signalHart(hart)
}
