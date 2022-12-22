//go:build baremetal && !tinygo.wasm

// Automatically generated file. DO NOT EDIT.
// This file implements standins for non-native atomics using critical sections.

package runtime

import (
	"runtime/interrupt"
	_ "unsafe"
)

// Documentation:
// * https://llvm.org/docs/Atomics.html
// * https://gcc.gnu.org/onlinedocs/gcc/_005f_005fsync-Builtins.html
//
// Some atomic operations are emitted inline while others are emitted as libcalls.
// How many are emitted as libcalls depends on the MCU arch and core variant.

// 16-bit atomics.

//export __atomic_load_2
func __atomic_load_2(ptr *uint16, ordering uintptr) uint16 {
	// The LLVM docs for this say that there is a val argument after the pointer.
	// That is a typo, and the GCC docs omit it.
	mask := interrupt.Disable()
	val := *ptr
	interrupt.Restore(mask)
	return val
}

//export __atomic_store_2
func __atomic_store_2(ptr *uint16, val uint16, ordering uintptr) {
	mask := interrupt.Disable()
	*ptr = val
	interrupt.Restore(mask)
}

//go:inline
func doAtomicCAS16(ptr *uint16, expected, desired uint16) uint16 {
	mask := interrupt.Disable()
	old := *ptr
	if old == expected {
		*ptr = desired
	}
	interrupt.Restore(mask)
	return old
}

//export __sync_val_compare_and_swap_2
func __sync_val_compare_and_swap_2(ptr *uint16, expected, desired uint16) uint16 {
	return doAtomicCAS16(ptr, expected, desired)
}

//export __atomic_compare_exchange_2
func __atomic_compare_exchange_2(ptr, expected *uint16, desired uint16, successOrder, failureOrder uintptr) bool {
	exp := *expected
	old := doAtomicCAS16(ptr, exp, desired)
	return old == exp
}

//go:inline
func doAtomicSwap16(ptr *uint16, new uint16) uint16 {
	mask := interrupt.Disable()
	old := *ptr
	*ptr = new
	interrupt.Restore(mask)
	return old
}

//export __sync_lock_test_and_set_2
func __sync_lock_test_and_set_2(ptr *uint16, new uint16) uint16 {
	return doAtomicSwap16(ptr, new)
}

//export __atomic_exchange_2
func __atomic_exchange_2(ptr *uint16, new uint16, ordering uintptr) uint16 {
	return doAtomicSwap16(ptr, new)
}

//go:inline
func doAtomicAdd16(ptr *uint16, value uint16) (old, new uint16) {
	mask := interrupt.Disable()
	old = *ptr
	new = old + value
	*ptr = new
	interrupt.Restore(mask)
	return old, new
}

//export __atomic_fetch_add_2
func __atomic_fetch_add_2(ptr *uint16, value uint16, ordering uintptr) uint16 {
	old, _ := doAtomicAdd16(ptr, value)
	return old
}

//export __sync_fetch_and_add_2
func __sync_fetch_and_add_2(ptr *uint16, value uint16) uint16 {
	old, _ := doAtomicAdd16(ptr, value)
	return old
}

//export __atomic_add_fetch_2
func __atomic_add_fetch_2(ptr *uint16, value uint16, ordering uintptr) uint16 {
	_, new := doAtomicAdd16(ptr, value)
	return new
}

// 32-bit atomics.

//export __atomic_load_4
func __atomic_load_4(ptr *uint32, ordering uintptr) uint32 {
	// The LLVM docs for this say that there is a val argument after the pointer.
	// That is a typo, and the GCC docs omit it.
	mask := interrupt.Disable()
	val := *ptr
	interrupt.Restore(mask)
	return val
}

//export __atomic_store_4
func __atomic_store_4(ptr *uint32, val uint32, ordering uintptr) {
	mask := interrupt.Disable()
	*ptr = val
	interrupt.Restore(mask)
}

//go:inline
func doAtomicCAS32(ptr *uint32, expected, desired uint32) uint32 {
	mask := interrupt.Disable()
	old := *ptr
	if old == expected {
		*ptr = desired
	}
	interrupt.Restore(mask)
	return old
}

//export __sync_val_compare_and_swap_4
func __sync_val_compare_and_swap_4(ptr *uint32, expected, desired uint32) uint32 {
	return doAtomicCAS32(ptr, expected, desired)
}

//export __atomic_compare_exchange_4
func __atomic_compare_exchange_4(ptr, expected *uint32, desired uint32, successOrder, failureOrder uintptr) bool {
	exp := *expected
	old := doAtomicCAS32(ptr, exp, desired)
	return old == exp
}

//go:inline
func doAtomicSwap32(ptr *uint32, new uint32) uint32 {
	mask := interrupt.Disable()
	old := *ptr
	*ptr = new
	interrupt.Restore(mask)
	return old
}

//export __sync_lock_test_and_set_4
func __sync_lock_test_and_set_4(ptr *uint32, new uint32) uint32 {
	return doAtomicSwap32(ptr, new)
}

//export __atomic_exchange_4
func __atomic_exchange_4(ptr *uint32, new uint32, ordering uintptr) uint32 {
	return doAtomicSwap32(ptr, new)
}

//go:inline
func doAtomicAdd32(ptr *uint32, value uint32) (old, new uint32) {
	mask := interrupt.Disable()
	old = *ptr
	new = old + value
	*ptr = new
	interrupt.Restore(mask)
	return old, new
}

//export __atomic_fetch_add_4
func __atomic_fetch_add_4(ptr *uint32, value uint32, ordering uintptr) uint32 {
	old, _ := doAtomicAdd32(ptr, value)
	return old
}

//export __sync_fetch_and_add_4
func __sync_fetch_and_add_4(ptr *uint32, value uint32) uint32 {
	old, _ := doAtomicAdd32(ptr, value)
	return old
}

//export __atomic_add_fetch_4
func __atomic_add_fetch_4(ptr *uint32, value uint32, ordering uintptr) uint32 {
	_, new := doAtomicAdd32(ptr, value)
	return new
}

// 64-bit atomics.

//export __atomic_load_8
func __atomic_load_8(ptr *uint64, ordering uintptr) uint64 {
	// The LLVM docs for this say that there is a val argument after the pointer.
	// That is a typo, and the GCC docs omit it.
	mask := interrupt.Disable()
	val := *ptr
	interrupt.Restore(mask)
	return val
}

//export __atomic_store_8
func __atomic_store_8(ptr *uint64, val uint64, ordering uintptr) {
	mask := interrupt.Disable()
	*ptr = val
	interrupt.Restore(mask)
}

//go:inline
func doAtomicCAS64(ptr *uint64, expected, desired uint64) uint64 {
	mask := interrupt.Disable()
	old := *ptr
	if old == expected {
		*ptr = desired
	}
	interrupt.Restore(mask)
	return old
}

//export __sync_val_compare_and_swap_8
func __sync_val_compare_and_swap_8(ptr *uint64, expected, desired uint64) uint64 {
	return doAtomicCAS64(ptr, expected, desired)
}

//export __atomic_compare_exchange_8
func __atomic_compare_exchange_8(ptr, expected *uint64, desired uint64, successOrder, failureOrder uintptr) bool {
	exp := *expected
	old := doAtomicCAS64(ptr, exp, desired)
	return old == exp
}

//go:inline
func doAtomicSwap64(ptr *uint64, new uint64) uint64 {
	mask := interrupt.Disable()
	old := *ptr
	*ptr = new
	interrupt.Restore(mask)
	return old
}

//export __sync_lock_test_and_set_8
func __sync_lock_test_and_set_8(ptr *uint64, new uint64) uint64 {
	return doAtomicSwap64(ptr, new)
}

//export __atomic_exchange_8
func __atomic_exchange_8(ptr *uint64, new uint64, ordering uintptr) uint64 {
	return doAtomicSwap64(ptr, new)
}

//go:inline
func doAtomicAdd64(ptr *uint64, value uint64) (old, new uint64) {
	mask := interrupt.Disable()
	old = *ptr
	new = old + value
	*ptr = new
	interrupt.Restore(mask)
	return old, new
}

//export __atomic_fetch_add_8
func __atomic_fetch_add_8(ptr *uint64, value uint64, ordering uintptr) uint64 {
	old, _ := doAtomicAdd64(ptr, value)
	return old
}

//export __sync_fetch_and_add_8
func __sync_fetch_and_add_8(ptr *uint64, value uint64) uint64 {
	old, _ := doAtomicAdd64(ptr, value)
	return old
}

//export __atomic_add_fetch_8
func __atomic_add_fetch_8(ptr *uint64, value uint64, ordering uintptr) uint64 {
	_, new := doAtomicAdd64(ptr, value)
	return new
}
