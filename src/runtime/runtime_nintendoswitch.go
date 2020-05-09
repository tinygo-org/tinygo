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

	return exit(0) // Call libc_exit to cleanup libnx
}

// sleepTicks argument are actually in microseconds
func sleepTicks(d timeUnit) {
	usleep(uint(d))
}

// getArmSystemTimeNs returns ARM cpu ticks converted to nanoseconds
func getArmSystemTimeNs() timeUnit {
	t := getArmSystemTick()
	return timeUnit(ticksToNanoseconds(timeUnit(t)))
}

// armTicksToNs converts cpu ticks to nanoseconds
// Nintendo Switch CPU ticks has a fixed rate at 19200000
// It is basically 52 ns per tick
func ticksToNanoseconds(tick timeUnit) int64 {
	return int64(tick * 52)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 52)
}

func ticks() timeUnit {
	return timeUnit(getArmSystemTimeNs())
}

var stdoutBuffer = make([]byte, 0, 120)

func putchar(c byte) {
	if c == '\n' || len(stdoutBuffer)+1 >= 120 {
		svcOutputDebugString(&stdoutBuffer[0], uint64(len(stdoutBuffer)))
		stdoutBuffer = stdoutBuffer[:0]
		return
	}

	stdoutBuffer = append(stdoutBuffer, c)
}

//export usleep
func usleep(usec uint) int

//export abort
func abort() {
	exit(1)
}

//export exit
func exit(code int) int

//export armGetSystemTick
func getArmSystemTick() int64

// armGetSystemTickFreq returns the system tick frequency
// means how many ticks per second
//export armGetSystemTickFreq
func armGetSystemTickFreq() int64
