//go:build wasm_unknown

package runtime

import "unsafe"

type timeUnit int64

// libc constructors
//
//export __wasm_call_ctors
func __wasm_call_ctors()

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	run()
}

func init() {
	__wasm_call_ctors()
}

var args []string

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

// with the wasm32-unknown-unknown target there is no way to determine any `precision`
const timePrecisionNanoseconds = 1000

func sleepTicks(d timeUnit) {
}

func ticks() timeUnit {
	return timeUnit(0)
}
