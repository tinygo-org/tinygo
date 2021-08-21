// +build wasm,!cosmwasm,!wasi

package runtime

import "unsafe"

type timeUnit float64 // time in milliseconds, just like Date.now() in JavaScript

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)

	run()
}

var handleEvent func()

//go:linkname setEventHandler syscall/js.setEventHandler
func setEventHandler(fn func()) {
	handleEvent = fn
}

//export resume
func resume() {
	go func() {
		handleEvent()
	}()
	scheduler()
}

//export go_scheduler
func go_scheduler() {
	scheduler()
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

// placeholder
func putchar(c byte) {}
