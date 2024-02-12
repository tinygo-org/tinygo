//go:build wasm_unknown

package runtime

import (
	"unsafe"
)

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

// Read the command line arguments from WASI.
// For example, they can be passed to a program with wasmtime like this:
//
//	wasmtime run ./program.wasm arg1 arg2
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

const timePrecisionNanoseconds = 1000 // TODO: how can we determine the appropriate `precision`?

func sleepTicks(d timeUnit) {
}

func ticks() timeUnit {
	return timeUnit(0)
}
