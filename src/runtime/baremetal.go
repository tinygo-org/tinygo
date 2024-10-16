//go:build baremetal

package runtime

import (
	"internal/itoa"
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
	// Note: this zeroes the returned buffer which is not necessary.
	// The same goes for bytealg.MakeNoZero.
	return alloc(size, nil)
}

//export calloc
func libc_calloc(nmemb, size uintptr) unsafe.Pointer {
	// No difference between calloc and malloc.
	return libc_malloc(nmemb * size)
}

//export free
func libc_free(ptr unsafe.Pointer) {
	free(ptr)
}

//export runtime_putchar
func runtime_putchar(c byte) {
	putchar(c)
}

// ErrExitPanic implements error
type ErrExitPanic int

func (e ErrExitPanic) Error() string {
	return "errno " + itoa.Itoa(int(e))
}

var panicOnExit bool

// PanicOnExit enables panic when syscall.Exit is called,
// as opposed to halting. It can be set by, e.g., a u-root
// shell to regain control when commands exit.
// This in turn makes it easier for building unmodified
// commands into u-root, for bare metal environments.
func PanicOnExit() {
	panicOnExit = true
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	if panicOnExit {
		panic(ErrExitPanic(code))
	}

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
