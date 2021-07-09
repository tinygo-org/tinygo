// +build baremetal

package runtime

import (
	"unsafe"
)

//go:extern _heap_start
var heapStartSymbol [0]byte

//go:extern _heap_end
var heapEndSymbol [0]byte

//go:extern _globals_start
var globalsStartSymbol [0]byte

//go:extern _globals_end
var globalsEndSymbol [0]byte

//go:extern _stack_top
var stackTopSymbol [0]byte

var (
	heapStart    = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd      = uintptr(unsafe.Pointer(&heapEndSymbol))
	globalsStart = uintptr(unsafe.Pointer(&globalsStartSymbol))
	globalsEnd   = uintptr(unsafe.Pointer(&globalsEndSymbol))
	stackTop     = uintptr(unsafe.Pointer(&stackTopSymbol))
)

// growHeap tries to grow the heap size. It returns true if it succeeds, false
// otherwise.
func growHeap() bool {
	// On baremetal, there is no way the heap can be grown.
	return false
}

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	return alloc(size, nil)
}

//export free
func libc_free(ptr unsafe.Pointer) {
	free(ptr)
}

//export runtime_putchar
func runtime_putchar(c byte) {
	putchar(c)
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	exit(code)
}

const baremetal = true

// timeOffset is how long the monotonic clock started after the Unix epoch. It
// should be a positive integer under normal operation or zero when it has not
// been set.
var timeOffset int64

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	mono = nanotime()
	sec = (mono + timeOffset) / (1000 * 1000 * 1000)
	nsec = int32((mono + timeOffset) - sec*(1000*1000*1000))
	return
}

// AdjustTimeOffset adds the given offset to the built-in time offset. A
// positive value adds to the time (skipping some time), a negative value moves
// the clock into the past.
func AdjustTimeOffset(offset int64) {
	// TODO: do this atomically?
	timeOffset += offset
}
