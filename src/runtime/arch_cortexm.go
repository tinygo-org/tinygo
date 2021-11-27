// +build cortexm

package runtime

import (
	"device/arm"
)

const GOARCH = "arm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

func getCurrentStackPointer() uintptr {
	return uintptr(stacksave())
}

// Documentation:
// * https://llvm.org/docs/Atomics.html
// * https://gcc.gnu.org/onlinedocs/gcc/_005f_005fsync-Builtins.html
//
// In the case of Cortex-M, some atomic operations are emitted inline while
// others are emitted as libcalls. How many are emitted as libcalls depends on
// the MCU core variant (M3 and higher support some 32-bit atomic operations
// while M0 and M0+ do not).

//export __sync_fetch_and_add_4
func __sync_fetch_and_add_4(ptr *uint32, value uint32) uint32 {
	mask := arm.DisableInterrupts()
	oldValue := *ptr
	*ptr = oldValue + value
	arm.EnableInterrupts(mask)
	return oldValue
}

//export __sync_fetch_and_add_8
func __sync_fetch_and_add_8(ptr *uint64, value uint64) uint64 {
	mask := arm.DisableInterrupts()
	oldValue := *ptr
	*ptr = oldValue + value
	arm.EnableInterrupts(mask)
	return oldValue
}

//export __sync_lock_test_and_set_4
func __sync_lock_test_and_set_4(ptr *uint32, value uint32) uint32 {
	mask := arm.DisableInterrupts()
	oldValue := *ptr
	*ptr = value
	arm.EnableInterrupts(mask)
	return oldValue
}

//export __sync_lock_test_and_set_8
func __sync_lock_test_and_set_8(ptr *uint64, value uint64) uint64 {
	mask := arm.DisableInterrupts()
	oldValue := *ptr
	*ptr = value
	arm.EnableInterrupts(mask)
	return oldValue
}

//export __sync_val_compare_and_swap_4
func __sync_val_compare_and_swap_4(ptr *uint32, expected, desired uint32) uint32 {
	mask := arm.DisableInterrupts()
	oldValue := *ptr
	if oldValue == expected {
		*ptr = desired
	}
	arm.EnableInterrupts(mask)
	return oldValue
}

//export __sync_val_compare_and_swap_8
func __sync_val_compare_and_swap_8(ptr *uint64, expected, desired uint64) uint64 {
	mask := arm.DisableInterrupts()
	oldValue := *ptr
	if oldValue == expected {
		*ptr = desired
	}
	arm.EnableInterrupts(mask)
	return oldValue
}

// The safest thing to do here would just be to disable interrupts for
// procPin/procUnpin. Note that a global variable is safe in this case, as any
// access to procPinnedMask will happen with interrupts disabled.

var procPinnedMask uintptr

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
	procPinnedMask = arm.DisableInterrupts()
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
	arm.EnableInterrupts(procPinnedMask)
}
