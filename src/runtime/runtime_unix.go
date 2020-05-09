// +build darwin linux,!baremetal freebsd,!baremetal
// +build !nintendoswitch

package runtime

import (
	"unsafe"
)

//export putchar
func _putchar(c int) int

//export usleep
func usleep(usec uint) int

//export malloc
func malloc(size uintptr) unsafe.Pointer

//export abort
func abort()

//export exit
func exit(code int)

//export clock_gettime
func clock_gettime(clk_id int32, ts *timespec)

type timeUnit int64

// Note: tv_sec and tv_nsec vary in size by platform. They are 32-bit on 32-bit
// systems and 64-bit on 64-bit systems (at least on macOS/Linux), so we can
// simply use the 'int' type which does the same.
type timespec struct {
	tv_sec  int // time_t: follows the platform bitness
	tv_nsec int // long: on Linux and macOS, follows the platform bitness
}

const CLOCK_MONOTONIC_RAW = 4

func postinit() {}

// Entry point for Go. Initialize all packages and call main.main().
//export main
func main() int {
	preinit()

	run()

	// For libc compatibility.
	return 0
}

func putchar(c byte) {
	_putchar(int(c))
}

const asyncScheduler = false

func ticksToNanoseconds(ticks timeUnit) int64 {
	// The OS API works in nanoseconds so no conversion necessary.
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	// The OS API works in nanoseconds so no conversion necessary.
	return timeUnit(ns)
}

func sleepTicks(d timeUnit) {
	// timeUnit is in nanoseconds, so need to convert to microseconds here.
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

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	exit(code)
}

func extalloc(size uintptr) unsafe.Pointer {
	return malloc(size)
}

//export free
func extfree(ptr unsafe.Pointer)

// TinyGo does not yet support any form of parallelism on an OS, so these can be
// left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}
