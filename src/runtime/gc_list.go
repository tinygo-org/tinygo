// +build gc.list

package runtime

import (
	"internal/task"
	"runtime/interrupt"
	"unsafe"
)

// This memory manager is partially inspired by the memory allocator included in avr-libc.
// Unlike avr-libc malloc, this memory manager stores an allocation list instead of a free list.
// Additionally, this memory manager uses strongly typed memory - making it a precise GC.
// There is an overhead of 3 pointer-widths/allocation (up from avr-libc's 1 pointer-width/allocation).
// The allocation list is stored in ascending address order.
// The set of free spans is implicitly derived from the gaps between allocations.
// This allocator has an effective allocation complexity of O(n) and worst-case GC complexity of O(n^2).
// Due to architectural quirks, as well as tiny heaps, this should almost always be faster than the standard conservative collector on AVR.

// isInHeap checks if an address is inside the heap.
func isInHeap(addr uintptr) bool {
	return addr >= heapStart && addr < heapEnd
}

// gcType is a chunk of metadata used to semi-precisely scan an allocation.
// The compiler will produce constant globals which can be used in this form.
type gcType struct {
	// size is the element size, measured in increments of pointer alignment.
	size uintptr

	// data is a bitmap following the size value.
	// It is organized as a []byte, and set bits indicate places where a pointer may be present.
	data struct{}
}

// allocNode is a node in the allocations list.
// It is prepended to every allocation, resulting in an overhead of 4 bytes per allocation.
type allocNode struct {
	// next is a pointer to the next allocation node.
	next *allocNode

	// len is the length of the body of this node in bytes.
	len uintptr

	// typ is the memory type of this node.
	// If it is nil, there are no pointer slots to scan in the node.
	typ *gcType

	// base is the start of the body of this node.
	base struct{}
}

// allocList is a singly linked list of allocations.
// It is stored in ascending address order.
var allocList *allocNode

// scanList is a stack of allocations to scan.
var scanList *allocNode

func markRoot(addr, root uintptr) {
	markAddr(root)
}

func markRoots(start, end uintptr) {
	// Scan the data as a []unsafe.Pointer, using a temporary gcType value.
	// In the future, we can do precise scanning of the globals instead.
	t := struct {
		t    gcType
		data [1]byte
	}{
		t: gcType{
			size: 1, // 1 pointer-width
		},
		data: [1]byte{1}, // scan the pointer-width data
	}
	t.t.scan(start, end)
}

// markAddr marks the allocation containing the specified address.
func markAddr(addr uintptr) {
	if !isInHeap(addr) {
		// The address is not in the heap.
		return
	}

	// Search the allocation list for the address.
	for node, prev := allocList, &allocList; node != nil; node, prev = node.next, &node.next {
		baseAddr := uintptr(unsafe.Pointer(&node.base))
		if addr < baseAddr {
			// The address comes before this node.
			// Therefore the address must either be that of a node header or a free span.
			return
		}

		endAddr := baseAddr + node.len
		if addr < endAddr {
			// The address is included in this allocation.
			// Move the allocation to the scan stack.
			*prev = node.next
			scanList, node.next = node, scanList
			return
		}
	}
}

// scan the memory in [start, end) using the specified element type.
// If the type is larger than the region, the memory will be scanned as if it is a slice of that type.
func (t *gcType) scan(start, end uintptr) {
	if t == nil {
		// There are no pointers in the type.
		return
	}

	ptrSize, ptrAlign := unsafe.Sizeof(unsafe.Pointer(nil)), unsafe.Alignof(unsafe.Pointer(nil))

	// Align the start and end.
	start = align(start)
	end &^= ptrAlign - 1

	// Shift the end down so that we do not read past it.
	end -= ptrSize - ptrAlign

	width := t.size

	for start < end {
		// Process the bitmap a byte at a time.
		for i := uintptr(0); i <= width/8; i++ {
			mask := *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&t.data)) + i))
			for j := uintptr(0); mask != 0; j++ {
				if mask&1 != 0 {
					markAddr(*(*uintptr)(unsafe.Pointer(start + 8*i*ptrAlign + j*ptrAlign)))
				}
				mask >>= 1
			}
		}

		// Shift the start up and try to scan the next repeat.
		start += width * ptrAlign
	}
}

// GC runs a garbage collection cycle.
func GC() {
	// Mark phase: mark all reachable objects, recursively.
	markGlobals()
	markStack()
	var markedTaskQueue task.Queue
runqueueScan:
	if baremetal && hasScheduler {
		// Channel operations in interrupts may move task pointers around while we are marking.
		// Therefore we need to scan the runqueue seperately.
		for !runqueue.Empty() {
			// Pop the next task off of the runqueue.
			t := runqueue.Pop()

			// Mark the task if it has not already been marked.
			markRoot(uintptr(unsafe.Pointer(&runqueue)), uintptr(unsafe.Pointer(t)))

			// Push the task onto our temporary queue.
			markedTaskQueue.Push(t)
		}
	}

	var keep *allocNode
	for scanList != nil {
		// Pop a node off of the scan list.
		node := scanList
		scanList = node.next

		// Scan the node.
		if node.typ != nil {
			// The code-gen for AVR on LLVM has. . . issues.
			// Hoisting the nil check here saves ~50 clock cycles per iteration.
			baseAddr := uintptr(unsafe.Pointer(&node.base))
			endAddr := baseAddr + node.len
			node.typ.scan(baseAddr, endAddr)
		}

		// Insert the node into the output heap.
		var prev *allocNode
		keepNode := keep
		for keepNode != nil && uintptr(unsafe.Pointer(node)) > uintptr(unsafe.Pointer(keepNode)) {
			// Move onto the next node.
			prev, keepNode = keepNode, keepNode.next
		}
		if prev == nil {
			keep, node.next = node, keep
		} else {
			prev.next, node.next = node, keepNode
		}
	}

	if baremetal && hasScheduler {
		// Restore the runqueue.
		i := interrupt.Disable()
		if !runqueue.Empty() {
			// Something new came in while finishing the mark.
			interrupt.Restore(i)
			goto runqueueScan
		}
		runqueue = markedTaskQueue
		interrupt.Restore(i)
	}

	// Sweep phase: replace the heap.
	allocList = keep
}

// findMem searches the heap for a free span large enough to contain an allocation of the specified size.
// If there are no sufficiently large free spans available, this returns nil.
func findMem(size uintptr) *allocNode {
	// This memory allocator implementation is effectively the same algorithm applied by avr-libc.
	// It loops through the set of all free spans, and selects the smallest span that is large enough to fit the allocation.

	nodeSize := unsafe.Sizeof(allocNode{}) + size

	// best* store the best-fit free span.
	var bestDst **allocNode
	var bestStart uintptr
	var bestSize uintptr

	start := heapStart
	dst := &allocList
searchLoop:
	for {
		// Find the allocation node after this free span.
		node := *dst

		// Find the end of this free span.
		var end uintptr
		if node != nil {
			// The node terminates the free span.
			end = uintptr(unsafe.Pointer(node))
		} else {
			// The free span ends at the end of the heap.
			end = heapEnd
		}

		// Calculate the size of the free span.
		freeSpanSize := end - start

		switch {
		case freeSpanSize == nodeSize:
			// This span is a perfect fit.
			bestDst = dst
			bestStart = start
			break searchLoop
		case freeSpanSize > nodeSize && (bestDst == nil || bestSize > nodeSize):
			// This span is a better fit than the previous best.
			bestDst = dst
			bestStart = start
			bestSize = freeSpanSize
		}

		// Move to the next free span.
		if node == nil {
			// That was the last free region.
			break searchLoop
		}
		start = align(uintptr(unsafe.Pointer(&node.base)) + node.len)
		dst = &node.next
	}

	if bestDst == nil {
		// There is no suitable allocation.
		return nil
	}

	mem := (*allocNode)(unsafe.Pointer(bestStart))
	*bestDst, mem.next = mem, *bestDst
	mem.len = size
	return mem
}

// anyPtrType is a special fake type that is used when the type of an allocation is not known.
var anyPtrType = struct {
	t    gcType
	data [1]byte
}{
	t: gcType{
		size: 1, // 1 pointer-width
	},
	data: [1]byte{1}, // scan the pointer-width data
}

// alloc a chunk of untyped memory.
//go:inline
func alloc(size uintptr) unsafe.Pointer {
	// Use a placeholder type to scan the entire thing.
	return allocTyped(size, &anyPtrType.t.size)
}

// allocTyped tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
//go:noinline
func allocTyped(size uintptr, typ *uintptr) unsafe.Pointer {
	var ranGC bool
tryAlloc:
	// Search for available memory.
	node := findMem(size)
	if node == nil {
		// There is no free span large enough for the allocation.

		if ranGC {
			// Even after running the GC, there is not enough memory.
			runtimePanic("out of memory")
		}

		// Run the garbage collector and try again.
		GC()
		ranGC = true
		goto tryAlloc
	}

	// Zero the allocation.
	ptr := unsafe.Pointer(&node.base)
	memzero(ptr, size)

	// Apply the type to the allocation.
	node.typ = (*gcType)(unsafe.Pointer(typ))

	return ptr
}

func free(ptr unsafe.Pointer) {
	// TODO: free memory on request, when the compiler knows it is unused.
}

func initHeap() {
	// This memory manager requires no initialization other than the zeroing of globals.
	// This function is provided for compatability with other memory managers.
}
