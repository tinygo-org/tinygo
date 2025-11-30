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
	freeRanges    *freeRange     // freeRanges is a linked list of free block ranges
	endBlock      gcBlock        // the block just past the end of the available space
	gcTotalAlloc  uint64         // total number of bytes allocated
	gcMallocs     uint64         // total number of allocations
	gcLock        task.PMutex    // lock to avoid race conditions on multicore systems
)

// zeroSizedAlloc is just a sentinel that gets returned when allocating 0 bytes.
var zeroSizedAlloc uint8

// Provide some abstraction over heap blocks.

// blockState stores the four states in which a block can be.
// It holds 1 bit in each nibble.
// When stored into a state byte, each bit in a nibble corresponds to a different block.
// For blocks A-D, a state byte would be laid out as 0bDCBA_DCBA.
type blockState uint8

const (
	blockStateLow  blockState = 1
	blockStateHigh blockState = 1 << blocksPerStateByte

	blockStateFree blockState = 0
	blockStateHead blockState = blockStateLow
	blockStateTail blockState = blockStateHigh
	blockStateMark blockState = blockStateLow | blockStateHigh
	blockStateMask blockState = blockStateLow | blockStateHigh
)

// blockStateEach is a mask that can be used to extract a nibble from the block state.
const blockStateEach = 1<<blocksPerStateByte - 1

// The byte value of a block where every block is a 'tail' block.
const blockStateByteAllTails = byte(blockStateTail) * blockStateEach

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
	return blockState(stateByte>>(b%blocksPerStateByte)) & blockStateMask
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
	*stateBytePtr |= uint8(newState << (b % blocksPerStateByte))
	if gcAsserts && b.state() != newState {
		runtimePanic("gc: setState() was not successful")
	}
}

// objHeader is a structure prepended to every heap object to hold metadata.
type objHeader struct {
	// next is the next object to scan after this.
	next *objHeader

	// layout holds the layout bitmap used to find pointers in the object.
	layout gcLayout
}

// freeRange is a node on the outer list of range lengths.
// The free ranges are structured as two nested singly-linked lists:
// - The outer level (freeRange) has one entry for each unique range length.
// - The inner level (freeRangeMore) has one entry for each additional range of the same length.
// This two-level structure ensures that insertion/removal times are proportional to the requested length.
type freeRange struct {
	// len is the length of this free range.
	len uintptr

	// nextLen is the next longer free range.
	nextLen *freeRange

	// nextWithLen is the next free range with this length.
	nextWithLen *freeRangeMore
}

// freeRangeMore is a node on the inner list of equal-length ranges.
type freeRangeMore struct {
	next *freeRangeMore
}

// insertFreeRange inserts a range of len blocks starting at ptr into the free list.
func insertFreeRange(ptr unsafe.Pointer, len uintptr) {
	if gcAsserts && len == 0 {
		runtimePanic("gc: insert 0-length free range")
	}

	// Find the insertion point by length.
	// Skip until the next range is at least the target length.
	insDst := &freeRanges
	for *insDst != nil && (*insDst).len < len {
		insDst = &(*insDst).nextLen
	}

	// Create the new free range.
	next := *insDst
	if next != nil && next.len == len {
		// Insert into the list with this length.
		newRange := (*freeRangeMore)(ptr)
		newRange.next = next.nextWithLen
		next.nextWithLen = newRange
	} else {
		// Insert into the list of lengths.
		newRange := (*freeRange)(ptr)
		*newRange = freeRange{
			len:         len,
			nextLen:     next,
			nextWithLen: nil,
		}
		*insDst = newRange
	}
}

// popFreeRange removes a range of len blocks from the freeRanges list.
// It returns nil if there are no sufficiently long ranges.
func popFreeRange(len uintptr) unsafe.Pointer {
	if gcAsserts && len == 0 {
		runtimePanic("gc: pop 0-length free range")
	}

	// Find the removal point by length.
	// Skip until the next range is at least the target length.
	remDst := &freeRanges
	for *remDst != nil && (*remDst).len < len {
		remDst = &(*remDst).nextLen
	}

	rangeWithLength := *remDst
	if rangeWithLength == nil {
		// No ranges are long enough.
		return nil
	}
	removedLen := rangeWithLength.len

	// Remove the range.
	var ptr unsafe.Pointer
	if nextWithLen := rangeWithLength.nextWithLen; nextWithLen != nil {
		// Remove from the list with this length.
		rangeWithLength.nextWithLen = nextWithLen.next
		ptr = unsafe.Pointer(nextWithLen)
	} else {
		// Remove from the list of lengths.
		*remDst = rangeWithLength.nextLen
		ptr = unsafe.Pointer(rangeWithLength)
	}

	if removedLen > len {
		// Insert the leftover range.
		insertFreeRange(unsafe.Add(ptr, len*bytesPerBlock), removedLen-len)
	}
	return ptr
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

	// Rebuild the free ranges list.
	buildFreeRanges()
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

	// Rebuild the free ranges list.
	buildFreeRanges()
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

	// Acquire a range of free blocks.
	var ranGC bool
	var grewHeap bool
	var pointer unsafe.Pointer
	for {
		pointer = popFreeRange(neededBlocks)
		if pointer != nil {
			break
		}

		if !ranGC {
			// Run the collector and try again.
			freeBytes := runGC()
			ranGC = true
			heapSize := uintptr(metadataStart) - heapStart
			if freeBytes < heapSize/3 {
				// Ensure there is at least 33% headroom.
				// This percentage was arbitrarily chosen, and may need to
				// be tuned in the future.
				growHeap()
			}
			continue
		}

		if gcDebug && !grewHeap {
			println("grow heap for request:", uint(neededBlocks))
			dumpFreeRangeCounts()
		}
		if growHeap() {
			grewHeap = true
			continue
		}

		// Unfortunately the heap could not be increased. This
		// happens on baremetal systems for example (where all
		// available RAM has already been dedicated to the heap).
		runtimePanicAt(returnAddress(0), "out of memory")
	}

	// Set the backing blocks as being allocated.
	block := blockFromAddr(uintptr(pointer))
	block.setState(blockStateHead)
	for i := block + 1; i != block+gcBlock(neededBlocks); i++ {
		i.setState(blockStateTail)
	}

	// Create the object header.
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
	sweep()

	// Rebuild the free ranges list.
	freeBytes = buildFreeRanges()

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
func sweep() {
	metadataEnd := unsafe.Add(metadataStart, (endBlock+(blocksPerStateByte-1))/blocksPerStateByte)
	var carry byte
	for meta := metadataStart; meta != metadataEnd; meta = unsafe.Add(meta, 1) {
		// Fetch the state byte.
		stateBytePtr := (*byte)(unsafe.Pointer(meta))
		stateByte := *stateBytePtr

		// Separate blocks by type.
		// Split the nibbles.
		// Each nibble is a mask of blocks.
		high := stateByte >> blocksPerStateByte
		low := stateByte & blockStateEach
		// Marked heads are in both nibbles.
		markedHeads := low & high
		// Unmarked heads are in the low nibble but not the high nibble.
		unmarkedHeads := low &^ high
		// Tails are in the high nibble but not the low nibble.
		tails := high &^ low

		// Clear all tail runs after unmarked (freed) heads.
		//
		// Adding 1 to the start of a bit run will clear the run and set the next bit:
		//   (2^k - 1) + 1 = 2^k
		//   e.g. 0b0011 + 1 = 0b0100
		// Bitwise-and with the original mask to clear the newly set bit.
		//   e.g. (0b0011 + 1) & 0b0011 = 0b0100 & 0b0011 = 0b0000
		// This will not clear bits after the run because the gap stops the carry:
		//   e.g. (0b1011 + 1) & 0b1011 = 0b1100 & 0b1011 = 0b1000
		// This can clear multiple runs in a single addition:
		//   e.g. (0b1101 + 0b0101) & 0b1101 = 0b10010 & 0b1101 = 0b0000
		//
		// In order to find tail run starts after unmarked heads we could use tails & (unmarkedHeads << 1).
		// It is possible omit the bitwise-and because the clear still works if the next block is not a tail.
		// A head is not a tail, so corresponding missing tail bit will stop the carry from a previous tail run.
		// As such it will set the next bit which will be cleared back away later.
		// e.g. HHTH: (0b0010 + (0b1101 << 1)) & 0b0010 = 0b11100 & 0b0010 = 0b0000
		//
		// Treat the whole heap as a single pair of integer masks.
		// This is accomplished for addition by carrying the overflow to the next state byte.
		// The unmarkedHeads << 1 is equivalent to unmarkedHeads + unmarkedHeads, so it can be merged with the sum.
		// This does not require any special work for the bitwise-and because it operates bitwise.
		tailClear := tails + (unmarkedHeads << 1) + carry
		carry = tailClear >> blocksPerStateByte
		tails &= tailClear

		// Construct the new state byte.
		*stateBytePtr = markedHeads | (tails << blocksPerStateByte)
	}
}

// buildFreeRanges rebuilds the freeRanges list.
// This must be called after a GC sweep or heap grow.
// It returns how many bytes are free in the heap.
func buildFreeRanges() uintptr {
	freeRanges = nil
	block := endBlock
	var totalBlocks uintptr
	for {
		// Skip backwards over occupied blocks.
		for block > 0 && (block-1).state() != blockStateFree {
			block--
		}
		if block == 0 {
			break
		}

		// Find the start of the free range.
		end := block
		for block > 0 && (block-1).state() == blockStateFree {
			block--
		}

		// Insert the free range.
		len := uintptr(end - block)
		totalBlocks += len
		insertFreeRange(block.pointer(), len)
	}

	if gcDebug {
		println("free ranges after rebuild:")
		dumpFreeRangeCounts()
	}

	return totalBlocks * bytesPerBlock
}

func dumpFreeRangeCounts() {
	for rangeWithLength := freeRanges; rangeWithLength != nil; rangeWithLength = rangeWithLength.nextLen {
		totalRanges := uintptr(1)
		for nextWithLen := rangeWithLength.nextWithLen; nextWithLen != nil; nextWithLen = nextWithLen.next {
			totalRanges++
		}
		println("-", uint(rangeWithLength.len), "x", uint(totalRanges))
	}
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

	// Calculate the raw size of the heap.
	heapEnd := heapEnd
	heapStart := heapStart
	m.Sys = uint64(heapEnd - heapStart)
	m.HeapSys = uint64(uintptr(metadataStart) - heapStart)
	metadataStart := metadataStart
	// TODO: should GCSys include objHeaders?
	m.GCSys = uint64(heapEnd - uintptr(metadataStart))
	m.HeapReleased = 0 // always 0, we don't currently release memory back to the OS.

	// Count live heads and tails.
	var liveHeads, liveTails uintptr
	endBlock := endBlock
	metadataEnd := unsafe.Add(metadataStart, (endBlock+(blocksPerStateByte-1))/blocksPerStateByte)
	for meta := metadataStart; meta != metadataEnd; meta = unsafe.Add(meta, 1) {
		// Since we are outside of a GC, nothing is marked.
		// A bit in the low nibble implies a head.
		// A bit in the high nibble implies a tail.
		stateByte := *(*byte)(unsafe.Pointer(meta))
		liveHeads += uintptr(count4LUT[stateByte&blockStateEach])
		liveTails += uintptr(count4LUT[stateByte>>blocksPerStateByte])
	}

	// Add heads and tails to count live blocks.
	liveBlocks := liveHeads + liveTails
	liveBytes := uint64(liveBlocks * bytesPerBlock)
	m.HeapInuse = liveBytes
	m.HeapAlloc = liveBytes
	m.Alloc = liveBytes

	// Subtract live blocks from total blocks to count free blocks.
	freeBlocks := uintptr(endBlock) - liveBlocks
	m.HeapIdle = uint64(freeBlocks * bytesPerBlock)

	// Record the number of allocated objects.
	gcMallocs := gcMallocs
	m.Mallocs = gcMallocs

	// Subtract live objects from allocated objects to count freed objects.
	m.Frees = gcMallocs - uint64(liveHeads)

	// Record the total allocated bytes.
	m.TotalAlloc = gcTotalAlloc

	gcLock.Unlock()
}

// count4LUT is a lookup table used to count set bits in a 4-bit mask.
// TODO: replace with popcnt when available
var count4LUT = [16]uint8{
	0b0000: 0,
	0b0001: 1,
	0b0010: 1,
	0b0011: 2,
	0b0100: 1,
	0b0101: 2,
	0b0110: 2,
	0b0111: 3,
	0b1000: 1,
	0b1001: 2,
	0b1010: 2,
	0b1011: 3,
	0b1100: 2,
	0b1101: 3,
	0b1110: 3,
	0b1111: 4,
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}
