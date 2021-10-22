// +build gc.extalloc

package runtime

import (
	"internal/task"
	"runtime/interrupt"
	"unsafe"
)

// This garbage collector implementation allows TinyGo to use an external memory allocator.
// It appends a header to the end of every allocation which the garbage collector uses for tracking purposes.
// This is also a conservative collector.

const (
	gcDebug   = false
	gcAsserts = false
)

func initHeap() {}

// memTreap is a treap which is used to track allocations for the garbage collector.
type memTreap struct {
	root *memTreapNode
}

// printNode recursively prints a subtree at a given indentation depth.
func (t *memTreap) printNode(n *memTreapNode, depth int) {
	for i := 0; i < depth; i++ {
		print(" ")
	}
	println(n, n.priority())
	if n == nil {
		return
	}
	if gcAsserts && n.parent == nil && t.root != n {
		runtimePanic("parent missing")
	}
	t.printNode(n.left, depth+1)
	t.printNode(n.right, depth+1)
}

// print the treap.
func (t *memTreap) print() {
	println("treap:")
	t.printNode(t.root, 1)
}

// empty returns whether the treap contains any nodes.
func (t *memTreap) empty() bool {
	return t.root == nil
}

// minAddr returns the lowest address contained in an allocation in the treap.
func (t *memTreap) minAddr() uintptr {
	// Find the rightmost node.
	n := t.root
	for n.right != nil {
		n = n.right
	}

	// The lowest address is the base of the rightmost node.
	return uintptr(unsafe.Pointer(&n.base))
}

// maxAddr returns the highest address contained in an allocation in the treap.
func (t *memTreap) maxAddr() uintptr {
	// Find the leftmost node.
	n := t.root
	for n.left != nil {
		n = n.left
	}

	// The highest address is the end of the leftmost node.
	return uintptr(unsafe.Pointer(&n.base)) + n.size
}

// rotateRight does a right rotation of p and q.
// https://en.wikipedia.org/wiki/Tree_rotation#/media/File:Tree_rotation.png
func (t *memTreap) rotateRight(p, q *memTreapNode) {
	if t.root == q {
		t.root = p
	} else {
		*q.parentSlot() = p
	}

	//a := p.left
	b := p.right
	//c := q.right

	p.parent = q.parent
	p.right = q

	q.parent = p
	q.left = b

	if b != nil {
		b.parent = q
	}
}

// rotateLeft does a left rotation of p and q.
// https://en.wikipedia.org/wiki/Tree_rotation#/media/File:Tree_rotation.png
func (t *memTreap) rotateLeft(p, q *memTreapNode) {
	if t.root == p {
		t.root = q
	} else {
		*p.parentSlot() = q
	}

	//a := p.left
	b := q.left
	//c := q.right

	q.parent = p.parent
	q.left = p

	p.parent = q
	p.right = b

	if b != nil {
		b.parent = p
	}
}

// rotate rotates a lower node up to its parent.
// The node n must be a child of m, and will be the parent of m after the rotation.
func (t *memTreap) rotate(n, m *memTreapNode) {
	// https://en.wikipedia.org/wiki/Tree_rotation#/media/File:Tree_rotation.png
	if uintptr(unsafe.Pointer(n)) > uintptr(unsafe.Pointer(m)) {
		t.rotateRight(n, m)
	} else {
		t.rotateLeft(m, n)
	}
}

// insert a node into the treap.
func (t *memTreap) insert(n *memTreapNode) {
	if gcAsserts && (n.parent != nil || n.left != nil || n.right != nil) {
		runtimePanic("tried to insert unzeroed treap node")
	}

	if t.root == nil {
		// This is the first node, and can be inserted directly into the root.
		t.root = n
		return
	}

	// Insert like a regular binary search tree.
	for n.parent = t.root; *n.parentSlot() != nil; n.parent = *n.parentSlot() {
	}
	*n.parentSlot() = n

	// Rotate the tree to restore the heap invariant.
	priority := n.priority()
	for n.parent != nil && priority > n.parent.priority() {
		t.rotate(n, n.parent)
	}
}

// lookupAddr finds the treap node with the allocation containing the specified address.
// If the address is not contained in any allocations in this treap, nil is returned.
// NOTE: fields of memTreapNodes are not considered part of the allocations.
func (t *memTreap) lookupAddr(addr uintptr) *memTreapNode {
	n := t.root
	for n != nil && !n.contains(addr) {
		if addr > uintptr(unsafe.Pointer(n)) {
			n = n.left
		} else {
			n = n.right
		}
	}

	return n
}

// replace a node with another node on the treap.
func (t *memTreap) replace(old, new *memTreapNode) {
	if gcAsserts && (old == nil || new == nil) {
		if gcDebug {
			println("tried to replace:", old, "->", new)
		}
		runtimePanic("invalid replacement")
	}
	if gcAsserts && old.parent == nil && old != t.root {
		if gcDebug {
			println("tried to replace:", old, "->", new)
			t.print()
		}
		runtimePanic("corrupted tree")
	}
	new.parent = old.parent
	if old == t.root {
		t.root = new
	} else {
		*new.parentSlot() = new
	}
}

// remove a node from the treap.
// This does not free the allocation.
func (t *memTreap) remove(n *memTreapNode) {
scan:
	for {
		switch {
		case n.left == nil && n.right == nil && n.parent == nil:
			// This is the only node - uproot it.
			t.root = nil
			break scan
		case n.left == nil && n.right == nil:
			// There are no nodes beneath here, so just remove this node from the parent.
			*n.parentSlot() = nil
			break scan
		case n.left != nil && n.right == nil:
			t.replace(n, n.left)
			break scan
		case n.right != nil && n.left == nil:
			t.replace(n, n.right)
			break scan
		default:
			// Rotate this node downward.
			if n.left.priority() > n.right.priority() {
				t.rotate(n.left, n)
			} else {
				t.rotate(n.right, n)
			}
		}
	}

	n.left = nil
	n.right = nil
	n.parent = nil
}

// memTreapNode is a treap node used to track allocations for the garbage collector.
// This struct is prepended to every allocation.
type memTreapNode struct {
	parent, left, right *memTreapNode
	size                uintptr
	base                struct{}
}

// priority computes a pseudo-random priority value for this treap node.
// This value is a fibonacci hash (https://en.wikipedia.org/wiki/Hash_function#Fibonacci_hashing) of the node's memory address.
func (n *memTreapNode) priority() uintptr {
	// Select fibonacci multiplier for this bit-width.
	var fibonacciMultiplier uint64
	switch 8 * unsafe.Sizeof(uintptr(0)) {
	case 16:
		fibonacciMultiplier = 40503
	case 32:
		fibonacciMultiplier = 2654435769
	case 64:
		fibonacciMultiplier = 11400714819323198485
	default:
		runtimePanic("invalid size of uintptr")
	}

	// Hash the pointer.
	return uintptr(fibonacciMultiplier) * uintptr(unsafe.Pointer(n))
}

// contains returns whether this allocation contains a given address.
func (n *memTreapNode) contains(addr uintptr) bool {
	return addr >= uintptr(unsafe.Pointer(&n.base)) && addr < uintptr(unsafe.Pointer(&n.base))+n.size
}

// parentSlot returns a pointer to the parent's reference to this node.
func (n *memTreapNode) parentSlot() **memTreapNode {
	if uintptr(unsafe.Pointer(n)) > uintptr(unsafe.Pointer(n.parent)) {
		return &n.parent.left
	} else {
		return &n.parent.right
	}
}

// memScanQueue is a queue of memTreapNodes.
type memScanQueue struct {
	head, tail *memTreapNode
}

// push adds an allocation onto the queue.
func (q *memScanQueue) push(n *memTreapNode) {
	if gcAsserts && (n.left != nil || n.right != nil || n.parent != nil) {
		runtimePanic("tried to push a treap node that is in use")
	}

	if q.head == nil {
		q.tail = n
	} else {
		q.head.left = n
	}
	n.right = q.head
	q.head = n
}

// pop removes the next allocation from the queue.
func (q *memScanQueue) pop() *memTreapNode {
	n := q.tail
	q.tail = n.left
	if q.tail == nil {
		q.head = nil
	}
	n.left = nil
	n.right = nil
	return n
}

// empty returns whether the queue contains any allocations.
func (q *memScanQueue) empty() bool {
	return q.tail == nil
}

// allocations is a treap containing all allocations.
var allocations memTreap

// usedMem is the total amount of allocated memory (including the space taken up by memory treap nodes).
var usedMem uintptr

// firstPtr and lastPtr are the bounds of memory used by the heap.
// They are computed before the collector starts marking, and are used to quickly eliminate false positives.
var firstPtr, lastPtr uintptr

// scanQueue is a queue of marked allocations to scan.
var scanQueue memScanQueue

// mark searches for an allocation containing the given address and marks it if found.
func mark(addr uintptr) bool {
	if addr < firstPtr || addr > lastPtr {
		// Pointer is outside of allocated bounds.
		return false
	}

	node := allocations.lookupAddr(addr)
	if node != nil {
		if gcDebug {
			println("mark:", addr)
		}
		allocations.remove(node)
		scanQueue.push(node)
	}

	return node != nil
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
	// Align start and end pointers.
	start = (start + unsafe.Alignof(unsafe.Pointer(nil)) - 1) &^ (unsafe.Alignof(unsafe.Pointer(nil)) - 1)
	end &^= unsafe.Alignof(unsafe.Pointer(nil)) - 1

	// Mark all pointers.
	for ptr := start; ptr < end; ptr += unsafe.Alignof(unsafe.Pointer(nil)) {
		mark(*(*uintptr)(unsafe.Pointer(ptr)))
	}
}

// scan marks all allocations referenced by this allocation.
// This should only be invoked by the garbage collector.
func (n *memTreapNode) scan() {
	start := uintptr(unsafe.Pointer(&n.base))
	end := start + n.size
	scan(start, end)
}

// destroy removes and frees all allocations in the treap.
func (t *memTreap) destroy() {
	n := t.root
	for n != nil {
		switch {
		case n.left != nil:
			// Destroy the left subtree.
			n = n.left
		case n.right != nil:
			// Destroy the right subtree.
			n = n.right
		default:
			// This is a leaf node, so delete it and jump back to the parent.

			// Save the parent to jump back to.
			parent := n.parent

			if parent != nil {
				*n.parentSlot() = nil
			} else {
				t.root = nil
			}

			// Update used memory.
			usedMem -= unsafe.Sizeof(memTreapNode{}) + n.size
			if gcDebug {
				println("collecting:", &n.base, "size:", n.size)
				println("used memory:", usedMem)
			}

			// Free the node.
			extfree(unsafe.Pointer(n))

			// Jump back to the parent node.
			n = parent
		}
	}
}

// gcrunning is used by gcAsserts to determine whether the garbage collector is running.
// This is used to detect if the collector is invoking itself or trying to allocate memory.
var gcrunning bool

// activeMem is a queue used to store marked allocations which have already been scanned.
// This is only used when the garbage collector is running.
var activeMem memScanQueue

func GC() {
	if gcDebug {
		println("running GC")
	}
	if allocations.empty() {
		// Skip collection because the heap is empty.
		if gcDebug {
			println("nothing to collect")
		}
		return
	}

	if gcAsserts {
		if gcrunning {
			runtimePanic("GC called itself")
		}
		gcrunning = true
	}

	if gcDebug {
		println("pre-GC allocations:")
		allocations.print()
	}

	// Before scanning, find the lowest and highest allocated pointers.
	// These can be quickly compared against to eliminate most false positives.
	firstPtr, lastPtr = allocations.minAddr(), allocations.maxAddr()

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

	// Scan all referenced allocations, building a new treap with marked allocations.
	// The marking process deletes the allocations from the old allocations treap, so they are only queued once.
	for !scanQueue.empty() {
		// Pop a marked node off of the scan queue.
		n := scanQueue.pop()

		// Scan and mark all nodes that this references.
		n.scan()

		// Insert this node into the active memory queue.
		activeMem.push(n)
	}

	i := interrupt.Disable()
	if !runqueue.Empty() {
		// Something new came in while finishing the mark.
		interrupt.Restore(i)
		goto runqueueScan
	}
	runqueue = markedTaskQueue
	interrupt.Restore(i)

	// The allocations treap now only contains unreferenced nodes. Destroy them all.
	allocations.destroy()
	if gcAsserts && !allocations.empty() {
		runtimePanic("failed to fully destroy allocations")
	}

	// Treapify the active memory queue.
	for !activeMem.empty() {
		allocations.insert(activeMem.pop())
	}

	if gcDebug {
		println("GC finished")
	}

	if gcAsserts {
		gcrunning = false
	}
}

// heapBound is used to control the growth of the heap.
// When the heap exceeds this size, the garbage collector is run.
// If the garbage collector cannot free up enough memory, the bound is doubled until the allocation fits.
var heapBound uintptr = 4 * unsafe.Sizeof(memTreapNode{})

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

	// Calculate size of allocation including treap node.
	allocSize := unsafe.Sizeof(memTreapNode{}) + size

	var gcRan bool
	for {
		// Try to bound heap growth.
		if usedMem+allocSize < usedMem {
			if gcDebug {
				println("current mem:", usedMem, "alloc size:", allocSize)
			}
			runtimePanic("target heap size exceeds address space size")
		}
		if usedMem+allocSize > heapBound {
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
				for heapBound != 0 && usedMem+allocSize > heapBound {
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

		// Allocate the memory.
		ptr := extalloc(allocSize)
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

		// Initialize the memory treap node.
		node := (*memTreapNode)(ptr)
		*node = memTreapNode{
			size: size,
		}

		// Insert allocation into the allocations treap.
		allocations.insert(node)

		// Extract the user's section of the allocation.
		ptr = unsafe.Pointer(&node.base)
		if gcAsserts && !node.contains(uintptr(ptr)) {
			runtimePanic("node is not self-contained")
		}
		if gcAsserts {
			check := allocations.lookupAddr(uintptr(ptr))
			if check == nil {
				if gcDebug {
					println("failed to find:", ptr)
					allocations.print()
				}
				runtimePanic("bad insert")
			}
		}

		// Zero the allocation.
		memzero(ptr, size)

		// Update used memory.
		usedMem += allocSize

		if gcDebug {
			println("allocated:", uintptr(ptr), "size:", size)
			println("used memory:", usedMem)
		}

		return ptr
	}
}

func realloc(ptr unsafe.Pointer, size uintptr) unsafe.Pointer {
	runtimePanic("unimplemented: gc_extalloc.realloc")
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
