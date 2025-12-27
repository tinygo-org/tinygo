//go:build tinygo.riscv

package runtime

import "device/riscv"

const deferExtraRegs = 0

const callInstSize = 4 // 8 without relaxation, maybe 4 with relaxation

// maxAlign is the maximum alignment required from the memory allocator.
// The RISC-V ABI requires 16-byte alignment for the stack and 128-bit floats.
// Source: https://riscv.org/wp-content/uploads/2015/01/riscv-calling.pdf
const maxAlign = 16

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
	if runqueue := schedulerRunQueue(); runqueue == nil || !runqueue.Empty() {
		riscv.Asm("wfi")
	}
	riscv.EnableInterrupts(mask)
}
