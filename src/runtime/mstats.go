//go:build gc.conservative
// +build gc.conservative

package runtime

// Memory statistics

// Subset of memory statistics from upstream Go.
// Works with conservative gc only.

// A MemStats records statistics about the memory allocator.
type MemStats struct {
	// General statistics.

	// Sys is the total bytes of memory obtained from the OS.
	//
	// Sys is the sum of the XSys fields below. Sys measures the
	// address space reserved by the runtime for the
	// heap, stacks, and other internal data structures.
	Sys uint64

	// Heap memory statistics.

	// HeapSys is bytes of heap memory, total.
	//
	// In TinyGo unlike upstream Go, we make no distinction between
	// regular heap blocks used by escaped-to-the-heap variables and
	// blocks occupied by goroutine stacks,
	// all such blocks are marked as in-use, see HeapInuse below.
	HeapSys uint64

	// HeapIdle is bytes in idle (unused) blocks.
	HeapIdle uint64

	// HeapInuse is bytes in in-use blocks.
	HeapInuse uint64

	// HeapReleased is bytes of physical memory returned to the OS.
	HeapReleased uint64

	// Off-heap memory statistics.
	//
	// The following statistics measure runtime-internal
	// structures that are not allocated from heap memory (usually
	// because they are part of implementing the heap).

	// GCSys is bytes of memory in garbage collection metadata.
	GCSys uint64
}

// ReadMemStats populates m with memory statistics.
//
// The returned memory statistics are up to date as of the
// call to ReadMemStats. This would not do GC implicitly for you.
func ReadMemStats(m *MemStats) {
	m.HeapIdle = 0
	m.HeapInuse = 0
	for block := gcBlock(0); block < endBlock; block++ {
		bstate := block.state()
		if bstate == blockStateFree {
			m.HeapIdle += uint64(bytesPerBlock)
		} else {
			m.HeapInuse += uint64(bytesPerBlock)
		}
	}
	m.HeapReleased = 0 // always 0, we don't currently release memory back to the OS.
	m.HeapSys = m.HeapInuse + m.HeapIdle
	m.GCSys = uint64(heapEnd - uintptr(metadataStart))
	m.Sys = uint64(heapEnd - heapStart)
}
