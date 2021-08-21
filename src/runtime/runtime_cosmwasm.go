// +build tinygo.wasm,cosmwasm

package runtime

import "unsafe"

type timeUnit int64 // time in milliseconds, just like Date.now() in JavaScript

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)

	run()
}

// This function is called by the scheduler.
// Schedule a call to runtime.scheduler, do not actually sleep.
//export runtime.sleepTicks
func sleepTicks(d timeUnit) {
	// placeholder, should never be used
}

//export runtime.ticks
func ticks() timeUnit {
	// placeholder, should never be used
	return 1234567890
}
