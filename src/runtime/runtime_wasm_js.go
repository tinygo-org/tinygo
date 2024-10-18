//go:build wasm && !wasip1

package runtime

type timeUnit float64 // time in milliseconds, just like Date.now() in JavaScript

var handleEvent func()

//go:linkname setEventHandler syscall/js.setEventHandler
func setEventHandler(fn func()) {
	handleEvent = fn
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	// The JavaScript API works in float64 milliseconds, so convert to
	// nanoseconds first before converting to a timeUnit (which is a float64),
	// to avoid precision loss.
	return int64(ticks * 1e6)
}

func nanosecondsToTicks(ns int64) timeUnit {
	// The JavaScript API works in float64 milliseconds, so convert to timeUnit
	// (which is a float64) first before dividing, to avoid precision loss.
	return timeUnit(ns) / 1e6
}

// This function is called by the scheduler.
// Schedule a call to runtime.scheduler, do not actually sleep.
//
//go:wasmimport gojs runtime.sleepTicks
func sleepTicks(d timeUnit)

//go:wasmimport gojs runtime.ticks
func ticks() timeUnit

func beforeExit() {
	__stdio_exit()
}
