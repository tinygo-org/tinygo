// +build linux

package runtime

import (
	"unsafe"
)

//go:export putchar
func _putchar(c int) int

//go:export usleep
func usleep(usec uint) int

//go:export malloc
func malloc(size uintptr) unsafe.Pointer

//go:export abort
func abort()

//go:export clock_gettime
func clock_gettime(clk_id uint, ts *timespec)

const heapSize = 1 * 1024 * 1024 // 1MB to start

var (
	heapStart = uintptr(malloc(heapSize))
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

	// Compiler-generated call to main.main().
	callMain()

	// For libc compatibility.
	return 0
}

func putchar(c byte) {
	_putchar(int(c))
}

const asyncScheduler = false

func sleepTicks(d timeUnit) {
	usleep(uint(d) / 1000)
}

// Return monotonic time in nanoseconds.
//
// TODO: noescape
func monotime() uint64 {
	ts := timespec{}
	clock_gettime(CLOCK_MONOTONIC_RAW, &ts)
	return uint64(ts.tv_sec)*1000*1000*1000 + uint64(ts.tv_nsec)
}

func ticks() timeUnit {
	return timeUnit(monotime())
}
