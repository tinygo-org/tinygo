// +build wasm,wasi

package runtime

import "unsafe"

type timeUnit float64

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)

	run()
}

const asyncScheduler = true

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

// This function is called by the scheduler.
// Schedule a call to runtime.scheduler, do not actually sleep.
//export runtime.sleepTicks
func sleepTicks(d timeUnit)

//go:wasm-module wasi_unstable
//export clock_time_get
func clock_time_get(id uint32, precision uint64, timePtr *uint64) (errno uint)

func ticks() timeUnit {
	var time uint64
	clock_time_get(0, 100, &time)
	return timeUnit(time)
}
