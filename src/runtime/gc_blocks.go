//go:build gc.conservative || gc.precise

package runtime

// The -gc=conservative and -gc=precise memory managers are conventional
// mark-sweep garbage collectors.
//
// This memory manager uses a single flat range of backing memory.
// The range is provided by the platform runtime through the heapStart
// (inclusive) and heapEnd (exclusive) variables.
//
// This range is subdivided by calculateHeapAddresses into 3 regions:
//  - The blocks array at [heapStart, endBlocksBitmap)
//  - The end blocks bitmap at [endBlocksBitmap, visitedBlocksBitmap)
//  - The visited blocks bitmap at [visitedBlocksBitmap, visitedBlocksBitmap+bitmapSize)
// The leftover memory after the visited blocks bitmap is unused.
//
// The blocks array is the region that memory is allocated in. It it is divided
// into blocks of 4 pointer widths (see bytesPerBlock). This size is always a
// multiple of the maximum required alignment, so each block is always
// appropriately aligned.
//
// During normal allocation, the memory manager maintains a list of free block
// ranges (see freeRanges). It removes the shortest range that is long enough
// to hold the requested heap object. If the removed range is longer than
// requested, it reinserts the leftover blocks into the list. The last block of
// the heap object range is added to the end blocks bitmap, and an objHeader is
// placed at the end of it. This objHeader will later be used by the mark pass.
// When using -gc=precise, the type information data is placed within this
// header.
//
// If no sufficiently-long ranges are found then the mark pass begins. The
// visited blocks bitmap is first cleared. The ending blocks of all remaining
// free ranges are added to both bitmaps. Next all stacks and globals are
// scanned for pointers, which are then "marked". The process of marking a
// pointer consists of:
//  1. Find the index of the block containing the address. Addresses outside
//     the blocks array are ignored.
//  2. Skip forwards until the next block that is in either the visited or ends
//     bitmap, adding them to the visited bitmap as we go. If we encounter an
//     already-visited block, then the pointer is to an already-marked object
//     or free range.
//  3. Add the objHeader in the block to the scanList to process later.
// Next, we loop through the scanList to mark the contents of all visited objects.
//
// After the marking pass is done, all live objects have been added to the
// visited blocks bitmap. The sweep pass begins by removing the previously-free
// ranges from the visited bitmap. It then seperates visited and unvisited
// object ends into two bitmaps:
//  - The visited objects are in use, and thus stay in the end blocks bitmap.
//  - The unvisited objects are now free. The memory of the visited blocks
//    bitmap is reused to track free ends.
// Finally, buildFreeRanges rebuilds the free ranges list based on these two
// bitmaps. At this point, the GC is done and the allocator can repeat the
// search for a usable free range.
//
// If there are still no free ranges, then it attempts to grow the heap's
// backing memory range. On hosted targets (Linux/WASM/etc.), this may extend
// the virtual memory used by the heap. If this is possible, setHeapEnd moves
// the ends bitmap and updates the free list. The allocator can repeat the
// search for a usable free range with the new list.
//
// If the heap cannot be grown enough to satisfy the request, we finally give
// up and panic with an "out of memory" message.

import (
	"internal/task"
	"math/bits"
	"runtime/interrupt"
	"unsafe"
)

const gcDebug = false
const gcTiming = false
const sweepMetrics = false
const needsStaticHeap = true

// Some globals + constants for the entire GC.

const (
	// wordsPerBlock is the number of pointers that can fit into a block without overlapping.
	wordsPerBlock = 4

	// bytesPerBlock is the size of a heap block in bytes.
	bytesPerBlock = wordsPerBlock * unsafe.Sizeof(uintptr(0))

	// maskSizeBytes is the size of the gcMask type in bytes.
	maskSizeBytes = unsafe.Sizeof(gcMask(0))

	// maskSizeBits is the size of the gcMask type in bits.
	maskSizeBits = 8 * maskSizeBytes
)

var (
	// endBlocksBitmap is the base address of the end blocks bitmap.
	// The last block in a heap object or free range is considered an end block.
	endBlocksBitmap uintptr

	// visitedBlocksBitmap is the base address of the visited blocks bitmap.
	// markRoot "visits" blocks from the marked address to the next end block.
	// It may stop early if it finds an already-visited block.
	visitedBlocksBitmap uintptr

	// blocks is the heap size in blocks.
	blocks uintptr

	// scanList is a singly linked list of heap objects that have been marked but not scanned.
	scanList *objHeader

	// freeRanges is a linked list of free block ranges.
	freeRanges *freeRange

	// gcTotalAlloc is the total number of bytes allocated since heap initialization.
	// This is used by ReadMemStats.
	gcTotalAlloc uint64

	// gcMallocs is the total number of allocations since heap initialization.
	// This is used by ReadMemStats.
	gcMallocs uint64

	// gcLock is used to control access to the GC on multicore systems.
	// The GC is not otherwise thread-safe.
	gcLock task.PMutex
)

// objHeader is a structure appended to every heap object to hold metadata.
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
	freeRangeMore

	// nextLen is the next longer free range.
	nextLen *freeRange

	// len is the length of this free range.
	len uintptr
}

// freeRangeMore is a node on the inner list of equal-length ranges.
type freeRangeMore struct {
	// nextWithLen is the next free range with the same length.
	nextWithLen *freeRangeMore
}

// insertFreeRange inserts a range of len blocks ending at endAddr into the free list.
//
//go:nobounds
func insertFreeRange(endAddr uintptr, len uintptr) {
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
		newRange := (*freeRangeMore)(unsafe.Pointer(endAddr - unsafe.Sizeof(freeRangeMore{})))
		newRange.nextWithLen = next.nextWithLen
		next.nextWithLen = newRange
	} else {
		// Insert into the list of lengths.
		newRange := (*freeRange)(unsafe.Pointer(endAddr - unsafe.Sizeof(freeRange{})))
		*newRange = freeRange{
			len:     len,
			nextLen: next,
		}
		*insDst = newRange
	}
}

// Initialize the memory allocator.
// No memory may be allocated before this is called. That means the runtime and
// any packages the runtime depends upon may not allocate memory during package
// initialization.
//
//go:nobounds
func initHeap() {
	calculateHeapAddresses()

	// Initialize the ends bitmap.
	endBlocksBitmap := endBlocksBitmap
	visitedBlocksBitmap := visitedBlocksBitmap
	bitmapSize := visitedBlocksBitmap - endBlocksBitmap
	if bitmapSize == 0 {
		// Empty heap.
		return
	}
	memzero(unsafe.Pointer(endBlocksBitmap), bitmapSize)

	// Insert the initial free range.
	r := (*freeRange)(unsafe.Pointer(endBlocksBitmap - unsafe.Sizeof(freeRange{})))
	freeRanges = r
	*r = freeRange{len: blocks}
}

// setHeapEnd is called to expand the heap. The heap can only grow, not shrink.
// Also, the heap should grow substantially each time otherwise growing the heap
// will be expensive.
func setHeapEnd(newHeapEnd uintptr) {
	if gcAsserts && newHeapEnd <= heapEnd {
		runtimePanic("gc: setHeapEnd didn't grow the heap")
	}

	// Save some old variables we need later.
	oldEndBlocksBitmap := endBlocksBitmap
	oldBitmapSize := visitedBlocksBitmap - endBlocksBitmap

	// Update the heap layout.
	heapEnd = newHeapEnd
	calculateHeapAddresses()

	// Move the old end blocks bitmap.
	endBlocksBitmap := endBlocksBitmap
	memmove(unsafe.Pointer(endBlocksBitmap), unsafe.Pointer(oldEndBlocksBitmap), oldBitmapSize)

	// Widen the bitmap.
	visitedBlocksBitmap := visitedBlocksBitmap
	newBitmapSize := visitedBlocksBitmap - endBlocksBitmap
	memzero(unsafe.Pointer(endBlocksBitmap+oldBitmapSize), newBitmapSize-oldBitmapSize)

	// Populate the visitedBlocksBitmap with free range ends (including the new free range).
	memzero(unsafe.Pointer(visitedBlocksBitmap), newBitmapSize-maskSizeBytes)
	*(*gcMask)(unsafe.Pointer(visitedBlocksBitmap + newBitmapSize - maskSizeBytes)) = 1 << ((blocks - 1) % maskSizeBits)
	toggleFree(visitedBlocksBitmap)

	// Rebuild the free ranges.
	buildFreeRanges()
}

// calculateHeapAddresses initializes the heap layout variables based on
// heapStart and heapEnd.
//
// This function can be called again when the heap size increases. The caller is
// responsible for copying the endBlockBitmap to the new location.
func calculateHeapAddresses() {
	totalSize := heapEnd - heapStart

	// Allocate some memory to keep 2 bits of information about every block.
	// Use the rest of the available memory as heap.
	const batchSize = maskSizeBits*bytesPerBlock + 2*maskSizeBytes
	bitmapSize := ((totalSize + batchSize - bytesPerBlock) / batchSize) * maskSizeBytes
	blocks = (totalSize - 2*bitmapSize) / bytesPerBlock
	endBlocksBitmap = heapStart + blocks*bytesPerBlock
	visitedBlocksBitmap = endBlocksBitmap + bitmapSize

	if gcDebug {
		println("heapStart:          ", heapStart)
		println("heapEnd:            ", heapEnd)
		println("total size:         ", totalSize)
		println("bitmap size:        ", bitmapSize)
		println("endBlocksBitmap:    ", endBlocksBitmap)
		println("visitedBlocksBitmap:", visitedBlocksBitmap)
		println("# of blocks:        ", blocks)
	}

	if gcAsserts {
		// sanity check
		if 8*bitmapSize < blocks {
			runtimePanic("gc: metadata array is too small")
		}
		if visitedBlocksBitmap+bitmapSize > heapEnd {
			runtimePanic("gc: heap bounds overrun")
		}
	}
}

// alloc tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
//
//go:noinline
//go:nobounds
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	if size == 0 {
		return alloc_zero(size, layout)
	}

	if interrupt.In() {
		runtimePanicAt(returnAddress(0), "heap alloc in interrupt")
	}

	// Round the size up to a multiple of blocks, adding space for the header.
	rawSize := size
	size += unsafe.Sizeof(objHeader{})
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
	var endAddr uintptr
	for {
		// Search the free ranges length list for neededBlocks.
		remDst := &freeRanges
		for *remDst != nil && (*remDst).len < neededBlocks {
			remDst = &(*remDst).nextLen
		}
		rangeWithLength := *remDst
		if rangeWithLength != nil {
			// We found a sufficiently-long range.
			removedLen := rangeWithLength.len

			// Remove the range.
			if nextWithLen := rangeWithLength.nextWithLen; nextWithLen != nil {
				// Remove from the list with this length.
				rangeWithLength.nextWithLen = nextWithLen.nextWithLen
				endAddr = uintptr(unsafe.Pointer(nextWithLen)) + unsafe.Sizeof(freeRangeMore{})
			} else {
				// Remove from the list of lengths.
				*remDst = rangeWithLength.nextLen
				endAddr = uintptr(unsafe.Pointer(rangeWithLength)) + unsafe.Sizeof(freeRange{})
			}

			if removedLen > neededBlocks {
				// Insert the leftover range.
				leftover := removedLen - neededBlocks
				insertFreeRange(endAddr, leftover)
				endAddr -= leftover * bytesPerBlock
			}

			break
		}

		if !ranGC {
			// Run the collector and try again.
			freeBytes := runGC()
			ranGC = true
			heapSize := endBlocksBitmap - heapStart
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

	// Add the new object to the ends bitmap.
	endBlock := ((endAddr - heapStart) / bytesPerBlock) - 1
	*(*gcMask)(unsafe.Pointer(endBlocksBitmap + maskSizeBytes*(endBlock/maskSizeBits))) |= 1 << (endBlock % maskSizeBits)

	// Create the object header.
	header := (*objHeader)(unsafe.Pointer(endAddr - unsafe.Sizeof(objHeader{})))
	header.layout = parseGCLayout(layout)

	// We've claimed this allocation, now we can unlock the heap.
	gcLock.Unlock()

	// Return a pointer to this allocation.
	pointer := unsafe.Pointer(endAddr - size)
	size -= unsafe.Sizeof(objHeader{})
	if gcDebug {
		println("alloc", pointer, "-", endAddr, "size:", size)
	}
	memzero(pointer, size)

	// Return a pointer to this allocation.
	return pointer
}

func realloc(ptr unsafe.Pointer, size uintptr) unsafe.Pointer {
	if ptr == nil {
		return alloc(size, nil)
	}

	gcLock.Lock()

	startBlock := (uintptr(ptr) - heapStart) / bytesPerBlock
	blocks := blocks
	endBlocksBitmap := endBlocksBitmap
	endBlock := startBlock
	for ; endBlock < blocks; endBlock++ {
		if *(*gcMask)(unsafe.Pointer(endBlocksBitmap + maskSizeBytes*(endBlock/maskSizeBits)))&(1<<(endBlock%maskSizeBits)) != 0 {
			break
		}
	}

	gcLock.Unlock()

	// this might be a few bytes longer than the original size of
	// ptr, because we align to full blocks of size bytesPerBlock
	oldSize := (endBlock-startBlock)*bytesPerBlock + (bytesPerBlock - unsafe.Sizeof(objHeader{}))
	if size <= oldSize {
		// The requested size is less than the old size.
		// There are likely scenarios for this:
		//  - The caller intended to grow the allocation, but the original size
		//    was rounded up by alloc to a multiple of the block size.
		//    The rounded size is already sufficient.
		//  - The caller intended to shrink the allocation.
		//    We currently ignore this case.
		// Either way, the current allocation can be left alone.
		return ptr
	}

	// Create a new allocation and copy the old data.
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

	var gcStart timeUnit
	if gcTiming {
		gcStart = ticks()
	}

	// Clear the visited bitmap.
	memzero(unsafe.Pointer(visitedBlocksBitmap), visitedBlocksBitmap-endBlocksBitmap)

	// Add the free ranges as visited ends.
	// This will prevent marking of addresses to within them.
	toggleFree(endBlocksBitmap)
	toggleFree(visitedBlocksBitmap)

	var gcPrepEnd timeUnit
	if gcTiming {
		gcPrepEnd = ticks()
	}

	// Mark phase: mark all reachable objects, recursively.
	gcMarkReachable()

	var gcPreMarkEnd timeUnit
	if gcTiming {
		gcPreMarkEnd = ticks()
	}

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

	var gcPostMarkEnd timeUnit
	if gcTiming {
		gcPostMarkEnd = ticks()
	}

	// If we're using threads, resume all other threads before starting the
	// sweep.
	gcResumeWorld()

	var gcCleanupEnd timeUnit
	if gcTiming {
		gcCleanupEnd = ticks()
	}

	// Unmark the free range ends.
	toggleFree(visitedBlocksBitmap)

	// Split the ends into two bitmaps: one with visited ends and one with unvisited ends.
	{
		endBlocksBitmap := endBlocksBitmap
		visitedBlocksBitmap := visitedBlocksBitmap
		for i := visitedBlocksBitmap - endBlocksBitmap; i > 0; {
			i -= maskSizeBytes
			endsPtr := (*gcMask)(unsafe.Pointer(endBlocksBitmap + i))
			visitedPtr := (*gcMask)(unsafe.Pointer(visitedBlocksBitmap + i))
			ends := *endsPtr
			visited := *visitedPtr
			*endsPtr = ends & visited
			*visitedPtr = ends &^ visited
		}
	}

	// Rebuild the free ranges based on these bitmaps.
	freeBytes = buildFreeRanges()

	var gcSweepEnd timeUnit
	if gcTiming {
		gcSweepEnd = ticks()
	}

	if gcTiming {
		println("gc timing:", ticksToNanoseconds(gcSweepEnd-gcStart), "ns")
		println("\tprep:     ", ticksToNanoseconds(gcPrepEnd-gcStart), "ns")
		println("\tpre-mark: ", ticksToNanoseconds(gcPreMarkEnd-gcPrepEnd), "ns")
		println("\tpost-mark:", ticksToNanoseconds(gcPostMarkEnd-gcPreMarkEnd), "ns")
		println("\tcleanup:  ", ticksToNanoseconds(gcCleanupEnd-gcPostMarkEnd), "ns")
		println("\tsweep:    ", ticksToNanoseconds(gcSweepEnd-gcCleanupEnd), "ns")
	}

	if gcDebug {
		println("free ranges after gc:")
		dumpFreeRangeCounts()
	}

	return
}

// toggleFree toggles the ends of free ranges in the provided bitmap.
//
//go:nobounds
func toggleFree(base uintptr) {
	heapStart := heapStart
	for rangeWithLength := freeRanges; rangeWithLength != nil; {
		r := &rangeWithLength.freeRangeMore
		rangeWithLength = rangeWithLength.nextLen
		for {
			block := (uintptr(unsafe.Pointer(r)) - heapStart) / bytesPerBlock
			*(*gcMask)(unsafe.Pointer(base + maskSizeBytes*(block/maskSizeBits))) ^= 1 << (block % maskSizeBits)
			r = r.nextWithLen
			if r == nil {
				break
			}
		}
	}
}

// markRoots reads all pointers from start to end (exclusive) and if they look
// like a heap pointer and are unmarked, marks them and adds them to the
// scanList. The starting address must be valid and aligned.
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

// scanConservative scans all possible pointer locations in a range and marks
// referenced heap allocations. The starting address must be valid and
// pointer-aligned.
//
//go:nobounds
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

// mark a GC root at the address addr. If root is an address within an umarked
// heap object, this adds the object to the scanList.
//
//go:nobounds
func markRoot(addr, root uintptr) {
	// Find the corresponding heap block index.
	heapStart := heapStart
	block := (root - heapStart) / bytesPerBlock
	if block >= blocks {
		// This is not on the heap.
		return
	}

	// Visit blocks until we reach an end.
	endBlocksBitmap := endBlocksBitmap
	visitedBlocksBitmap := visitedBlocksBitmap
	for {
		// Split the bitmap position into a word and a bit.
		wordIdx := block / maskSizeBits
		bit := gcMask(1) << (block % maskSizeBits)

		// Subtracting the selected bit from the ends mask will clear the end and set all bits inbetween.
		// We can xor with the original ends mask to get an inclusive range of blocks.
		ends := *(*gcMask)(unsafe.Pointer(endBlocksBitmap + wordIdx*maskSizeBytes))
		newVisit := (ends - bit) ^ ends

		// Add these bits to the visited mask.
		visitedPtr := (*gcMask)(unsafe.Pointer(visitedBlocksBitmap + wordIdx*maskSizeBytes))
		oldVisit := *visitedPtr
		*visitedPtr = oldVisit | newVisit
		if oldVisit&newVisit != 0 {
			// We reached a block that has already been visited.
			// This markRoot is redundant.
			if gcDebug {
				println("root already visited", root, "from", addr)
			}
			return
		}

		if newVisit&ends != 0 {
			// We reached an unvisited end.
			// Compute the final block index.
			if hasFastCLZ {
				block &^= maskSizeBits - 1
				// NOTE: LLVM can narrow this to the appropriate type.
				block += 63 - uintptr(bits.LeadingZeros64(uint64(newVisit)))
			} else {
				tmp := newVisit
				for {
					tmp >>= 1
					if tmp < bit {
						break
					}
					block++
				}
			}
			break
		}

		// Skip to the next bitmap word.
		block = (block | (maskSizeBits - 1)) + 1
	}

	if gcAsserts && *(*gcMask)(unsafe.Pointer(endBlocksBitmap + maskSizeBytes*(block/maskSizeBits)))&(1<<(block%maskSizeBits)) == 0 {
		runtimePanic("wrong end")
	}

	if gcDebug {
		println("mark root", root, "from", addr, "end", heapStart+block*bytesPerBlock+bytesPerBlock)
	}

	// Add the object to the scan list.
	hdr := (*objHeader)(unsafe.Pointer(heapStart + block*bytesPerBlock + (bytesPerBlock - unsafe.Sizeof(objHeader{}))))
	hdr.next = scanList
	scanList = hdr
}

// finishMark finishes the marking process by scanning all heap objects on scanList.
//
//go:nobounds
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
		end := uintptr(unsafe.Pointer(obj))
		heapStart := heapStart
		endBlock := (end - heapStart) / bytesPerBlock
		startBlock := gcBitmapScanBackwards(endBlocksBitmap, endBlock) + 1
		start := heapStart + startBlock*bytesPerBlock

		// Scan the object.
		obj.layout.scan(start, end-start)
	}
}

// buildFreeRanges discards and rebuilds the free ranges list. It expects the
// GC or setHeapEnd to first populate visitedBlocksBitmap with all free or dead
// range ends.
//
//go:nobounds
func buildFreeRanges() uintptr {
	// Clear the free ranges list.
	freeRanges = nil

	// Loop backwards over the heap to find free ranges.
	heapStart := heapStart
	var totalFreeBlocks uintptr
	var totalFreeRanges uintptr
	for block := blocks; ; {
		// Find the next free or dead end.
		groupEnd := gcBitmapScanBackwards(visitedBlocksBitmap, block)
		if groupEnd == ^uintptr(0) {
			// There is no empty space left in the heap.
			break
		}

		// Find the next live end.
		block = gcBitmapScanBackwards(endBlocksBitmap, groupEnd)

		// Add the range between these ends to the free list.
		groupBlocks := groupEnd - block
		totalFreeBlocks += groupBlocks
		if gcDebug {
			println("insert free range", heapStart+block*bytesPerBlock+bytesPerBlock, "-", heapStart+groupEnd*bytesPerBlock+bytesPerBlock, "blocks:", groupBlocks)
		}
		insertFreeRange(heapStart+groupEnd*bytesPerBlock+bytesPerBlock, groupBlocks)
		totalFreeRanges++

		if block == ^uintptr(0) {
			// The range reached the start of the heap.
			break
		}
	}

	if sweepMetrics {
		var sourceFrees uintptr
		for i := (visitedBlocksBitmap - endBlocksBitmap) / maskSizeBytes; i > 0; {
			i--
			mask := *(*gcMask)(unsafe.Pointer(visitedBlocksBitmap + maskSizeBytes*i))
			for mask != 0 {
				sourceFrees++
				mask &= mask - 1
			}
		}
		println("sweep metrics:")
		println("\tsource free ranges:  ", uint(sourceFrees))
		println("\tfree blocks:         ", uint(totalFreeBlocks))
		println("\tfree ranges:         ", uint(totalFreeRanges))
		println("\tavg blocks per range:", uint(totalFreeBlocks/totalFreeRanges))
		println("\tavg merged:          ", uint(sourceFrees/totalFreeRanges))
	}

	return totalFreeBlocks * bytesPerBlock
}

// gcBitmapScanBackwards finds the next index less than idx set in the provided
// bitmap. It returns ^uintptr(0) if no set bits are found.
//
//go:nobounds
func gcBitmapScanBackwards(base uintptr, idx uintptr) uintptr {
	// Select the next valid index.
	idx--
	if idx == ^uintptr(0) {
		// There are no more valid indices.
		return idx
	}

	// Select the word containing idx.
	// Shift off bits after idx.
	maskAddr := base + maskSizeBytes*(idx/maskSizeBits)
	mask := *(*gcMask)(unsafe.Pointer(maskAddr)) << ((maskSizeBits - 1) - (idx % maskSizeBits))
	if mask == 0 {
		// There were no more set bits in that word.
		// Skip backwards to find the next nonzero word.
		idx |= maskSizeBits - 1
		for {
			idx -= maskSizeBits
			if idx == ^uintptr(0) {
				return idx
			}
			maskAddr -= maskSizeBytes
			mask = *(*gcMask)(unsafe.Pointer(maskAddr))
			if mask != 0 {
				break
			}
		}
	}

	// The current idx is at the top bit of mask.
	// Move idx to the highest set bit in mask.
	if hasFastCLZ {
		// Subtract the leading zeroes from idx. When using bits.LeadingZeros64
		// on a wider type, we must compensate for the zeroes added by
		// zero-extending to uint64.
		// NOTE: LLVM can narrow this to the appropriate type.
		idx -= uintptr(bits.LeadingZeros64(uint64(mask))) - (64 - maskSizeBits)
	} else {
		// Shift mask up until the top bit is set.
		// Decrement the index every time we shift.
		for mask < 1<<(maskSizeBits-1) {
			mask <<= 1
			idx--
		}
	}

	return idx
}

// dumpFreeRangeCounts prints the distribution of range lengths in the current freeRanges list.
// This is useful for debugging memory fragmentation.
func dumpFreeRangeCounts() {
	for rangeWithLength := freeRanges; rangeWithLength != nil; rangeWithLength = rangeWithLength.nextLen {
		totalRanges := uintptr(1)
		for nextWithLen := rangeWithLength.nextWithLen; nextWithLen != nil; nextWithLen = nextWithLen.nextWithLen {
			totalRanges++
		}
		println("-", uint(rangeWithLength.len), "x", uint(totalRanges))
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
	endBlocksBitmap := endBlocksBitmap
	m.HeapSys = uint64(endBlocksBitmap - heapStart)
	// TODO: should GCSys include objHeaders?
	m.GCSys = uint64(heapEnd - endBlocksBitmap)
	m.HeapReleased = 0 // always 0, we don't currently release memory back to the OS.

	// Count live objects.
	var liveObjects uintptr
	for i := visitedBlocksBitmap - endBlocksBitmap; i > 0; {
		// Select the next mask.
		i -= maskSizeBytes
		mask := *(*gcMask)(unsafe.Pointer(endBlocksBitmap + i))

		// Add the bits in this mask to liveObjects.
		// NOTE: We could use bits.OnesCount* here on some platforms?
		for ; mask != 0; mask &= mask - 1 {
			liveObjects++
		}
	}
	m.HeapObjects = uint64(liveObjects)

	// Count free ranges and their contained space.
	var freeRangeCount uintptr
	var freeBlocks uintptr
	for rangeWithLength := freeRanges; rangeWithLength != nil; {
		len := rangeWithLength.len
		r := &rangeWithLength.freeRangeMore
		rangeWithLength = rangeWithLength.nextLen
		for {
			freeRangeCount++
			freeBlocks += len
			r = r.nextWithLen
			if r == nil {
				break
			}
		}
	}

	// Record the free space.
	m.HeapIdle = uint64(freeBlocks * bytesPerBlock)

	// Subtract free blocks from total blocks to count live blocks.
	blocks := blocks
	liveBlocks := blocks - freeBlocks
	liveBytes := uint64(liveBlocks * bytesPerBlock)
	m.HeapInuse = liveBytes
	m.HeapAlloc = liveBytes
	m.Alloc = liveBytes

	// Record the lifetime allocation count of the GC.
	gcMallocs := gcMallocs
	m.Mallocs = gcMallocs

	// Subtract live objects from allocated objects to count freed objects.
	m.Frees = gcMallocs - uint64(liveObjects)

	// Record the total allocated bytes.
	m.TotalAlloc = gcTotalAlloc

	gcLock.Unlock()
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}
