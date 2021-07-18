// +build gc.list

// This memory manager is partially inspired by the memory allocator included in avr-libc.
// Unlike avr-libc malloc, this memory manager stores an allocation list instead of a free list.
// There is an overhead of 2 pointer-widths/allocation (up from avr-libc's 1 pointer-width/allocation).
// The allocation list is stored in ascending address order.
// The set of free spans is implicitly derived from the gaps between allocations.
// This allocator has an effective allocation complexity of O(n) and worst-case GC complexity of O(n^2).
// Due to architectural quirks, as well as tiny heaps, this should almost always be faster than the standard conservative collector on AVR.

package runtime

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

// setHeapEnd is called to expand the heap. The heap can only grow, not shrink.
func setHeapEnd(newHeapEnd uintptr) {
	if gcAsserts && newHeapEnd <= heapEnd {
		panic("gc: setHeapEnd didn't grow the heap")
	}

	heapEnd = newHeapEnd

	// There is no other metadata to update since the free space is just computed from the heap bounds and the allocations list.
}

// Initialize the memory allocator.
func initHeap() {
	// No initialization is required: the only state is the allocation list, which starts empty.
	// This function must exist anyway since other GCs need initialization.
	// It should be inlined and optimized away by the compiler.
}

var gcRunning bool

// alloc tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
//go:noinline
func alloc(size uintptr) unsafe.Pointer {
	if gcAsserts && gcRunning {
		runtimePanic("alloc while GC is running")
	}

	if size == 0 {
		return unsafe.Pointer(&zeroSizedAlloc)
	}

	var ranGC bool
tryAlloc:
	// Search for available memory.
	alloc := findMem(size)
	if alloc == nil {
		// There is no free span large enough for the allocation.

		if ranGC {
			// Even after running the GC, there is not enough memory.
			if growHeap() {
				// The heap was able to grow to fit more data.
				// Try allocating again.
				// When doing a huge allocation this will loop until the heap is big enough.
				goto tryAlloc
			}

			// There is no more available memory.
			runtimePanic("out of memory")
		}

		// Run the garbage collector and try again.
		GC()
		ranGC = true
		goto tryAlloc
	}

	// Zero the allocation.
	ptr := unsafe.Pointer(uintptr(unsafe.Pointer(alloc)) + unsafe.Sizeof(allocHeader{}))
	memzero(ptr, size)

	return ptr
}

// zeroSizedAlloc is just a sentinel that gets returned when allocating 0 bytes.
var zeroSizedAlloc uint8

// findMem searches the heap for a free span large enough to contain an allocation of the specified size.
// If there are no sufficiently large free spans available, this returns nil.
func findMem(size uintptr) *allocHeader {
	// This memory allocator implementation is effectively the same algorithm applied by avr-libc.
	// It loops through the set of all free spans, and selects the smallest span that is large enough to fit the allocation.

	allocSize := unsafe.Sizeof(allocHeader{}) + size

	// best* store the best-fit free span.
	var bestDst **allocHeader
	var bestStart uintptr
	var bestSize uintptr

	start := heapStart
	dst := &activeAllocs.first
searchLoop:
	for {
		// Find the allocation after this free span.
		alloc := *dst

		// Find the end of this free span.
		var end uintptr
		if alloc != nil {
			// The allocation terminates the free span.
			end = uintptr(unsafe.Pointer(alloc))
		} else {
			// The free span ends at the end of the heap.
			end = heapEnd
		}

		// Calculate the size of the free span.
		freeSpanSize := end - start

		switch {
		case freeSpanSize == allocSize:
			// This span is a perfect fit.
			bestDst = dst
			bestStart = start
			break searchLoop
		case freeSpanSize > allocSize && (bestDst == nil || bestSize > allocSize):
			// This span is a better fit than the previous best.
			bestDst = dst
			bestStart = start
			bestSize = freeSpanSize
		}

		// Move to the next free span.
		if alloc == nil {
			// That was the last free region.
			break searchLoop
		}
		start = align(uintptr(unsafe.Pointer(alloc)) + unsafe.Sizeof(allocHeader{}) + alloc.len)
		dst = &alloc.next
	}

	if bestDst == nil {
		// There is no suitable allocation.
		return nil
	}

	mem := (*allocHeader)(unsafe.Pointer(bestStart))
	*bestDst, mem.next = mem, *bestDst
	mem.len = size
	return mem
}

func free(ptr unsafe.Pointer) {
	// TODO: free allocations on request, when the compiler knows they're unused.
}

// GC performs a garbage collection cycle.
func GC() {
	if gcAsserts {
		gcRunning = true
	}

	if gcDebug {
		println("running collection cycle...")
	}

	// Mark in-use allocations.
	remaining := mark()

	if gcDebug {
		println("found unreferenced allocations:")
		activeAllocs.dump()
	}

	// Update the active allocations list.
	activeAllocs = remaining

	if gcDebug {
		println("sorting active allocations")
	}

	activeAllocs.sort()

	if gcDebug {
		println("garbage collection cycle complete")
	}

	if gcAsserts {
		gcRunning = false
	}
}

// mark all in-use allocations and save them as a sorted list.
func mark() allocList {
	if gcDebug {
		println("running GC mark phase...")
		println("pre-mark allocations:")
		activeAllocs.dump()
	}

	// Mark all allocations referenced by currently-executing functions.
	markStack()

	// Mark all allocations referenced by global variables.
	markGlobals()

	// Save remaining allocations to a list after they have been marked and scanned.
	var remainingAllocs allocList

	// Channel operations in interrupts may move task pointers around while we are marking.
	// Therefore we need to scan the runqueue seperately.
	var markedTaskQueue task.Queue
runqueueScan:
	if baremetal && hasScheduler {
		for !runqueue.Empty() {
			// Pop the next task off of the runqueue.
			t := runqueue.Pop()

			// Mark the task if it has not already been marked.
			markRoot(uintptr(unsafe.Pointer(&runqueue)), uintptr(unsafe.Pointer(t)))

			// Push the task onto our temporary queue.
			markedTaskQueue.Push(t)
		}
	}

	// Mark allocations referenced by other marked allocations.
	for {
		// Pop an allocation off of the scan list.
		alloc := scanList.pop()
		if alloc == nil {
			break
		}

		// Scan the allocation.
		alloc.scan()

		// Save the allocation in the remaining allocations list.
		remainingAllocs.push(alloc)
	}

	if baremetal && hasScheduler {
		// Restore the runqueue.
		i := interrupt.Disable()
		if !runqueue.Empty() {
			// Something new came in while finishing the mark.
			interrupt.Restore(i)
			if gcDebug {
				println("scanning tasks queued during GC")
			}
			goto runqueueScan
		}
		runqueue = markedTaskQueue
		interrupt.Restore(i)
	}

	if gcDebug {
		println("GC mark phase complete")
		println("remaining allocations:")
		remainingAllocs.dump()
	}

	return remainingAllocs
}

// sort the allocation list in ascending address order.
func (list *allocList) sort() {
	// Sort the list by repeatedly moving the lowest-address element to the front.
	// This is selection sort (https://en.wikipedia.org/wiki/Selection_sort).
	for dst := &list.first; *dst != nil; dst = &((*dst).next) {
		// Find the next lowest-address allocation.
		src := dst
		min := *src
		for ref, alloc := src, min; alloc != nil; ref, alloc = &alloc.next, alloc.next {
			if uintptr(unsafe.Pointer(alloc)) < uintptr(unsafe.Pointer(min)) {
				src, min = ref, alloc
			}
		}

		// Remove the min from the list.
		*src = min.next

		// Insert the min at dst.
		*dst, min.next = min, *dst
	}

	if gcDebug {
		println("sorted allocations:")
		list.dump()
	}

	if gcAsserts && list.first != nil {
		prev := list.first
		for alloc := list.first.next; alloc != nil; alloc, prev = alloc.next, alloc {
			if uintptr(unsafe.Pointer(alloc)) <= uintptr(unsafe.Pointer(prev)) {
				runtimePanic("alloc list not sorted")
			}
		}
	}

	if gcAsserts {
		var prevEnd uintptr
		for alloc := list.first; alloc != nil; alloc, prevEnd = alloc.next, uintptr(unsafe.Pointer(alloc))+unsafe.Sizeof(allocHeader{})+alloc.len {
			start := uintptr(unsafe.Pointer(alloc))
			if start < prevEnd {
				runtimePanic("overlapping allocations")
			}
		}
	}
}

// scan the allocation for pointers to other allocations.
func (alloc *allocHeader) scan() {
	start := uintptr(unsafe.Pointer(alloc)) + unsafe.Sizeof(allocHeader{})
	end := start + alloc.len

	if gcDebug {
		println("scanning allocation at", start, "of", alloc.len, "bytes")
	}

	markRoots(start, end)
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

	for addr := start; addr+unsafe.Sizeof(unsafe.Pointer(nil)) <= end; addr += unsafe.Alignof(addr) {
		root := *(*uintptr)(unsafe.Pointer(addr))
		markRoot(addr, root)
	}
}

// mark a GC root at the address addr.
func markRoot(addr, root uintptr) {
	if !looksLikePointer(root) {
		// This is not a pointer.
		// Skip searching for it.
		return
	}

	// Search the list of unmarked allocations for this root address.
	alloc := activeAllocs.search(root)
	if alloc == nil {
		// The pointer is either a false positive (not a pointer) or already marked.
		// Either way, there is nothing to do.
		return
	}
	if gcDebug {
		println("found unmarked pointer", root, "at address", addr)
	}

	// Push the allocation onto the scan list.
	scanList.push(alloc)
}

// looksLikePointer returns whether this could be a pointer. Currently, it
// simply returns whether it lies anywhere in the heap. Go allows interior
// pointers so we can't check alignment or anything like that.
func looksLikePointer(ptr uintptr) bool {
	return ptr >= heapStart && ptr < heapEnd
}

// activeAllocs is a list of allocations which are currently in use.
// During the mark pass, marked allocations are removed from here and placed onto the scan list.
// This list is stored in ascending address order.
var activeAllocs allocList

// scanList is a list of allocations which have been marked but not yet scanned.
var scanList allocList

// allocList is a list of allocations.
type allocList struct {
	first *allocHeader
}

// push an allocation to the front of the list.
func (list *allocList) push(node *allocHeader) {
	list.first, node.next = node, list.first
}

// pop an allocation from the front of the list.
// If the list is empty, this will return nil.
func (list *allocList) pop() *allocHeader {
	if list.first == nil {
		return nil
	}

	popped := list.first
	list.first = popped.next

	return popped
}

// search finds and removes the allocation containing an address.
// This assumes that the list is sorted.
// If no allocation contains this address, this will return nil.
func (list *allocList) search(addr uintptr) *allocHeader {
	// While traversing the list, keep the pointer to the pointer to the current allocation list entry.
	// This is used later to remove it from the list.
	prev := &list.first
	for alloc := list.first; alloc != nil; alloc, prev = alloc.next, &alloc.next {
		// Find the start of the memory after the allocation header.
		start := uintptr(unsafe.Pointer(alloc)) + unsafe.Sizeof(allocHeader{})
		if addr < start {
			return nil
		}

		// Check if the address is before the end of the allocation.
		// This check calculates the distance between the start of the allocation and compares that to the length.
		// This works around an overflow when the last byte of an allocation is ^uintptr(0).
		if addr-start >= alloc.len {
			continue
		}

		// The address is within this allocation.
		// Remove the allocation from the list by replacing the reference to it with the next allocation.
		*prev = alloc.next
		alloc.next = nil

		return alloc
	}

	return nil
}

// dump the contents of the list to the debug output.
func (list *allocList) dump() {
	for alloc := list.first; alloc != nil; alloc = alloc.next {
		start := uintptr(unsafe.Pointer(alloc)) + unsafe.Sizeof(allocHeader{})
		println("-", "[", start, ",", start+alloc.len, "):", alloc.len, "bytes")
	}
}

// allocHeader is a header placed before an allocation in order to store metadata for the garbage collector.
// It is used to form a linked list of allocations.
type allocHeader struct {
	// next is a pointer to the next allocation.
	next *allocHeader

	// len is the length of the body of this node in bytes.
	len uintptr
}

func KeepAlive(x interface{}) {
	// Unimplemented. Only required with SetFinalizer().
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}
