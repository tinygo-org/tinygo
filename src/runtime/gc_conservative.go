//go:build gc.conservative
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
// Every block has some metadata, which is stored at the end of the heap.
// The four states are "free", "head", "tail", and "mark". During normal
// operation, there are no marked blocks. Every allocated object starts with a
// "head" and is followed by "tail" blocks. The reason for this distinction is
// that this way, the start and end of every object can be found easily.
//
// Metadata is stored in a special area at the end of the heap, in the area
// metadataStart..heapEnd. The actual blocks are stored in
// heapStart..metadataStart.
//
// More information:
// https://github.com/micropython/micropython/wiki/Memory-Manager
// "The Garbage Collection Handbook" by Richard Jones, Antony Hosking, Eliot
// Moss.

import (
	"internal/task"
	"runtime/interrupt"
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
	metadataStart unsafe.Pointer // pointer to the start of the heap metadata
	nextAlloc     gcBlock        // the next block that should be tried by the allocator
	endBlock      gcBlock        // the block just past the end of the available space
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
	if gcAsserts && (addr < heapStart || addr >= uintptr(metadataStart)) {
		runtimePanic("gc: trying to get block from invalid address")
	}
	return gcBlock((addr - heapStart) / bytesPerBlock)
}

// Return a pointer to the start of the allocated object.
func (b gcBlock) pointer() unsafe.Pointer {
	return unsafe.Pointer(b.address())
}

// Return the address of the start of the allocated object.
func (b gcBlock) address() uintptr {
	return heapStart + uintptr(b)*bytesPerBlock
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
	stateBytePtr := (*uint8)(unsafe.Pointer(uintptr(metadataStart) + uintptr(b/blocksPerStateByte)))
	return blockState(*stateBytePtr>>((b%blocksPerStateByte)*stateBits)) & blockStateMask
}

// setState sets the current block to the given state, which must contain more
// bits than the current state. Allowed transitions: from free to any state and
// from head to mark.
func (b gcBlock) setState(newState blockState) {
	stateBytePtr := (*uint8)(unsafe.Pointer(uintptr(metadataStart) + uintptr(b/blocksPerStateByte)))
	*stateBytePtr |= uint8(newState << ((b % blocksPerStateByte) * stateBits))
	if gcAsserts && b.state() != newState {
		runtimePanic("gc: setState() was not successful")
	}
}

// markFree sets the block state to free, no matter what state it was in before.
func (b gcBlock) markFree() {
	stateBytePtr := (*uint8)(unsafe.Pointer(uintptr(metadataStart) + uintptr(b/blocksPerStateByte)))
	*stateBytePtr &^= uint8(blockStateMask << ((b % blocksPerStateByte) * stateBits))
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
	stateBytePtr := (*uint8)(unsafe.Pointer(uintptr(metadataStart) + uintptr(b/blocksPerStateByte)))
	*stateBytePtr &^= uint8(clearMask << ((b % blocksPerStateByte) * stateBits))
	if gcAsserts && b.state() != blockStateHead {
		runtimePanic("gc: unmark() was not successful")
	}
}

// Initialize the memory allocator.
// No memory may be allocated before this is called. That means the runtime and
// any packages the runtime depends upon may not allocate memory during package
// initialization.
func initHeap() {
	calculateHeapAddresses()

	// Set all block states to 'free'.
	metadataSize := heapEnd - uintptr(metadataStart)
	memzero(unsafe.Pointer(metadataStart), metadataSize)
}

// setHeapEnd is called to expand the heap. The heap can only grow, not shrink.
// Also, the heap should grow substantially each time otherwise growing the heap
// will be expensive.
func setHeapEnd(newHeapEnd uintptr) {
	if gcAsserts && newHeapEnd <= heapEnd {
		panic("gc: setHeapEnd didn't grow the heap")
	}

	// Save some old variables we need later.
	oldMetadataStart := metadataStart
	oldMetadataSize := heapEnd - uintptr(metadataStart)

	// Increase the heap. After setting the new heapEnd, calculateHeapAddresses
	// will update metadataStart and the memcpy will copy the metadata to the
	// new location.
	// The new metadata will be bigger than the old metadata, but a simple
	// memcpy is fine as it only copies the old metadata and the new memory will
	// have been zero initialized.
	heapEnd = newHeapEnd
	calculateHeapAddresses()
	memcpy(metadataStart, oldMetadataStart, oldMetadataSize)

	// Note: the memcpy above assumes the heap grows enough so that the new
	// metadata does not overlap the old metadata. If that isn't true, memmove
	// should be used to avoid corruption.
	// This assert checks whether that's true.
	if gcAsserts && uintptr(metadataStart) < uintptr(oldMetadataStart)+oldMetadataSize {
		panic("gc: heap did not grow enough at once")
	}
}

// calculateHeapAddresses initializes variables such as metadataStart and
// numBlock based on heapStart and heapEnd.
//
// This function can be called again when the heap size increases. The caller is
// responsible for copying the metadata to the new location.
func calculateHeapAddresses() {
	if GOARCH == "wasm" {
		// This is a workaround for a bug in wasm-ld: wasm-ld doesn't always
		// align __heap_base and when this memory is shared through an API, it
		// might result in unaligned memory. For details, see:
		// https://reviews.llvm.org/D106499
		// It should be removed once we switch to LLVM 13, where this is fixed.
		heapStart = align(heapStart)
	}
	totalSize := heapEnd - heapStart

	// Allocate some memory to keep 2 bits of information about every block.
	metadataSize := (totalSize + blocksPerStateByte*bytesPerBlock) / (1 + blocksPerStateByte*bytesPerBlock)
	metadataStart = unsafe.Pointer(heapEnd - metadataSize)

	// Use the rest of the available memory as heap.
	numBlocks := (uintptr(metadataStart) - heapStart) / bytesPerBlock
	endBlock = gcBlock(numBlocks)
	if gcDebug {
		println("heapStart:        ", heapStart)
		println("heapEnd:          ", heapEnd)
		println("total size:       ", totalSize)
		println("metadata size:    ", metadataSize)
		println("metadataStart:    ", metadataStart)
		println("# of blocks:      ", numBlocks)
		println("# of block states:", metadataSize*blocksPerStateByte)
	}
	if gcAsserts && metadataSize*blocksPerStateByte < numBlocks {
		// sanity check
		runtimePanic("gc: metadata array is too small")
	}
}

// alloc tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
//go:noinline
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
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
				// Try to increase heap size.
				if growHeap() {
					// Success, the heap was increased in size. Try again with a
					// larger heap.
				} else {
					// Unfortunately the heap could not be increased. This
					// happens on baremetal systems for example (where all
					// available RAM has already been dedicated to the heap).
					runtimePanic("out of memory")
				}
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

func realloc(ptr unsafe.Pointer, size uintptr) unsafe.Pointer {
	if ptr == nil {
		return alloc(size, nil)
	}

	ptrAddress := uintptr(ptr)
	endOfTailAddress := blockFromAddr(ptrAddress).findNext().address()

	// this might be a few bytes longer than the original size of
	// ptr, because we align to full blocks of size bytesPerBlock
	oldSize := endOfTailAddress - ptrAddress
	if size <= oldSize {
		return ptr
	}

	newAlloc := alloc(size, nil)
	memcpy(newAlloc, ptr, oldSize)
	free(ptr)

	return newAlloc
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
	markStack()
	markGlobals()

	if baremetal && hasScheduler {
		// Channel operations in interrupts may move task pointers around while we are marking.
		// Therefore we need to scan the runqueue seperately.
		var markedTaskQueue task.Queue
	runqueueScan:
		for !runqueue.Empty() {
			// Pop the next task off of the runqueue.
			t := runqueue.Pop()

			// Mark the task if it has not already been marked.
			markRoot(uintptr(unsafe.Pointer(&runqueue)), uintptr(unsafe.Pointer(t)))

			// Push the task onto our temporary queue.
			markedTaskQueue.Push(t)
		}

		finishMark()

		// Restore the runqueue.
		i := interrupt.Disable()
		if !runqueue.Empty() {
			// Something new came in while finishing the mark.
			interrupt.Restore(i)
			goto runqueueScan
		}
		runqueue = markedTaskQueue
		interrupt.Restore(i)
	} else {
		finishMark()
	}

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
		if start%unsafe.Alignof(start) != 0 {
			runtimePanic("gc: unaligned start pointer")
		}
		if end%unsafe.Alignof(end) != 0 {
			runtimePanic("gc: unaligned end pointer")
		}
	}

	// Reduce the end bound to avoid reading too far on platforms where pointer alignment is smaller than pointer size.
	// If the size of the range is 0, then end will be slightly below start after this.
	end -= unsafe.Sizeof(end) - unsafe.Alignof(end)

	for addr := start; addr < end; addr += unsafe.Alignof(addr) {
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
	return ptr >= heapStart && ptr < uintptr(metadataStart)
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
