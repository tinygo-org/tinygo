// +build linux

package runtime

import (
	"unsafe"
)

const Microsecond = 1

func _Cfunc_putchar(c int) int
func _Cfunc_usleep(usec uint) int
func _Cfunc_calloc(nmemb, size uintptr) unsafe.Pointer
func _Cfunc_exit(status int)
func _Cfunc_clock_gettime(clk_id uint, ts *timespec)

// TODO: Linux/amd64-specific
type timespec struct {
	tv_sec  int64
	tv_nsec int64
}

const CLOCK_MONOTONIC_RAW = 4

func putchar(c byte) {
	_Cfunc_putchar(int(c))
}

func sleep(d Duration) {
	_Cfunc_usleep(uint(d))
}

// Return monotonic time in microseconds.
//
// TODO: use nanoseconds?
// TODO: noescape
func monotime() uint64 {
	ts := timespec{}
	_Cfunc_clock_gettime(CLOCK_MONOTONIC_RAW, &ts)
	return uint64(ts.tv_sec)*1000*1000 + uint64(ts.tv_nsec)/1000
}

func abort() {
	// panic() exits with exit code 2.
	_Cfunc_exit(2)
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
