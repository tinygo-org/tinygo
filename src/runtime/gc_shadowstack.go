// +build gc.shadowstack

package runtime

import (
	"unsafe"
)

//go:extern __data_start
var dataStartSymbol unsafe.Pointer

//go:extern _edata
var dataEndSymbol unsafe.Pointer

//go:extern __bss_start
var bssStartSymbol unsafe.Pointer

//go:extern _ebss
var bssEndSymbol unsafe.Pointer

var (
	dataStart = uintptr(unsafe.Pointer(&dataStartSymbol))
	dataEnd   = uintptr(unsafe.Pointer(&dataEndSymbol))
	bssStart  = uintptr(unsafe.Pointer(&bssStartSymbol))
	bssEnd    = uintptr(unsafe.Pointer(&bssEndSymbol))
)

// Constant value with metadata about a given function.
type frameMap struct {
	numRoots int32
	numMeta  int32
	meta     [0]uintptr // unknown number of 'meta' pointers
}

// Stack-allocated struct with live pointers.
type stackEntry struct {
	next  *stackEntry
	frame *frameMap
	roots [0]uintptr // unknown number of root pointers
}

// Note: this symbol must be redefined to llvm_gc_root_chain by the linker.
// For example, you can use this linker option:
//     --defsym=tinygo_gc_root_chain=llvm_gc_root_chain
// The reason is that otherwise the shadow stack GC pass in LLVM tries to set an
// initial value to llvm_gc_root_chain of a different type.
//go:extern tinygo_gc_root_chain
var gcRootChain *stackEntry

func init() {
	initHeap()
}

func alloc(size uintptr) unsafe.Pointer {
	GC()
	return heapAlloc(size)
}

func free(ptr unsafe.Pointer) {
	heapFree(ptr)
}

// GC performs a garbage collection cycle.
func GC() {
	if gcDebug {
		println("running collection cycle...")
	}

	// Mark phase: mark all reachable objects, recursively.
	markRoots(globalsStart, globalsEnd)
	markStack()

	// Sweep phase: free all non-marked objects and unmark marked objects for
	// the next collection cycle.
	sweep()

	// Show how much has been sweeped, for debugging.
	if gcDebug {
		dumpHeap()
	}
}

// Mark all pointers from the stack using the LLVM shadow stack.
//go:nobounds
func markStack() {
	chain := gcRootChain
	// Traverse the linked-list of stack roots.
	for chain != nil {
		// Check each root.
		for i := int32(0); i < chain.frame.numRoots; i++ {
			root := chain.roots[i]
			size := uintptr(1)
			if i < chain.frame.numMeta {
				// This root has no metadata associated with it.
				// Therefore, it must be a single pointer.
				size = chain.frame.meta[i]
			}
			for i := uintptr(0); i < size; i++ {
				if looksLikePointer(root) {
					markRoot(root, 0)
				}
			}
		}
		chain = chain.next
	}
}
