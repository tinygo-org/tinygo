//go:build stm32 && stm32h7

package machine

import (
	"device/arm"
	"runtime/volatile"
	"unsafe"
)

// Cortex-M7 cache maintenance by-address registers (ARMv7-M ARM Table B3-7).
var (
	scbDCIMVAC  = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE000EF5C))) // Invalidate D-cache line by address (W)
	scbDCCMVAC  = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE000EF68))) // Clean D-cache line by address (W)
	scbDCCIMVAC = (*volatile.Register32)(unsafe.Pointer(uintptr(0xE000EF70))) // Clean+Invalidate D-cache line by address (W)
)

const dCacheLineSize = 32 // bytes; fixed on Cortex-M7

// DCacheClean writes dirty cache lines covering [addr, addr+size) back to
// memory without invalidating them. Call before the CPU hands a buffer to a
// DMA controller that only reads the buffer.
func DCacheClean(addr uintptr, size uintptr) {
	arm.Asm("dsb 0xF")
	end := addr + size
	for a := addr &^ (dCacheLineSize - 1); a < end; a += dCacheLineSize {
		scbDCCMVAC.Set(uint32(a))
	}
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")
}

// DCacheInvalidate marks cache lines covering [addr, addr+size) as invalid so
// the next access re-fetches from memory. Call after a DMA write completes
// before the CPU reads the buffer.
func DCacheInvalidate(addr uintptr, size uintptr) {
	arm.Asm("dsb 0xF")
	end := addr + size
	for a := addr &^ (dCacheLineSize - 1); a < end; a += dCacheLineSize {
		scbDCIMVAC.Set(uint32(a))
	}
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")
}

// DCacheFlush cleans and invalidates cache lines covering [addr, addr+size).
// Use when the region is both written by the CPU and read by DMA (or vice
// versa) and you want to synchronize in a single pass.
func DCacheFlush(addr uintptr, size uintptr) {
	arm.Asm("dsb 0xF")
	end := addr + size
	for a := addr &^ (dCacheLineSize - 1); a < end; a += dCacheLineSize {
		scbDCCIMVAC.Set(uint32(a))
	}
	arm.Asm("dsb 0xF")
	arm.Asm("isb 0xF")
}
