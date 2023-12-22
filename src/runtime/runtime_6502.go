//go:build 6502

package runtime

// timeUnit in nanoseconds
type timeUnit int64

func putchar(c byte) {
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

func ticks() timeUnit {
	return 0
}

func sleepTicks(d timeUnit) {
	return
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return 0
}

func nanosecondsToTicks(ns int64) timeUnit {
	return 0
}
func exit(code int) {
	abort()
}

func abort() {
	// TODO
	for {
	}
}
