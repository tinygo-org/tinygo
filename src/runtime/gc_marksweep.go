// +build gc.marksweep

package runtime

import (
	"unsafe"
)

func init() {
	initHeap()
}

func alloc(size uintptr) unsafe.Pointer {
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
	markRoots(getCurrentStackPointer(), stackTop) // assume a descending stack

	// Sweep phase: free all non-marked objects and unmark marked objects for
	// the next collection cycle.
	sweep()

	// Show how much has been sweeped, for debugging.
	if gcDebug {
		dumpHeap()
	}
}
