//go:build tinygo.riscv
// +build tinygo.riscv

package runtime

import "device/riscv"

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}

// The safest thing to do here would just be to disable interrupts for
// procPin/procUnpin. Note that a global variable is safe in this case, as any
// access to procPinnedMask will happen with interrupts disabled.

var procPinnedMask uintptr

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
	procPinnedMask = riscv.DisableInterrupts()
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
	riscv.EnableInterrupts(procPinnedMask)
}

func waitForEvents() {
	mask := riscv.DisableInterrupts()
	if !runqueue.Empty() {
		riscv.Asm("wfi")
	}
	riscv.EnableInterrupts(mask)
}
