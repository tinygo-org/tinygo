//go:build wasm && !wasi
// +build wasm,!wasi

package runtime

type timeUnit float64 // time in milliseconds, just like Date.now() in JavaScript

// wasmNested is used to detect scheduler nesting (WASM calls into JS calls back into WASM).
// When this happens, we need to use a reduced version of the scheduler.
var wasmNested bool

//export _start
func _start() {
	wasmNested = true
	run()
	wasmNested = false
}

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
//export runtime.sleepTicks
func sleepTicks(d timeUnit)

//export runtime.ticks
func ticks() timeUnit
