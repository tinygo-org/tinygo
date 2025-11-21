//go:build py32

package runtime

//export Reset_Handler
func main() {
	// initSystem()
	// arm.Asm("CPSIE i")
	// initInternal()

	run()
	exit(0)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return 0
}

func nanosecondsToTicks(ns int64) timeUnit {
	return 0
}

//go:linkname ticks runtime.ticks
func ticks() timeUnit {
	return 0
}

func sleepTicks(d timeUnit) {
}

func putchar(c byte) {
}
