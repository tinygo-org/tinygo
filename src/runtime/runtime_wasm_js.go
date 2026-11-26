package runtime

const GOOS = "js"

var handleEvent func()

//go:linkname setEventHandler syscall/js.setEventHandler
func setEventHandler(fn func()) {
	handleEvent = fn
}

// We use 1ns per tick, to simplify things.
// It would probably be fine to use 1µs per tick, since performance.now only
// promises a resolution of 5µs, but 1ns makes the conversions here a bit more
// straightforward (since nothing needs to be converted).
func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	wall := walltime()
	sec = wall / 1e9
	nsec = int32(wall % 1e9)
	mono = nanotime()
	return
}

//go:wasmimport gojs runtime.walltime
func walltime() int64

// This function is called by the scheduler.
// Schedule a call to runtime.scheduler, do not actually sleep.
//
//go:wasmimport gojs runtime.sleepTicks
func sleepTicks(d timeUnit)

//go:wasmimport gojs runtime.ticks
func ticks() timeUnit
