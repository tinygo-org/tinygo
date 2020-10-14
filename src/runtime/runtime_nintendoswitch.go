// +build nintendoswitch

package runtime

import "unsafe"

type timeUnit int64

const asyncScheduler = false

var stackTop uintptr

func postinit() {}

// Entry point for Go. Initialize all packages and call main.main().
//export main
func main() int {
	preinit()

	// Obtain the initial stack pointer right before calling the run() function.
	// The run function has been moved to a separate (non-inlined) function so
	// that the correct stack pointer is read.
	stackTop = getCurrentStackPointer()
	runMain()

	// Call exit to correctly finish the program
	// Without this, the application crashes at start, not sure why
	return exit(0)
}

// Must be a separate function to get the correct stack pointer.
//go:noinline
func runMain() {
	run()
}

// sleepTicks sleeps for the specified system ticks
func sleepTicks(d timeUnit) {
	sleepThread(uint64(ticksToNanoseconds(d)))
}

// armTicksToNs converts cpu ticks to nanoseconds
// Nintendo Switch CPU ticks has a fixed rate at 19200000
// It is basically 52 ns per tick
// The formula 625 / 12 is equivalent to 1e9 / 19200000
func ticksToNanoseconds(tick timeUnit) int64 {
	return int64(tick * 625 / 12)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(12 * ns / 625)
}

func ticks() timeUnit {
	return timeUnit(ticksToNanoseconds(timeUnit(getArmSystemTick())))
}

var stdoutBuffer = make([]byte, 120)
var position = 0

func nativePutchar(c byte) {
	if c == '\n' || position >= len(stdoutBuffer) {
		nxOutputString(&stdoutBuffer[0], uint64(position))
		position = 0
		return
	}

	stdoutBuffer[position] = c
	position++
}

func abort() {
	for {
		exit(1)
	}
}

//export write
func write(fd int32, buf *byte, count int) int {
	// TODO: Proper handling write
	for i := 0; i < count; i++ {
		putchar(*buf)
		buf = (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(buf)) + 1))
	}
	return count
}

//export sleepThread
func sleepThread(nanos uint64)

//export exit
func exit(code int) int

//export armGetSystemTick
func getArmSystemTick() int64
