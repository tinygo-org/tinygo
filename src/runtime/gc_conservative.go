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
	if gcAsserts && (addr < poolStart || addr >= heapEnd) {
		runtimePanic("gc: trying to get block from invalid address")
	}
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
	if gcAsserts {
		if b.state() != blockStateHead && b.state() != blockStateMark {
			runtimePanic("gc: found tail without head")
		}
	}
	return b
}

// findNext returns the first block just past the end of the tail. This may or
// may not be the head of an object.
func (b gcBlock) findNext() gcBlock {
	if b.state() == blockStateHead || b.state() == blockStateMark {
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
func initHeap() {
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

	// Initialize the freeList with a single node encompassing the entire heap.
	freeList = endBlock
	freeListInsert(0, numBlocks)
}

var freeList gcBlock

type freeListNode struct {
	prev, next *freeListNode
	size       uintptr
}

// freeListHead returns a pointer to the first entry in the free list.
// If the free list is empty, this returns nil.
func freeListHead() *freeListNode {
	if freeList == endBlock {
		return nil
	}

	return (*freeListNode)(unsafe.Pointer(freeList.pointer()))
}

// freeListFind searches the free list for the smallest gap that is at least the specified size (in blocks).
func freeListFind(size uintptr) gcBlock {
	for node := freeListHead(); node != nil; node = node.next {
		if node.size >= size {
			return blockFromAddr(uintptr(unsafe.Pointer(node)))
		}
	}

	// The entire free list was scanned, and no blocks were sufficiently large.
	return endBlock
}

// freeListInsert adds a free region starting at block, with the specified size (in blocks).
func freeListInsert(block gcBlock, size uintptr) {
	// Search for an insertion point.
	var prev *freeListNode
	next := freeListHead()
	for ; next != nil && next.size < size; next = next.next {
		prev = next
	}

	// Build the new free list entry.
	node := (*freeListNode)(block.pointer())
	*node = freeListNode{
		prev: prev,
		next: next,
		size: size,
	}

	if prev != nil {
		// Update previous entry to reference this.
		prev.next = node
	} else {
		// Update list head.
		freeList = block
	}

	if next != nil {
		// Update next entry to reference this.
		next.prev = node
	}
}

// freeListRemove removes the gap starting at the specified block from the free list.
// It returns the size of the gap that was removed.
func freeListRemove(block gcBlock) uintptr {
	node := (*freeListNode)(block.pointer())
	prev, next := node.prev, node.next

	if prev != nil {
		// Update previous node.
		prev.next = next
	} else {
		// Update list head.
		if next == nil {
			// List is now empty.
			freeList = endBlock
		} else {
			// The former next node becomes the head.
			freeList = blockFromAddr(uintptr(unsafe.Pointer(next)))
		}
	}

	if next != nil {
		// Update next node.
		next.prev = prev
	}

	return node.size
}

func freeListDump() {
	println("free list:")
	for node := freeListHead(); node != nil; node = node.next {
		println("", node, "size:", node.size)
	}
}

// alloc tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
//go:noinline
func alloc(size uintptr) unsafe.Pointer {
	if size == 0 {
		return unsafe.Pointer(&zeroSizedAlloc)
	}

	neededBlocks := (size + (bytesPerBlock - 1)) / bytesPerBlock

	// Find an unused gap in the heap to fill with the allocation.
	var ranGC bool
findMem:
	block := freeListFind(neededBlocks)
	if block == endBlock {
		// There are no available gaps big enough for the allocation.
		if ranGC {
			// Even after garbage collection, no free memory could be found.
			if gcDebug {
				freeListDump()
				dumpHeap()
			}
			runtimePanic("out of memory")
		}

		// Run the garbage collector and try again.
		GC()
		ranGC = true
		goto findMem
	}

	// Update the free list.
	gapSize := freeListRemove(block)
	if gcDebug {
		println("using gap", block, "of size", gapSize)
	}
	if gapSize > neededBlocks {
		// Create a new free list entry with the leftover space.
		if gcDebug {
			println("creating shrunk gap", block+gcBlock(neededBlocks), "of size", gapSize-neededBlocks)
		}
		freeListInsert(block+gcBlock(neededBlocks), gapSize-neededBlocks)
	}

	// Update block metadata.
	block.setState(blockStateHead)
	for offset := uintptr(1); offset < neededBlocks; offset++ {
		(block + gcBlock(offset)).setState(blockStateTail)
	}

	// Return a pointer to this allocation.
	pointer := block.pointer()
	memzero(pointer, neededBlocks*bytesPerBlock)
	return pointer
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
	markGlobals()
	markStack()

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
	if gcAsserts {
		if start >= end {
			runtimePanic("gc: unexpected range to mark")
		}
	}

	for addr := start; addr != end; addr += unsafe.Alignof(addr) {
		root := *(*uintptr)(unsafe.Pointer(addr))
		markRoot(addr, root)
	}
}

func markRoot(addr, root uintptr) {
	if looksLikePointer(root) {
		block := blockFromAddr(root)
		if block.state() == blockStateFree {
			// The to-be-marked object doesn't actually exist.
			// This could either be a dangling pointer (oops!) but most likely
			// just a false positive.
			return
		}
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

// Sweep goes through all memory and frees unmarked memory.
func sweep() {
	// Clear free list. It will be rebuilt while sweeping.
	freeList = endBlock

	// Sweep and rebuild free list.
	var freeStart gcBlock
	freeStarted := false
	for block := gcBlock(0); block < endBlock; block++ {
		state := block.state()
		switch state {
		case blockStateFree:
			if !freeStarted {
				// Start a new free block.
				freeStart = block
				freeStarted = true
			}
		case blockStateHead:
			// Unmarked head. Free it, including all tail blocks following it.
			block.markFree()
			if !freeStarted {
				// Start a new free block with this head.
				freeStart = block
				freeStarted = true
			}
		case blockStateTail:
			if freeStarted {
				// This is a tail object following an unmarked head.
				// Free it now.
				block.markFree()
			}
		case blockStateMark:
			// This is a marked object. The next tail blocks must not be freed,
			// but the mark bit must be removed so the next GC cycle will
			// collect this object if it is unreferenced then.
			block.unmark()

			if freeStarted {
				// Add the previous gap to the free list.
				freeListInsert(freeStart, uintptr(block-freeStart))
				freeStarted = false
			}
		}
	}

	if freeStarted {
		// Add the last gap to the free list.
		freeListInsert(freeStart, uintptr(endBlock-freeStart))
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
