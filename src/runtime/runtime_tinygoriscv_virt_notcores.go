//go:build tinygo.riscv && virt && !scheduler.cores

package runtime

import (
	"device/riscv"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

// ACLINT MTIMER compare register of each hart. The layout is the same as the
// SiFive CLINT. See https://github.com/riscvarchive/riscv-aclint/blob/main/riscv-aclint.adoc
var aclintMTIMECMP = (*[4095]volatile.Register64)(unsafe.Pointer(uintptr(0x0200_4000)))

// Wait with wfi for the machine timer, as fe310 does. wfi wakes on a pending
// MTIP even with MSTATUS.MIE clear, so no handler runs and no wakeup is lost.
func sleepTicks(d timeUnit) {
	hart := currentCPU()
	target := uint64(ticks() + d)
	state := interrupt.Disable()
	aclintMTIMECMP[hart].Set(target)
	riscv.MIE.SetBits(riscv.MIE_MTIE)
	for uint64(ticks()) < target {
		riscv.Asm("wfi")
	}
	riscv.MIE.ClearBits(riscv.MIE_MTIE)
	aclintMTIMECMP[hart].Set(0xffff_ffff_ffff_ffff)
	interrupt.Restore(state)
}
