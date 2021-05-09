// +build gc.extalloc

package runtime

import (
	"internal/task"
	"runtime/interrupt"
	"unsafe"
)

// This garbage collector implementation allows TinyGo to use an external memory allocator.
// It keeps a list of allocations for tracking purposes.
// This is also a conservative collector.

const (
	gcDebug   = false
	gcAsserts = false
)

func initHeap() {}

var allocList []allocListEntry

// allocListEntry is a listing of a single heap allocation.
type allocListEntry struct {
	start, end uintptr
	next       *allocListEntry
}

// sortAllocList sorts an allocation list using heapsort.
// HeapSort performance is awful but is simple-ish and doesn't require extra memory.
//go:noinline
func sortAllocList(list []allocListEntry) {
	// Turn the array into a max heap.
	for i, v := range list {
		// Repeatedly swap v up the heap until the node above is at a greater address (or the top of the heap is reached).
		for i > 0 && v.start > list[(i-1)/2].start {
			list[i] = list[(i-1)/2]
			i = (i - 1) / 2
		}
		list[i] = v
	}

	// Repeatedly remove the max and place it at the end of the array.
	for len(list) > 1 {
		// Remove the max and place it at the end of the array.
		list[0], list[len(list)-1] = list[len(list)-1], list[0]
		list = list[:len(list)-1]

		// Fix the position of the element we swapped into the root.
		i := 0
		for {
			// Find the element that should actually be at this position.
			max := i
			if l := 2*i + 1; l < len(list) && list[l].start > list[max].start {
				max = l
			}
			if r := 2*i + 2; r < len(list) && list[r].start > list[max].start {
				max = r
			}

			if max == i {
				// The element is where it is supposed to be.
				break
			}

			// Swap this element down the heap.
			list[i], list[max] = list[max], list[i]
			i = max
		}
	}
}

// searchAllocList searches a sorted alloc list for an address.
// If the address is found in an allocation, a pointer to the corresponding entry is returned.
// Otherwise, this returns nil.
func searchAllocList(list []allocListEntry, addr uintptr) *allocListEntry {
	for len(list) > 0 {
		mid := len(list) / 2
		switch {
		case addr < list[mid].start:
			list = list[:mid]
		case addr > list[mid].end:
			list = list[mid+1:]
		default:
			return &list[mid]
		}
	}

	return nil
}

// usedMem is the total amount of allocated memory (including the alloc list).
var usedMem uintptr

// scanQueue is a queue of marked allocations to scan.
var scanQueue *allocListEntry

// mark searches for an allocation containing the given address and marks it if found.
func mark(addr uintptr) bool {
	if len(allocList) == 0 {
		// The heap is empty.
		return false
	}

	if addr < allocList[0].start || addr > allocList[len(allocList)-1].end {
		// Pointer is outside of allocated bounds.
		return false
	}

	// Search the allocation list for this address.
	alloc := searchAllocList(allocList, addr)
	if alloc != nil && alloc.next == nil {
		if gcDebug {
			println("mark:", addr)
		}

		// Push the allocation onto the scan queue.
		next := scanQueue
		if next == nil {
			// Insert a loop so we can tell that this isn't marked.
			next = alloc
		}
		scanQueue, alloc.next = alloc, next

		return true
	}

	// The address does not reference an unmarked allocation.
	return false
}

func markRoot(addr uintptr, root uintptr) {
	marked := mark(root)
	if gcDebug {
		if marked {
			println("marked root:", root, "at", addr)
		} else if addr != 0 {
			println("did not mark root:", root, "at", addr)
		}
	}
}

func markRoots(start uintptr, end uintptr) {
	scan(start, end)
}

// scan loads all pointer-aligned words and marks any pointers that it finds.
func scan(start uintptr, end uintptr) {
	// Align start pointer.
	start = (start + unsafe.Alignof(unsafe.Pointer(nil)) - 1) &^ (unsafe.Alignof(unsafe.Pointer(nil)) - 1)

	// Mark all pointers.
	for ptr := start; ptr+unsafe.Sizeof(unsafe.Pointer(nil)) <= end; ptr += unsafe.Alignof(unsafe.Pointer(nil)) {
		mark(*(*uintptr)(unsafe.Pointer(ptr)))
	}
}

// scan marks all allocations referenced by this allocation.
// This should only be invoked by the garbage collector.
func (e *allocListEntry) scan() {
	scan(e.start, e.end)
}

// gcrunning is used by gcAsserts to determine whether the garbage collector is running.
// This is used to detect if the collector is invoking itself or trying to allocate memory.
var gcrunning bool

func GC() {
	if gcDebug {
		println("running GC")
	}

	if gcAsserts {
		if gcrunning {
			runtimePanic("GC called itself")
		}
		gcrunning = true
	}

	// Sort the allocation list so that it can be efficiently searched.
	sortAllocList(allocList)

	// Unmark all allocations.
	for i := range allocList {
		allocList[i].next = nil
	}

	// Reset the scan queue.
	scanQueue = nil

	if gcDebug {
		println("pre-GC allocations:")
		for _, v := range allocList {
			println("  [", v.start, ",", v.end, "]")
		}
	}

	if gcAsserts && len(allocList) > 1 {
		for i, v := range allocList[1:] {
			if allocList[i].start >= v.start {
				runtimePanic("alloc list not sorted")
			}
		}
	}

	// Start by scanning the stack.
	markStack()

	// Scan all globals.
	markGlobals()

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

	// Scan all referenced allocations.
	for scanQueue != nil {
		// Pop a marked allocation off of the scan queue.
		alloc := scanQueue
		next := alloc.next
		if next == alloc {
			// This is the last value on the queue.
			next = nil
		}
		scanQueue = next

		// Scan and mark all allocations that this references.
		alloc.scan()
	}

	i := interrupt.Disable()
	if !runqueue.Empty() {
		// Something new came in while finishing the mark.
		interrupt.Restore(i)
		goto runqueueScan
	}
	runqueue = markedTaskQueue
	interrupt.Restore(i)

	// Free all remaining unmarked allocations.
	usedMem = 0
	j := 0
	for _, v := range allocList {
		if v.next == nil {
			// This was never marked.
			extfree(unsafe.Pointer(v.start))
			continue
		}

		// Move this down in the list.
		allocList[j] = v
		j++

		// Re-calculate used memory.
		usedMem += v.end - v.start
	}
	allocList = allocList[:j]

	if gcDebug {
		println("post-GC allocations:")
		for _, v := range allocList {
			println("  [", v.start, ",", v.end, "]")
		}
		println("GC finished")
	}

	if gcAsserts {
		gcrunning = false
	}
}

// heapBound is used to control the growth of the heap.
// When the heap exceeds this size, the garbage collector is run.
// If the garbage collector cannot free up enough memory, the bound is doubled until the allocation fits.
var heapBound uintptr = 4 * unsafe.Sizeof(unsafe.Pointer(nil))

// zeroSizedAlloc is just a sentinel that gets returned when allocating 0 bytes.
var zeroSizedAlloc uint8

// alloc tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
//go:noinline
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	if size == 0 {
		return unsafe.Pointer(&zeroSizedAlloc)
	}

	if gcAsserts && gcrunning {
		runtimePanic("allocated inside the garbage collector")
	}

	var gcRan bool
	for {
		// Try to bound heap growth.
		if usedMem+size < usedMem {
			if gcDebug {
				println("current mem:", usedMem, "alloc size:", size)
			}
			runtimePanic("target heap size exceeds address space size")
		}
		if usedMem+size > heapBound {
			if !gcRan {
				// Run the garbage collector before growing the heap.
				if gcDebug {
					println("heap reached size limit")
				}
				GC()
				gcRan = true
				continue
			} else {
				// Grow the heap bound to fit the allocation.
				for heapBound != 0 && usedMem+size > heapBound {
					heapBound <<= 1
				}
				if heapBound == 0 {
					// This is only possible on hosted 32-bit systems.
					// Allow the heap bound to encompass everything.
					heapBound = ^uintptr(0)
				}
				if gcDebug {
					println("raising heap size limit to", heapBound)
				}
			}
		}

		// Ensure that there is space in the alloc list.
		if len(allocList) == cap(allocList) {
			// Attempt to double the size of the alloc list.
			newCap := 2 * uintptr(cap(allocList))
			if newCap == 0 {
				newCap = 1
			}

			oldList := allocList

			newListPtr := extalloc(newCap * unsafe.Sizeof(allocListEntry{}))
			if newListPtr == nil {
				if gcRan {
					// Garbage collector was not able to free up enough memory.
					runtimePanic("out of memory")
				} else {
					// Run the garbage collector and try again.
					GC()
					gcRan = true
					continue
				}
			}

			var newList []allocListEntry
			newListHeader := (*struct {
				ptr unsafe.Pointer
				len uintptr
				cap uintptr
			})(unsafe.Pointer(&newList))
			newListHeader.ptr = newListPtr
			newListHeader.len = uintptr(len(oldList))
			newListHeader.cap = newCap

			copy(newList, oldList)
			allocList = newList

			if cap(oldList) != 0 {
				free(unsafe.Pointer(&oldList[0]))
			}
		}

		// Allocate the memory.
		ptr := extalloc(size)
		if ptr == nil {
			if gcDebug {
				println("extalloc failed")
			}
			if gcRan {
				// Garbage collector was not able to free up enough memory.
				runtimePanic("out of memory")
			} else {
				// Run the garbage collector and try again.
				GC()
				gcRan = true
				continue
			}
		}

		// Add the allocation to the list.
		i := len(allocList)
		allocList = allocList[:i+1]
		allocList[i] = allocListEntry{
			start: uintptr(ptr),
			end:   uintptr(ptr) + size,
		}

		// Zero the allocation.
		memzero(ptr, size)

		// Update used memory.
		usedMem += size

		if gcDebug {
			println("allocated:", uintptr(ptr), "size:", size)
			println("used memory:", usedMem)
		}

		return ptr
	}
}

func free(ptr unsafe.Pointer) {
	// Currently unimplemented due to bugs in coroutine lowering.
}

func KeepAlive(x interface{}) {
	// Unimplemented. Only required with SetFinalizer().
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}
