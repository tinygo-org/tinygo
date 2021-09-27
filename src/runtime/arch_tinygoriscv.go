// +build tinygo.riscv

package runtime

import "device/riscv"

func getCurrentStackPointer() uintptr {
	return riscv.AsmFull("mv {}, sp", nil)
}

// Documentation:
// * https://llvm.org/docs/Atomics.html
// * https://gcc.gnu.org/onlinedocs/gcc/_005f_005fsync-Builtins.html
//
// In the case of RISC-V, some operations may be implemented with libcalls if
// the operation is too big to be handled by assembly. Officially, these calls
// should be implemented with a lock-free algorithm but as (as of this time) all
// supported RISC-V chips have a single hart, we can simply disable interrupts
// to get the same behavior.

//export __atomic_load_4
func __atomic_load_4(ptr *uint32, ordering int32) uint32 {
	mask := riscv.DisableInterrupts()
	value := *ptr
	riscv.EnableInterrupts(mask)
	return value
}

//export __atomic_store_4
func __atomic_store_4(ptr *uint32, value uint32, ordering int32) {
	mask := riscv.DisableInterrupts()
	*ptr = value
	riscv.EnableInterrupts(mask)
}

//export __atomic_exchange_4
func __atomic_exchange_4(ptr *uint32, value uint32, ordering int32) uint32 {
	mask := riscv.DisableInterrupts()
	oldValue := *ptr
	*ptr = value
	riscv.EnableInterrupts(mask)
	return oldValue
}

//export __atomic_compare_exchange_4
func __atomic_compare_exchange_4(ptr, expected *uint32, desired uint32, success_ordering, failure_ordering int32) bool {
	mask := riscv.DisableInterrupts()
	oldValue := *ptr
	success := oldValue == *expected
	if success {
		*ptr = desired
	} else {
		*expected = oldValue
	}
	riscv.EnableInterrupts(mask)
	return success
}

//export __atomic_fetch_add_4
func __atomic_fetch_add_4(ptr *uint32, value uint32, ordering int32) uint32 {
	mask := riscv.DisableInterrupts()
	oldValue := *ptr
	*ptr = oldValue + value
	riscv.EnableInterrupts(mask)
	return oldValue
}

//export __atomic_load_8
func __atomic_load_8(ptr *uint64, ordering int32) uint64 {
	mask := riscv.DisableInterrupts()
	value := *ptr
	riscv.EnableInterrupts(mask)
	return value
}

//export __atomic_store_8
func __atomic_store_8(ptr *uint64, value uint64, ordering int32) {
	mask := riscv.DisableInterrupts()
	*ptr = value
	riscv.EnableInterrupts(mask)
}

//export __atomic_exchange_8
func __atomic_exchange_8(ptr *uint64, value uint64, ordering int32) uint64 {
	mask := riscv.DisableInterrupts()
	oldValue := *ptr
	*ptr = value
	riscv.EnableInterrupts(mask)
	return oldValue
}

//export __atomic_compare_exchange_8
func __atomic_compare_exchange_8(ptr, expected *uint64, desired uint64, success_ordering, failure_ordering int32) bool {
	mask := riscv.DisableInterrupts()
	oldValue := *ptr
	success := oldValue == *expected
	if success {
		*ptr = desired
	} else {
		*expected = oldValue
	}
	riscv.EnableInterrupts(mask)
	return success
}

//export __atomic_fetch_add_8
func __atomic_fetch_add_8(ptr *uint64, value uint64, ordering int32) uint64 {
	mask := riscv.DisableInterrupts()
	oldValue := *ptr
	*ptr = oldValue + value
	riscv.EnableInterrupts(mask)
	return oldValue
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
