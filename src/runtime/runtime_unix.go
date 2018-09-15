// +build linux

package runtime

import (
	"unsafe"
)

func _Cfunc_putchar(c int) int
func _Cfunc_usleep(usec uint) int
func _Cfunc_calloc(nmemb, size uintptr) unsafe.Pointer
func _Cfunc_abort()
func _Cfunc_clock_gettime(clk_id uint, ts *timespec)

type timeUnit int64

const tickMicros = 1

// TODO: Linux/amd64-specific
type timespec struct {
	tv_sec  int64
	tv_nsec int64
}

const CLOCK_MONOTONIC_RAW = 4

func putchar(c byte) {
	_Cfunc_putchar(int(c))
}

//go:linkname sleep time.Sleep
func sleep(d int64) {
	_Cfunc_usleep(uint(d) / 1000)
}

func sleepTicks(d timeUnit) {
	sleep(int64(d))
}

// Return monotonic time in nanoseconds.
//
// TODO: noescape
func monotime() uint64 {
	ts := timespec{}
	_Cfunc_clock_gettime(CLOCK_MONOTONIC_RAW, &ts)
	return uint64(ts.tv_sec)*1000*1000*1000 + uint64(ts.tv_nsec)
}

func ticks() timeUnit {
	return timeUnit(monotime())
}

func abort() {
	// panic() exits with exit code 2.
	_Cfunc_abort()
}

func alloc(size uintptr) unsafe.Pointer {
	buf := _Cfunc_calloc(1, size)
	if buf == nil {
		runtimePanic("cannot allocate memory")
	}
	return buf
}

func free(ptr unsafe.Pointer) {
	//C.free(ptr) // TODO
}

func GC() {
	// Unimplemented.
}

func KeepAlive(x interface{}) {
	// Unimplemented. Only required with SetFinalizer().
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}
