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
	markStackSize      = 4 * unsafe.Sizeof((*int)(nil)) // number of to-be-marked blocks to queue before forcing a rescan
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

// setAlloc sets states of a series of gc blocks to list it as allocated.
func (b gcBlock) setAlloc(len uintptr) {
	if len == 0 {
		return
	}

	b.setState(blockStateHead)
	for i := uintptr(1); i < len; i++ {
		(b + gcBlock(i)).setState(blockStateTail)
	}
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

// gcFreeSpan stores information about a span of free blocks.
// It is stored in the first free block of the span.
type gcFreeSpan struct {
	// next is a pointer to the next free span of equal size.
	next *gcFreeSpan
}

// gcBlock returns the first block in the free span.
func (span *gcFreeSpan) gcBlock() gcBlock {
	return blockFromAddr(uintptr(unsafe.Pointer(span)))
}

// gcFreeRootSpan stores information about available free block spans.
type gcFreeRootSpan struct {
	gcFreeSpan

	// len is the length of the free span (measured in blocks).
	len uintptr

	// nextSize is a pointer to the root span of the next smallest span size.
	nextSize *gcFreeRootSpan
}

// gcFreeSpans is a tree which is used to track spans of free blocks in the heap.
// This tree is asymmetric, and can also be viewed as a linked list of linked lists.
// The outer list tracks available node sizes, and is sorted in ascending size order.
// The inner lists are of individual free spans of a given size, sorted in no particular order.
// This ensures that the worst case lookup time for the smallest span fitting an allocation is proportional to the size of the allocation.
// However, the minimum heap size for n different sized spans is ½n² + (3/2)n - 2.
// Therefore the worst case lookup complexity with respect to heap size is O(√(8n+17)) ≈ O(√n).
var gcFreeSpans *gcFreeRootSpan

// insertFreeSpan inserts a new span of free blocks into the span tree.
func insertFreeSpan(block gcBlock, len uintptr) {
	if len == 0 {
		return
	}

	// Search for the appropriate root insertion point.
	rins := &gcFreeSpans
	for *rins != nil && (*rins).len < len {
		rins = &(*rins).nextSize
	}

	if root := *rins; root != nil && root.len == len {
		// There is already a root for this size.
		// Add a span to this root.
		span := (*gcFreeSpan)(block.pointer())
		*span = gcFreeSpan{
			next: root.next,
		}
		root.next = span
		return
	}

	// Create a new root for this size.
	root := (*gcFreeRootSpan)(block.pointer())
	*root = gcFreeRootSpan{
		len:      len,
		nextSize: *rins,
	}
	*rins = root
}

// rebuildSpanTree rebuilds the free span tree.
// This is used after the GC sweep.
func rebuildSpanTree() {
	// Clear the existing span tree.
	gcFreeSpans = nil

	// Search for all free blocks.
	for block := gcBlock(0); block < endBlock; block++ {
		if block.state() != blockStateFree {
			// This block is in use.
			continue
		}

		// This is the first block in a free span.
		spanStart := block

		// Find the end of the free span.
		for block < endBlock && block.state() == blockStateFree {
			block++
		}

		// Add the span to the tree.
		insertFreeSpan(spanStart, uintptr(block-spanStart))
	}
}

// allocBlocks allocates a series of gc blocks.
// It attempts to use the smallest possible free span.
func allocBlocks(len uintptr) (gcBlock, bool) {
	// Search for a root with a span length of at least len.
	rins := &gcFreeSpans
	for *rins != nil && (*rins).len < len {
		rins = &(*rins).nextSize
	}
	if *rins == nil {
		// There are no sufficiently-large free spans.
		return 0, false
	}

	// Get a free span.
	root := *rins
	var span *gcFreeSpan
	if span = root.next; span != nil {
		// Pop a leaf span off of the tree.
		root.next = span.next
	} else {
		// There is only one free span of this size.
		// Extract the root.
		*rins = root.nextSize
		span = &root.gcFreeSpan
	}

	if root.len > len {
		// The span is longer than requested.
		// Insert the remainder of the span.
		insertFreeSpan(span.gcBlock()+gcBlock(len), root.len-len)
	}

	return span.gcBlock(), true
}

// dumpFreeSpanTree prints the available free spans.
// This can be used in order to debug heap fragmentation.
func dumpFreeSpanTree() {
	if gcFreeSpans == nil {
		println("no free spans")
		return
	}

	println("free spans:")
	for root := gcFreeSpans; root != nil; root = root.nextSize {
		var n uint
		for s := &root.gcFreeSpan; s != nil; s = s.next {
			n++
		}
		print("  ", n, "x ", uint(root.len*bytesPerBlock), " bytes: ", root)
		for s := root.next; s != nil; s = s.next {
			print(", ", s)
		}
		println()
	}
}

// Initialize the memory allocator.
// No memory may be allocated before this is called.
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

	// Insert an initial free span.
	insertFreeSpan(0, numBlocks)
}

// alloc tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
//go:noinline
func alloc(size uintptr) unsafe.Pointer {
	if size == 0 {
		return unsafe.Pointer(&zeroSizedAlloc)
	}

	neededBlocks := (size + (bytesPerBlock - 1)) / bytesPerBlock
	if neededBlocks > uintptr(endBlock) {
		runtimePanic("oversized allocation")
	}

	var gcRun bool
tryAlloc:
	// Find a span of free blocks.
	block, ok := allocBlocks(neededBlocks)
	if !ok {
		// There is no sufficiently large span of free blocks available.

		if !gcRun {
			// Run the garbage collector and try again.
			GC()
			gcRun = true
			goto tryAlloc
		}

		// There is not enough space for the allocation.
		runtimePanic("out of memory")
	}

	if gcDebug {
		println("found memory:", block.pointer(), int(size))
	}

	// Set the states of the blocks as allocated.
	block.setAlloc(neededBlocks)

	// Clear the allocation.
	pointer := block.pointer()
	memzero(pointer, size)

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
	finishMark()

	// Sweep phase: free all non-marked objects and unmark marked objects for
	// the next collection cycle.
	sweep()

	// Build a tree of free memory spans.
	rebuildSpanTree()

	// Show how much has been sweeped, for debugging.
	if gcDebug {
		dumpHeap()
		dumpFreeSpanTree()
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

// stackOverflow is a flag which is set when the GC scans too deep while marking.
// After it is set, all marked allocations must be re-scanned.
var stackOverflow bool

// startMark starts the marking process on a root and all of its children.
func startMark(root gcBlock) {
	var stack [markStackSize]gcBlock
	stack[0] = root
	root.setState(blockStateMark)
	stackLen := 1
	for stackLen > 0 {
		// Pop a block off of the stack.
		stackLen--
		block := stack[stackLen]
		if gcDebug {
			println("stack popped, remaining stack:", stackLen)
		}

		// Scan all pointers inside the block.
		start, end := block.address(), block.findNext().address()
		for addr := start; addr != end; addr += unsafe.Alignof(addr) {
			// Load the word.
			word := *(*uintptr)(unsafe.Pointer(addr))

			if !looksLikePointer(word) {
				// Not a heap pointer.
				continue
			}

			// Find the corresponding memory block.
			referencedBlock := blockFromAddr(word)

			if referencedBlock.state() == blockStateFree {
				// The to-be-marked object doesn't actually exist.
				// This is probably a false positive.
				if gcDebug {
					println("found reference to free memory:", word, "at:", addr)
				}
				continue
			}

			// Move to the block's head.
			referencedBlock = referencedBlock.findHead()

			if referencedBlock.state() == blockStateMark {
				// The block has already been marked by something else.
				continue
			}

			// Mark block.
			if gcDebug {
				println("marking block:", referencedBlock)
			}
			referencedBlock.setState(blockStateMark)

			if stackLen == len(stack) {
				// The stack is full.
				// It is necessary to rescan all marked blocks once we are done.
				stackOverflow = true
				if gcDebug {
					println("gc stack overflowed")
				}
				continue
			}

			// Push the pointer onto the stack to be scanned later.
			stack[stackLen] = referencedBlock
			stackLen++
		}
	}
}

// finishMark finishes the marking process by processing all stack overflows.
func finishMark() {
	for stackOverflow {
		// Re-mark all blocks.
		stackOverflow = false
		for block := gcBlock(0); block < endBlock; block++ {
			if block.state() != blockStateMark {
				// Block is not marked, so we do not need to rescan it.
				continue
			}

			// Re-mark the block.
			startMark(block)
		}
	}
}

// mark a GC root at the address addr.
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
			startMark(head)
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
