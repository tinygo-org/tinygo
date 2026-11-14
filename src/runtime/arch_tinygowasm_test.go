//go:build tinygo.wasm && (gc.leaking || gc.conservative || gc.precise)

package runtime_test

import (
	"runtime"
	"testing"
	_ "unsafe"
)

//export llvm.wasm.memory.size.i32
func wasm_memory_size(index int32) int32

//export llvm.wasm.memory.grow.i32
func wasm_memory_grow(index int32, delta int32) int32

//go:linkname setHeapEnd runtime.setHeapEnd
func setHeapEnd(newHeapEnd uintptr)

//go:linkname growHeap runtime.growHeap
func growHeap() bool

func TestGrowHeapAbove2GiB(t *testing.T) {
	const (
		wasmMemoryIndex = 0
		wasmPageSize    = 64 * 1024
		targetPages     = 32800 // > 2 GiB (32768 pages) + heapStart
		maxPages        = 65535
	)

	// Advance linear memory past 2 GiB (32768 pages) so doubling memorySize
	// would exceed the 65536-page wasm32 limit.
	cur := wasm_memory_size(wasmMemoryIndex)
	if maxPages-cur <= cur/64 {
		t.Skip("wasm memory already at cap")
	}
	if cur < targetPages {
		if wasm_memory_grow(wasmMemoryIndex, targetPages-cur) == -1 {
			t.Skip("host could not grow wasm memory past 2 GiB")
		}
		setHeapEnd(uintptr(wasm_memory_size(wasmMemoryIndex) * wasmPageSize))
	}

	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	if before.Sys <= 1<<31 {
		t.Fatalf("expected MemStats.Sys > 2 GiB before growHeap, got %#x", before.Sys)
	}

	if !growHeap() {
		t.Fatalf("growHeap failed at %d pages (Sys=%#x)", wasm_memory_size(wasmMemoryIndex), before.Sys)
	}

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	if after.Sys <= before.Sys {
		t.Fatalf("expected MemStats.Sys to grow above %#x, got %#x", before.Sys, after.Sys)
	}

	buf := make([]byte, 256<<20)
	buf[0] = 1
	buf[len(buf)-1] = 2
	if buf[0] != 1 || buf[len(buf)-1] != 2 {
		t.Fatalf("unexpected buffer contents")
	}

	if wasm_memory_size(wasmMemoryIndex) == maxPages && growHeap() {
		t.Fatalf("expected growHeap to return false at %d-page cap", maxPages)
	}
}
