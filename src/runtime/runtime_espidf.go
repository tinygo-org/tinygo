//go:build espidf

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

func growHeap() bool {
	return false
}

//export abort
func abort()

//export exit
func exit(code int)

//export putchar
func libc_putchar(c byte)

func putchar(c byte) {
	libc_putchar(c)
}

//export getchar
func libc_getchar() byte

func getchar() byte {
	return libc_getchar()
}

func buffered() int {
	return 0
}

type timespec struct {
	tv_sec  int64
	tv_nsec int32
}

//export clock_gettime
func clock_gettime(clk_id int32, ts *timespec)

func getTime(clock int32) uint64 {
	ts := timespec{}
	clock_gettime(clock, &ts)
	return uint64(ts.tv_sec)*1000*1000*1000 + uint64(ts.tv_nsec)
}

const clock_MONOTONIC = 1

func monotime() uint64 {
	return getTime(clock_MONOTONIC)
}

func ticks() timeUnit {
	return timeUnit(monotime())
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

//export nanosleep
func nanosleep(req, rem *timespec) int

func sleepTicks(d timeUnit) {
	nanosleep(&timespec{int64(d / 1e9), int32(d % 1e9)}, nil)
}

const baremetal = true

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	mono = nanotime()
	sec = mono / 1e9
	nsec = int32(mono % 1e9)
	return
}

// Picolibc is not configured to define its own errno value, instead it calls
// __errno_location.
// TODO: a global works well enough for now (same as errno on Linux with
// -scheduler=tasks), but this should ideally be a thread-local variable stored
// in task.Task.
// Especially when we add multicore support for microcontrollers.
var errno int32

//export __errno_location
func libc_errno_location() *int32 {
	return &errno
}
