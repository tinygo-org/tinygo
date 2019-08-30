package runtime

import _ "unsafe"

// This file contains implementations for the sync/atomic package.

// All implementations assume there are no goroutines, threads or interrupts.

//go:linkname loadUint64 sync/atomic.LoadUint64
func loadUint64(addr *uint64) uint64 {
	return *addr
}

//go:linkname storeUint32 sync/atomic.StoreUint32
func storeUint32(addr *uint32, val uint32) {
	*addr = val
}

//go:linkname compareAndSwapUint64 sync/atomic.CompareAndSwapUint64
func compareAndSwapUint64(addr *uint64, old, new uint64) bool {
	if *addr == old {
		*addr = new
		return true
	}
	return false
}
