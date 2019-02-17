// +build gc.conservative

package runtime

// This memory manager is a textbook mark/sweep implementation, heavily inspired
// by the MicroPython garbage collector.
//
// The memory manager internally uses blocks of 4 pointers big (see
// bytesPerBlock). Every allocation first rounds up to this size to align every
// block. It will first try to find a chain of blocks that is big enough to
// satisfy the allocation. If it finds one, it marks the first one as the "head"
// and the following ones (if any) as the "tail" (see below). If it cannot find
// any free space, it will perform a garbage collection cycle and try again. If
// it still cannot find any free space, it gives up.
//
// Every block has some metadata, which is stored at the beginning of the heap.
// The four states are "free", "head", "tail", and "mark". During normal
// operation, there are no marked blocks. Every allocated object starts with a
// "head" and is followed by "tail" blocks. The reason for this distinction is
// that this way, the start and end of every object can be found easily.
//
// Metadata is stored in a special area at the beginning of the heap, in the
// area heapStart..poolStart. The actual blocks are stored in
// poolStart..heapEnd.
//
// More information:
// https://github.com/micropython/micropython/wiki/Memory-Manager
// "The Garbage Collection Handbook" by Richard Jones, Antony Hosking, Eliot
// Moss.

import (
	"unsafe"
)

// Set gcDebug to true to print debug information.
const (
	gcDebug   = false   // print debug info
	gcAsserts = gcDebug // perform sanity checks
)

// Some globals + constants for the entire GC.

const (
	wordsPerBlock      = 4 // number of pointers in an allocated block
	bytesPerBlock      = wordsPerBlock * unsafe.Sizeof(heapStart)
	stateBits          = 2 // how many bits a block state takes (see blockState type)
	blocksPerStateByte = 8 / stateBits
)

var (
	poolStart uintptr // the first heap pointer
	nextAlloc gcBlock // the next block that should be tried by the allocator
	endBlock  gcBlock // the block just past the end of the available space
)

// zeroSizedAlloc is just a sentinel that gets returned when allocating 0 bytes.
var zeroSizedAlloc uint8

// Provide some abstraction over heap blocks.

// blockState stores the four states in which a block can be. It is two bits in
// size.
type blockState uint8

const (
	blockStateFree blockState = 0 // 00
	blockStateHead blockState = 1 // 01
	blockStateTail blockState = 2 // 10
	blockStateMark blockState = 3 // 11
	blockStateMask blockState = 3 // 11
)

// String returns a human-readable version of the block state, for debugging.
func (s blockState) String() string {
	switch s {
	case blockStateFree:
		return "free"
	case blockStateHead:
		return "head"
	case blockStateTail:
		return "tail"
	case blockStateMark:
		return "mark"
	default:
		// must never happen
		return "!err"
	}
}

// The block number in the pool.
type gcBlock uintptr

// blockFromAddr returns a block given an address somewhere in the heap (which
// might not be heap-aligned).
func blockFromAddr(addr uintptr) gcBlock {
	return gcBlock((addr - poolStart) / bytesPerBlock)
}

// Return a pointer to the start of the allocated object.
func (b gcBlock) pointer() unsafe.Pointer {
	return unsafe.Pointer(b.address())
}

// Return the address of the start of the allocated object.
func (b gcBlock) address() uintptr {
	return poolStart + uintptr(b)*bytesPerBlock
}

// findHead returns the head (first block) of an object, assuming the block
// points to an allocated object. It returns the same block if this block
// already points to the head.
func (b gcBlock) findHead() gcBlock {
	for b.state() == blockStateTail {
		b--
	}
	return b
}

// findNext returns the first block just past the end of the tail. This may or
// may not be the head of an object.
func (b gcBlock) findNext() gcBlock {
	if b.state() == blockStateHead {
		b++
	}
	for b.state() == blockStateTail {
		b++
	}
	return b
}

// State returns the current block state.
func (b gcBlock) state() blockState {
	stateBytePtr := (*uint8)(unsafe.Pointer(heapStart + uintptr(b/blocksPerStateByte)))
	return blockState(*stateBytePtr>>((b%blocksPerStateByte)*2)) % 4
}

// setState sets the current block to the given state, which must contain more
// bits than the current state. Allowed transitions: from free to any state and
// from head to mark.
func (b gcBlock) setState(newState blockState) {
	stateBytePtr := (*uint8)(unsafe.Pointer(heapStart + uintptr(b/blocksPerStateByte)))
	*stateBytePtr |= uint8(newState << ((b % blocksPerStateByte) * 2))
	if gcAsserts && b.state() != newState {
		runtimePanic("gc: setState() was not successful")
	}
}

// markFree sets the block state to free, no matter what state it was in before.
func (b gcBlock) markFree() {
	stateBytePtr := (*uint8)(unsafe.Pointer(heapStart + uintptr(b/blocksPerStateByte)))
	*stateBytePtr &^= uint8(blockStateMask << ((b % blocksPerStateByte) * 2))
	if gcAsserts && b.state() != blockStateFree {
		runtimePanic("gc: markFree() was not successful")
	}
}

// unmark changes the state of the block from mark to head. It must be marked
// before calling this function.
func (b gcBlock) unmark() {
	if gcAsserts && b.state() != blockStateMark {
		runtimePanic("gc: unmark() on a block that is not marked")
	}
	clearMask := blockStateMask ^ blockStateHead // the bits to clear from the state
	stateBytePtr := (*uint8)(unsafe.Pointer(heapStart + uintptr(b/blocksPerStateByte)))
	*stateBytePtr &^= uint8(clearMask << ((b % blocksPerStateByte) * 2))
	if gcAsserts && b.state() != blockStateHead {
		runtimePanic("gc: unmark() was not successful")
	}
}

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

// alloc tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
func alloc(size uintptr) unsafe.Pointer {
	if size == 0 {
		return unsafe.Pointer(&zeroSizedAlloc)
	}

	neededBlocks := (size + (bytesPerBlock - 1)) / bytesPerBlock

	// Continue looping until a run of free blocks has been found that fits the
	// requested size.
	index := nextAlloc
	numFreeBlocks := uintptr(0)
	heapScanCount := uint8(0)
	for {
		if index == nextAlloc {
			if heapScanCount == 0 {
				heapScanCount = 1
			} else if heapScanCount == 1 {
				// The entire heap has been searched for free memory, but none
				// could be found. Run a garbage collection cycle to reclaim
				// free memory and try again.
				heapScanCount = 2
				GC()
			} else {
				// Even after garbage collection, no free memory could be found.
				runtimePanic("out of memory")
			}
		}

		// Wrap around the end of the heap.
		if index == endBlock {
			index = 0
			// Reset numFreeBlocks as allocations cannot wrap.
			numFreeBlocks = 0
		}

		// Is the block we're looking at free?
		if index.state() != blockStateFree {
			// This block is in use. Try again from this point.
			numFreeBlocks = 0
			index++
			continue
		}
		numFreeBlocks++
		index++

		// Are we finished?
		if numFreeBlocks == neededBlocks {
			// Found a big enough range of free blocks!
			nextAlloc = index
			thisAlloc := index - gcBlock(neededBlocks)
			if gcDebug {
				println("found memory:", thisAlloc.pointer(), int(size))
			}

			// Set the following blocks as being allocated.
			thisAlloc.setState(blockStateHead)
			for i := thisAlloc + 1; i != nextAlloc; i++ {
				i.setState(blockStateTail)
			}

			// Return a pointer to this allocation.
			pointer := thisAlloc.pointer()
			memzero(pointer, size)
			return pointer
		}
	}
}

func free(ptr unsafe.Pointer) {
	// TODO: free blocks on request, when the compiler knows they're unused.
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
		if looksLikePointer(root) {
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

// Sweep goes through all memory and frees unmarked memory.
func sweep() {
	freeCurrentObject := false
	for block := gcBlock(0); block < endBlock; block++ {
		switch block.state() {
		case blockStateHead:
			// Unmarked head. Free it, including all tail blocks following it.
			block.markFree()
			freeCurrentObject = true
		case blockStateTail:
			if freeCurrentObject {
				// This is a tail object following an unmarked head.
				// Free it now.
				block.markFree()
			}
		case blockStateMark:
			// This is a marked object. The next tail blocks must not be freed,
			// but the mark bit must be removed so the next GC cycle will
			// collect this object if it is unreferenced then.
			block.unmark()
			freeCurrentObject = false
		}
	}
}

// looksLikePointer returns whether this could be a pointer. Currently, it
// simply returns whether it lies anywhere in the heap. Go allows interior
// pointers so we can't check alignment or anything like that.
func looksLikePointer(ptr uintptr) bool {
	return ptr >= poolStart && ptr < heapEnd
}

// dumpHeap can be used for debugging purposes. It dumps the state of each heap
// block to standard output.
func dumpHeap() {
	println("heap:")
	for block := gcBlock(0); block < endBlock; block++ {
		switch block.state() {
		case blockStateHead:
			print("*")
		case blockStateTail:
			print("-")
		case blockStateMark:
			print("#")
		default: // free
			print("·")
		}
		if block%64 == 63 || block+1 == endBlock {
			println()
		}
	}
}

func KeepAlive(x interface{}) {
	// Unimplemented. Only required with SetFinalizer().
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}
