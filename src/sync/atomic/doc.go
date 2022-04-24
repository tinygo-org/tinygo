// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package atomic provides low-level atomic memory primitives
// useful for implementing synchronization algorithms.
//
// These functions require great care to be used correctly.
// Except for special, low-level applications, synchronization is better
// done with channels or the facilities of the sync package.
// Share memory by communicating;
// don't communicate by sharing memory.
//
// The swap operation, implemented by the SwapT functions, is the atomic
// equivalent of:
//
//	old = *addr
//	*addr = new
//	return old
//
// The compare-and-swap operation, implemented by the CompareAndSwapT
// functions, is the atomic equivalent of:
//
//	if *addr == old {
//		*addr = new
//		return true
//	}
//	return false
//
// The add operation, implemented by the AddT functions, is the atomic
// equivalent of:
//
//	*addr += delta
//	return *addr
//
// The load and store operations, implemented by the LoadT and StoreT
// functions, are the atomic equivalents of "return *addr" and
// "*addr = val".
//
package atomic

import (
	"unsafe"
)

// BUG(rsc): On 386, the 64-bit functions use instructions unavailable before the Pentium MMX.
//
// On non-Linux ARM, the 64-bit functions use instructions unavailable before the ARMv6k core.
//
// On ARM, 386, and 32-bit MIPS, it is the caller's responsibility
// to arrange for 64-bit alignment of 64-bit words accessed atomically.
// The first word in a variable or in an allocated struct, array, or slice can
// be relied upon to be 64-bit aligned.

// SwapInt32 atomically stores new into *addr and returns the previous *addr value.
// go:inline
func SwapInt32(addr *int32, new int32) int32 {
	return llvm_SwapInt32(addr, new)
}

func llvm_SwapInt32(addr *int32, new int32) (old int32)

// SwapInt64 atomically stores new into *addr and returns the previous *addr value.
// go:inline
func SwapInt64(addr *int64, new int64) int64 {
	return llvm_SwapInt64(addr, new)
}

func llvm_SwapInt64(addr *int64, new int64) (old int64)

// SwapUint32 atomically stores new into *addr and returns the previous *addr value.
// go:inline
func SwapUint32(addr *uint32, new uint32) uint32 {
	return llvm_SwapUint32(addr, new)
}

func llvm_SwapUint32(addr *uint32, new uint32) (old uint32)

// SwapUint64 atomically stores new into *addr and returns the previous *addr value.
// go:inline
func SwapUint64(addr *uint64, new uint64) uint64 {
	return llvm_SwapUint64(addr, new)
}

func llvm_SwapUint64(addr *uint64, new uint64) (old uint64)

// SwapUintptr atomically stores new into *addr and returns the previous *addr value.
// go:inline
func SwapUintptr(addr *uintptr, new uintptr) uintptr {
	return llvm_SwapUintptr(addr, new)
}

func llvm_SwapUintptr(addr *uintptr, new uintptr) (old uintptr)

// SwapPointer atomically stores new into *addr and returns the previous *addr value.
func SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer {
	return llvm_SwapPointer(addr, new)
}

func llvm_SwapPointer(addr *unsafe.Pointer, new unsafe.Pointer) (old unsafe.Pointer)

// CompareAndSwapInt32 executes the compare-and-swap operation for an int32 value.
// go:inline
func CompareAndSwapInt32(addr *int32, old, new int32) bool {
	return llvm_CompareAndSwapInt32(addr, old, new)
}

func llvm_CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)

// CompareAndSwapInt64 executes the compare-and-swap operation for an int64 value.
// go:inline
func CompareAndSwapInt64(addr *int64, old, new int64) bool {
	return llvm_CompareAndSwapInt64(addr, old, new)
}

func llvm_CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool)

// CompareAndSwapUint32 executes the compare-and-swap operation for a uint32 value.
// go:inline
func CompareAndSwapUint32(addr *uint32, old, new uint32) bool {
	return llvm_CompareAndSwapUint32(addr, old, new)
}

func llvm_CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool)

// CompareAndSwapUint64 executes the compare-and-swap operation for a uint64 value.
// go:inline
func CompareAndSwapUint64(addr *uint64, old, new uint64) bool {
	return llvm_CompareAndSwapUint64(addr, old, new)
}

func llvm_CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool)

// CompareAndSwapUintptr executes the compare-and-swap operation for a uintptr value.
// go:inline
func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) bool {
	return llvm_CompareAndSwapUintptr(addr, old, new)
}

func llvm_CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool)

// CompareAndSwapPointer executes the compare-and-swap operation for a unsafe.Pointer value.
// go:inline
func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) bool {
	return llvm_CompareAndSwapPointer(addr, old, new)
}

func llvm_CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool)

// AddInt32 atomically adds delta to *addr and returns the new value.
// go:inline
func AddInt32(addr *int32, delta int32) int32 {
	return llvm_AddInt32(addr, delta)
}

func llvm_AddInt32(addr *int32, delta int32) (new int32)

// AddUint32 atomically adds delta to *addr and returns the new value.
// To subtract a signed positive constant value c from x, do AddUint32(&x, ^uint32(c-1)).
// In particular, to decrement x, do AddUint32(&x, ^uint32(0)).
// go:inline
func AddUint32(addr *uint32, delta uint32) uint32 {
	return llvm_AddUint32(addr, delta)
}

func llvm_AddUint32(addr *uint32, delta uint32) (new uint32)

// AddInt64 atomically adds delta to *addr and returns the new value.
// go:inline
func AddInt64(addr *int64, delta int64) int64 {
	return llvm_AddInt64(addr, delta)
}

func llvm_AddInt64(addr *int64, delta int64) (new int64)

// AddUint64 atomically adds delta to *addr and returns the new value.
// To subtract a signed positive constant value c from x, do AddUint64(&x, ^uint64(c-1)).
// In particular, to decrement x, do AddUint64(&x, ^uint64(0)).
// go:inline
func AddUint64(addr *uint64, delta uint64) uint64 {
	return llvm_AddUint64(addr, delta)
}

func llvm_AddUint64(addr *uint64, delta uint64) (new uint64)

// AddUintptr atomically adds delta to *addr and returns the new value.
// go:inline
func AddUintptr(addr *uintptr, delta uintptr) uintptr {
	return llvm_AddUintptr(addr, delta)
}

func llvm_AddUintptr(addr *uintptr, delta uintptr) (new uintptr)

// LoadInt32 atomically loads *addr.
// go:inline
func LoadInt32(addr *int32) int32 {
	return llvm_LoadInt32(addr)
}

func llvm_LoadInt32(addr *int32) (val int32)

// LoadInt64 atomically loads *addr.
// go:inline
func LoadInt64(addr *int64) int64 {
	return llvm_LoadInt64(addr)
}

func llvm_LoadInt64(addr *int64) (val int64)

// LoadUint32 atomically loads *addr.
// go:inline
func LoadUint32(addr *uint32) (val uint32) {
	return llvm_LoadUint32(addr)
}

func llvm_LoadUint32(addr *uint32) (val uint32)

// LoadUint64 atomically loads *addr.
// go:inline
func LoadUint64(addr *uint64) uint64 {
	return llvm_LoadUint64(addr)
}

func llvm_LoadUint64(addr *uint64) (val uint64)

// LoadUintptr atomically loads *addr.
// go:inline
func LoadUintptr(addr *uintptr) uintptr {
	return llvm_LoadUintptr(addr)
}

func llvm_LoadUintptr(addr *uintptr) (val uintptr)

// LoadPointer atomically loads *addr.
// go:inline
func LoadPointer(addr *unsafe.Pointer) unsafe.Pointer {
	return llvm_LoadPointer(addr)
}

func llvm_LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer)

// StoreInt32 atomically stores val into *addr.
// go:inline
func StoreInt32(addr *int32, val int32) {
	llvm_StoreInt32(addr, val)
}

func llvm_StoreInt32(addr *int32, val int32)

// StoreInt64 atomically stores val into *addr.
// go:inline
func StoreInt64(addr *int64, val int64) {
	llvm_StoreInt64(addr, val)
}

func llvm_StoreInt64(addr *int64, val int64)

// StoreUint32 atomically stores val into *addr.
// go:inline
func StoreUint32(addr *uint32, val uint32) {
	llvm_StoreUint32(addr, val)
}

func llvm_StoreUint32(addr *uint32, val uint32)

// StoreUint64 atomically stores val into *addr.
func StoreUint64(addr *uint64, val uint64) {
	llvm_StoreUint64(addr, val)
}

func llvm_StoreUint64(addr *uint64, val uint64)

// StoreUintptr atomically stores val into *addr.
func StoreUintptr(addr *uintptr, val uintptr) {
	llvm_StoreUintptr(addr, val)
}

func llvm_StoreUintptr(addr *uintptr, val uintptr)

// StorePointer atomically stores val into *addr.
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer) {
	llvm_StorePointer(addr, val)
}

func llvm_StorePointer(addr *unsafe.Pointer, val unsafe.Pointer)
