//go:build baremetal

package runtime

import (
	"unsafe"
)

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
