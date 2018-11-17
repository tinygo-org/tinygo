// +build linux

package runtime

import (
	"unsafe"
)

func _Cfunc_putchar(c int) int
func _Cfunc_usleep(usec uint) int
func _Cfunc_malloc(size uintptr) unsafe.Pointer
func _Cfunc_abort()
func _Cfunc_clock_gettime(clk_id uint, ts *timespec)

const heapSize = 1 * 1024 * 1024 // 1MB to start

var (
	heapStart = uintptr(_Cfunc_malloc(heapSize))
	heapEnd   = heapStart + heapSize
)

type timeUnit int64

const tickMicros = 1

// TODO: Linux/amd64-specific
type timespec struct {
	tv_sec  int64
	tv_nsec int64
}

const CLOCK_MONOTONIC_RAW = 4

// Entry point for Go. Initialize all packages and call main.main().
//go:export main
func main() int {
	// Run initializers of all packages.
	initAll()

	// Compiler-generated wrapper to main.main().
	mainWrapper()

	// For libc compatibility.
	return 0
}

func putchar(c byte) {
	_Cfunc_putchar(int(c))
}

func sleepTicks(d timeUnit) {
	_Cfunc_usleep(uint(d) / 1000)
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
