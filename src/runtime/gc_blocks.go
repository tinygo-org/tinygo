//go:build gc.conservative || gc.precise

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
// https://aykevl.nl/2020/09/gc-tinygo
// https://github.com/micropython/micropython/wiki/Memory-Manager
// https://github.com/micropython/micropython/blob/master/py/gc.c
// "The Garbage Collection Handbook" by Richard Jones, Antony Hosking, Eliot
// Moss.

import (
	"internal/task"
	"runtime/interrupt"
	"unsafe"
)

const gcDebug = false
const needsStaticHeap = true

// Some globals + constants for the entire GC.

const (
	wordsPerBlock      = 4 // number of pointers in an allocated block
	bytesPerBlock      = wordsPerBlock * unsafe.Sizeof(heapStart)
	stateBits          = 2 // how many bits a block state takes (see blockState type)
	blocksPerStateByte = 8 / stateBits
)

var (
	metadataStart unsafe.Pointer // pointer to the start of the heap metadata
	scanList      *objHeader     // scanList is a singly linked list of heap objects that have been marked but not scanned
	nextAlloc     gcBlock        // the next block that should be tried by the allocator
	endBlock      gcBlock        // the block just past the end of the available space
	gcTotalAlloc  uint64         // total number of bytes allocated
	gcTotalBlocks uint64         // total number of allocated blocks
	gcMallocs     uint64         // total number of allocations
	gcFrees       uint64         // total number of objects freed
	gcFreedBlocks uint64         // total number of freed blocks
	gcLock        task.PMutex    // lock to avoid race conditions on multicore systems
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

// The byte value of a block where every block is a 'tail' block.
const blockStateByteAllTails = 0 |
	uint8(blockStateTail<<(stateBits*3)) |
	uint8(blockStateTail<<(stateBits*2)) |
	uint8(blockStateTail<<(stateBits*1)) |
	uint8(blockStateTail<<(stateBits*0))

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
	addr := heapStart + uintptr(b)*bytesPerBlock
	if gcAsserts && addr > uintptr(metadataStart) {
		runtimePanic("gc: block pointing inside metadata")
	}
	return addr
}

// findHead returns the head (first block) of an object, assuming the block
// points to an allocated object. It returns the same block if this block
// already points to the head.
func (b gcBlock) findHead() gcBlock {
	for {
		// Optimization: check whether the current block state byte (which
		// contains the state of multiple blocks) is composed entirely of tail
		// blocks. If so, we can skip back to the last block in the previous
		// state byte.
		// This optimization speeds up findHead for pointers that point into a
		// large allocation.
		stateByte := b.stateByte()
		if stateByte == blockStateByteAllTails {
			b -= (b % blocksPerStateByte) + 1
			continue
		}

		// Check whether we've found a non-tail block, which means we found the
		// head.
		state := b.stateFromByte(stateByte)
		if state != blockStateTail {
			break
		}
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
	for b.address() < uintptr(metadataStart) && b.state() == blockStateTail {
		b++
	}
	return b
}

func (b gcBlock) stateByte() byte {
	return *(*uint8)(unsafe.Add(metadataStart, b/blocksPerStateByte))
}

// Return the block state given a state byte. The state byte must have been
// obtained using b.stateByte(), otherwise the result is incorrect.
func (b gcBlock) stateFromByte(stateByte byte) blockState {
	return blockState(stateByte>>((b%blocksPerStateByte)*stateBits)) & blockStateMask
}

// State returns the current block state.
func (b gcBlock) state() blockState {
	return b.stateFromByte(b.stateByte())
}

// setState sets the current block to the given state, which must contain more
// bits than the current state. Allowed transitions: from free to any state and
// from head to mark.
func (b gcBlock) setState(newState blockState) {
	stateBytePtr := (*uint8)(unsafe.Add(metadataStart, b/blocksPerStateByte))
	*stateBytePtr |= uint8(newState << ((b % blocksPerStateByte) * stateBits))
	if gcAsserts && b.state() != newState {
		runtimePanic("gc: setState() was not successful")
	}
}

// markFree sets the block state to free, no matter what state it was in before.
func (b gcBlock) markFree() {
	stateBytePtr := (*uint8)(unsafe.Add(metadataStart, b/blocksPerStateByte))
	*stateBytePtr &^= uint8(blockStateMask << ((b % blocksPerStateByte) * stateBits))
	if gcAsserts && b.state() != blockStateFree {
		runtimePanic("gc: markFree() was not successful")
	}
	if gcAsserts {
		*(*[wordsPerBlock]uintptr)(unsafe.Pointer(b.address())) = [wordsPerBlock]uintptr{}
	}
}

// unmark changes the state of the block from mark to head. It must be marked
// before calling this function.
func (b gcBlock) unmark() {
	if gcAsserts && b.state() != blockStateMark {
		runtimePanic("gc: unmark() on a block that is not marked")
	}
	clearMask := blockStateMask ^ blockStateHead // the bits to clear from the state
	stateBytePtr := (*uint8)(unsafe.Add(metadataStart, b/blocksPerStateByte))
	*stateBytePtr &^= uint8(clearMask << ((b % blocksPerStateByte) * stateBits))
	if gcAsserts && b.state() != blockStateHead {
		runtimePanic("gc: unmark() was not successful")
	}
}

// objHeader is a structure prepended to every heap object to hold metadata.
type objHeader struct {
	// next is the next object to scan after this.
	next *objHeader

	// layout holds the layout bitmap used to find pointers in the object.
	layout gcLayout
}

func isOnHeap(ptr uintptr) bool {
	return ptr >= heapStart && ptr < uintptr(metadataStart)
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
		runtimePanic("gc: setHeapEnd didn't grow the heap")
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
		runtimePanic("gc: heap did not grow enough at once")
	}
}

// calculateHeapAddresses initializes variables such as metadataStart and
// numBlock based on heapStart and heapEnd.
//
// This function can be called again when the heap size increases. The caller is
// responsible for copying the metadata to the new location.
func calculateHeapAddresses() {
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
//
//go:noinline
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	if size == 0 {
		return unsafe.Pointer(&zeroSizedAlloc)
	}

	if interrupt.In() {
		runtimePanicAt(returnAddress(0), "heap alloc in interrupt")
	}

	// Round the size up to a multiple of blocks, adding space for the header.
	rawSize := size
	size += align(unsafe.Sizeof(objHeader{}))
	size += bytesPerBlock - 1
	if size < rawSize {
		// The size overflowed.
		runtimePanicAt(returnAddress(0), "out of memory")
	}
	neededBlocks := size / bytesPerBlock
	size = neededBlocks * bytesPerBlock

	// Make sure there are no concurrent allocations. The heap is not currently
	// designed for concurrent alloc/GC.
	gcLock.Lock()

	// Update the total allocation counters.
	gcTotalAlloc += uint64(rawSize)
	gcMallocs++
	gcTotalBlocks += uint64(neededBlocks)

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
				freeBytes := runGC()
				heapSize := uintptr(metadataStart) - heapStart
				if freeBytes < heapSize/3 {
					// Ensure there is at least 33% headroom.
					// This percentage was arbitrarily chosen, and may need to
					// be tuned in the future.
					growHeap()
				}
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
					runtimePanicAt(returnAddress(0), "out of memory")
				}
			}
		}

		// Wrap around the end of the heap.
		if index == endBlock {
			index = 0
			// Reset numFreeBlocks as allocations cannot wrap.
			numFreeBlocks = 0
			// In rare cases, the initial heap might be so small that there are
			// no blocks at all. In this case, it's better to jump back to the
			// start of the loop and try again, until the GC realizes there is
			// no memory and grows the heap.
			// This can sometimes happen on WebAssembly, where the initial heap
			// is created by whatever is left on the last memory page.
			continue
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

			// Create the object header.
			pointer := thisAlloc.pointer()
			header := (*objHeader)(pointer)
			header.layout = parseGCLayout(layout)

			// We've claimed this allocation, now we can unlock the heap.
			gcLock.Unlock()

			// Return a pointer to this allocation.
			add := align(unsafe.Sizeof(objHeader{}))
			pointer = unsafe.Add(pointer, add)
			size -= add
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
	gcLock.Lock()
	runGC()
	gcLock.Unlock()
}

// runGC performs a garbage collection cycle. It is the internal implementation
// of the runtime.GC() function. The difference is that it returns the number of
// free bytes in the heap after the GC is finished.
func runGC() (freeBytes uintptr) {
	if gcDebug {
		println("running collection cycle...")
	}

	// Mark phase: mark all reachable objects, recursively.
	gcMarkReachable()

	if baremetal && hasScheduler {
		// Channel operations in interrupts may move task pointers around while we are marking.
		// Therefore we need to scan the runqueue separately.
		var markedTaskQueue task.Queue
	runqueueScan:
		runqueue := schedulerRunQueue()
		for !runqueue.Empty() {
			// Pop the next task off of the runqueue.
			t := runqueue.Pop()

			// Mark the task if it has not already been marked.
			markRoot(uintptr(unsafe.Pointer(runqueue)), uintptr(unsafe.Pointer(t)))

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
		*runqueue = markedTaskQueue
		interrupt.Restore(i)
	} else {
		finishMark()
	}

	// If we're using threads, resume all other threads before starting the
	// sweep.
	gcResumeWorld()

	// Sweep phase: free all non-marked objects and unmark marked objects for
	// the next collection cycle.
	freeBytes = sweep()

	// Show how much has been sweeped, for debugging.
	if gcDebug {
		dumpHeap()
	}

	return
}

// markRoots reads all pointers from start to end (exclusive) and if they look
// like a heap pointer and are unmarked, marks them and scans that object as
// well (recursively). The starting address must be valid and aligned.
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
	}

	// Scan the range conservatively.
	scanConservative(start, end-start)
}

// scanConservative scans all possible pointer locations in a range and marks referenced heap allocations.
// The starting address must be valid and pointer-aligned.
func scanConservative(addr, len uintptr) {
	for len >= unsafe.Sizeof(addr) {
		root := *(*uintptr)(unsafe.Pointer(addr))
		markRoot(addr, root)

		addr += unsafe.Alignof(addr)
		len -= unsafe.Alignof(addr)
	}
}

func markCurrentGoroutineStack(sp uintptr) {
	// This could be optimized by only marking the stack area that's currently
	// in use.
	markRoot(0, sp)
}

// finishMark finishes the marking process by scanning all heap objects on scanList.
func finishMark() {
	for {
		// Remove an object from the scan list.
		obj := scanList
		if obj == nil {
			return
		}
		scanList = obj.next

		// Check if the object may contain pointers.
		if obj.layout.pointerFree() {
			// This object doesn't contain any pointers.
			// This is a fast path for objects like make([]int, 4096).
			// It skips the length calculation.
			continue
		}

		// Compute the scan bounds.
		objAddr := uintptr(unsafe.Pointer(obj))
		start := objAddr + align(unsafe.Sizeof(objHeader{}))
		end := blockFromAddr(objAddr).findNext().address()

		// Scan the object.
		obj.layout.scan(start, end-start)
	}
}

// mark a GC root at the address addr.
func markRoot(addr, root uintptr) {
	// Find the heap block corresponding to the root.
	if !isOnHeap(root) {
		// This is not a heap pointer.
		return
	}
	block := blockFromAddr(root)

	// Find the head of the corresponding object.
	if block.state() == blockStateFree {
		// The to-be-marked object doesn't actually exist.
		// This could either be a dangling pointer (oops!) but most likely
		// just a false positive.
		return
	}
	head := block.findHead()

	// Mark the object.
	if head.state() == blockStateMark {
		// This object is already marked.
		return
	}
	if gcDebug {
		println("found unmarked pointer", root, "at address", addr)
	}
	head.setState(blockStateMark)

	// Add the object to the scan list.
	header := (*objHeader)(head.pointer())
	header.next = scanList
	scanList = header
}

// Sweep goes through all memory and frees unmarked memory.
// It returns how many bytes are free in the heap after the sweep.
func sweep() (freeBytes uintptr) {
	freeCurrentObject := false
	var freed uint64
	for block := gcBlock(0); block < endBlock; block++ {
		switch block.state() {
		case blockStateHead:
			// Unmarked head. Free it, including all tail blocks following it.
			block.markFree()
			freeCurrentObject = true
			gcFrees++
			freed++
		case blockStateTail:
			if freeCurrentObject {
				// This is a tail object following an unmarked head.
				// Free it now.
				block.markFree()
				freed++
			}
		case blockStateMark:
			// This is a marked object. The next tail blocks must not be freed,
			// but the mark bit must be removed so the next GC cycle will
			// collect this object if it is unreferenced then.
			block.unmark()
			freeCurrentObject = false
		case blockStateFree:
			freeBytes += bytesPerBlock
		}
	}
	gcFreedBlocks += freed
	freeBytes += uintptr(freed) * bytesPerBlock
	return
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

// ReadMemStats populates m with memory statistics.
//
// The returned memory statistics are up to date as of the
// call to ReadMemStats. This would not do GC implicitly for you.
func ReadMemStats(m *MemStats) {
	gcLock.Lock()
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
	m.TotalAlloc = gcTotalAlloc
	m.Mallocs = gcMallocs
	m.Frees = gcFrees
	m.Sys = uint64(heapEnd - heapStart)
	m.HeapAlloc = (gcTotalBlocks - gcFreedBlocks) * uint64(bytesPerBlock)
	m.Alloc = m.HeapAlloc
	gcLock.Unlock()
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}
