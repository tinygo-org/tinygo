//go:build tinygo.riscv && virt && scheduler.cores

package runtime

import "device/riscv"

func currentCPU() uint32 {
	return uint32(riscv.MHARTID.Get())
}

// Spin until the time is up, or until another hart ends the sleep. esp32p4
// sleeps the same way.
func sleepTicks(d timeUnit) {
	core := currentCPU()
	target := ticks() + d
	for ticks() < target && !sleepWoken(core) {
	}
}

// QEMU virt has no external interrupt that can wake a hart, so a wait on
// every hart at once is a deadlock.
const hasExternalWakeups = false

// wfi wakes on a pending interrupt even with MSTATUS.MIE clear, so a wake sent
// after the unlock is not lost. MIE goes on briefly so a GC pause request can run.
func coreWait() {
	riscv.Asm("wfi")
	riscv.MSTATUS.SetBits(riscv.MSTATUS_MIE)
	riscv.MSTATUS.ClearBits(riscv.MSTATUS_MIE)
}

// End schedulerUnlockAndWait on the given hart.
func schedulerEndWait(hart uint32) {
	schedulerWakePending[hart].Store(1)
	signalHart(hart)
}
