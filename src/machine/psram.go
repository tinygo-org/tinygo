//go:build machine.psram

package machine

import "unsafe"

//go:linkname runtimePSRAM runtime.machinePSRAM
func runtimePSRAM() (uintptr, int)

// PSRAM returns a slice covering the remaining external PSRAM after any
// static //go:section .psram variables. Returns nil if none is left.
// The returned memory is not zero-initialized and is not scanned by the GC,
// so it must not be used to store pointers.
// Atomic operations do not work in PSRAM on RP2350, so do not put sync/atomic
// values or sync types in PSRAM.
func PSRAM() []byte {
	start, length := runtimePSRAM()
	if length == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(start)), length)
}
