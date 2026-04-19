//go:build baremetal && buildmode.c_archive

package runtime

var (
	heapStart    uintptr
	heapEnd      uintptr
	globalsStart uintptr
	globalsEnd   uintptr
	stackTop     uintptr
)

// Allows C consumers of the library to set the GC variables.
//
//export tinygo_init
func tinygo_init(heap, heapSize, glob, globEnd, stack uintptr) {
	heapStart, heapEnd = heap, heap+heapSize
	globalsStart, globalsEnd = glob, globEnd
	stackTop = stack
	initRand()
	initHeap()
	initAll()
}
