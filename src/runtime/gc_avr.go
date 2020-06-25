// +build gc.conservative
// +build avr

package runtime

// This memory manager is partially inspired by the memory allocator included in avr-libc.
// Unlike avr-libc malloc, this memory manager stores an allocation list instead of a free list.
// This gives an overhead of 4 bytes/allocation (up from avr-libc's 2 bytes/allocation).
// The allocation list is stored in ascending address order.
// The set of free spans is implicitly derived from the gaps between allocations.
// This allocator has an effective allocation complexity of O(n) and worst-case GC complexity of O(n^2).
// Due to architectural quirks of AVR, as well as tiny heaps, this should almost always be faster than the standard conservative collector.

import "unsafe"

// isInHeap checks if an address is inside the heap.
func isInHeap(addr uintptr) bool {
	return addr >= heapStart && addr < heapEnd
}

// allocNode is a node in the allocations list.
// It is prepended to every allocation, resulting in an overhead of 4 bytes per allocation.
type allocNode struct {
	// next is a pointer to the next allocation node.
	next *allocNode

	// len is the length of the body of this node in bytes.
	len uintptr

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
	markMem(start, end)
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

// markMem scans a memory region for pointers and marks anything that is pointed to.
func markMem(start, end uintptr) {
	if start >= end {
		return
	}
	prevByte := *(*byte)(unsafe.Pointer(start))
	for ; start != end; start++ {
		b := *(*byte)(unsafe.Pointer(start))
		addr := (uintptr(b) << 8) | uintptr(prevByte)
		markAddr(addr)
		prevByte = b
	}
}

// GC runs a garbage collection cycle.
func GC() {
	// Mark phase: mark all reachable objects, recursively.
	markGlobals()
	markStack()
	var keep *allocNode
	for scanList != nil {
		// Pop a node off of the scan list.
		node := scanList
		scanList = node.next

		// Scan the node.
		baseAddr := uintptr(unsafe.Pointer(&node.base))
		endAddr := baseAddr + node.len
		markMem(baseAddr, endAddr)

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
		start = uintptr(unsafe.Pointer(&node.base)) + node.len
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

// alloc tries to find some free space on the heap, possibly doing a garbage
// collection cycle if needed. If no space is free, it panics.
//go:noinline
func alloc(size uintptr) unsafe.Pointer {
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

	// Zero the allocation and return it.
	ptr := unsafe.Pointer(&node.base)
	memzero(ptr, size)
	return ptr
}

func free(ptr unsafe.Pointer) {
	// TODO: free memory on request, when the compiler knows it is unused.
}

func initHeap() {
	// This memory manager requires no initialization other than the zeroing of globals.
	// This function is provided for compatability with other memory managers.
}
