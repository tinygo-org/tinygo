// +build nintendoswitch

package runtime

type timeUnit int64

const asyncScheduler = false

func postinit() {}

// Entry point for Go. Initialize all packages and call main.main().
//export main
func main() int {
	preinit()
	run()

	// Call exit to correctly finish the program
	// Without this, the application crashes at start, not sure why
	return exit(0)
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

var stdoutBuffer = make([]byte, 0, 120)

func putchar(c byte) {
	if c == '\n' || len(stdoutBuffer)+1 >= 120 {
		NxOutputString(string(stdoutBuffer))
		stdoutBuffer = stdoutBuffer[:0]
		return
	}

	stdoutBuffer = append(stdoutBuffer, c)
}

func usleep(usec uint) int {
	sleepThread(uint64(usec) * 1000)
	return 0
}

func abort() {
	for {
		exit(1)
	}
}

//export sleepThread
func sleepThread(nanos uint64)

//export exit
func exit(code int) int

//export armGetSystemTick
func getArmSystemTick() int64
