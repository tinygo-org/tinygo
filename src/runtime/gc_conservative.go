// +build gc.conservative

package runtime

import (
	"unsafe"
)

// Initialize the memory allocator.
// No memory may be allocated before this is called. That means the runtime and
// any packages the runtime depends upon may not allocate memory during package
// initialization.
func init() {
	totalSize := heapEnd - heapStart

	// Allocate some memory to keep 2 bits of information about every block.
	metadataSize := totalSize / (blocksPerStateByte * bytesPerBlock)

	// Align the pool.
	poolStart = (heapStart + metadataSize + (bytesPerBlock - 1)) &^ (bytesPerBlock - 1)
	poolEnd := heapEnd &^ (bytesPerBlock - 1)
	numBlocks := (poolEnd - poolStart) / bytesPerBlock
	endBlock = gcBlock(numBlocks)
	if gcDebug {
		println("heapStart:        ", heapStart)
		println("heapEnd:          ", heapEnd)
		println("total size:       ", totalSize)
		println("metadata size:    ", metadataSize)
		println("poolStart:        ", poolStart)
		println("# of blocks:      ", numBlocks)
		println("# of block states:", metadataSize*blocksPerStateByte)
	}
	if gcAsserts && metadataSize*blocksPerStateByte < numBlocks {
		// sanity check
		runtimePanic("gc: metadata array is too small")
	}

	// Set all block states to 'free'.
	memzero(unsafe.Pointer(heapStart), metadataSize)
}

func alloc(size uintptr) unsafe.Pointer {
	return heapAlloc(size)
}

// GC performs a garbage collection cycle.
func GC() {
	if gcDebug {
		println("running collection cycle...")
	}

	// Mark phase: mark all reachable objects, recursively.
	markRoots(globalsStart, globalsEnd)
	markRoots(getCurrentStackPointer(), stackTop) // assume a descending stack

	// Sweep phase: free all non-marked objects and unmark marked objects for
	// the next collection cycle.
	sweep()

	// Show how much has been sweeped, for debugging.
	if gcDebug {
		dumpHeap()
	}
}

// markRoots reads all pointers from start to end (exclusive) and if they look
// like a heap pointer and are unmarked, marks them and scans that object as
// well (recursively). The start and end parameters must be valid pointers and
// must be aligned.
func markRoots(start, end uintptr) {
	if gcDebug {
		println("mark from", start, "to", end, int(end-start))
	}

	for addr := start; addr != end; addr += unsafe.Sizeof(addr) {
		root := *(*uintptr)(unsafe.Pointer(addr))
		if addressOnHeap(root) {
			block := blockFromAddr(root)
			head := block.findHead()
			if head.state() != blockStateMark {
				if gcDebug {
					println("found unmarked pointer", root, "at address", addr)
				}
				head.setState(blockStateMark)
				next := block.findNext()
				// TODO: avoid recursion as much as possible
				markRoots(head.address(), next.address())
			}
		}
	}
}
