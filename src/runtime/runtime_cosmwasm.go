// +build wasm,cosmwasm

package runtime

import "unsafe"

type timeUnit float64 // time in milliseconds, just like Date.now() in JavaScript

const tickMicros = 1000000

func postinit() {}

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)

	run()
}

// putchar() is a no-op. Maybe we add an implementation later.
// this is needed to link various debug print statements in tinygo
func putchar(c byte) {}

const asyncScheduler = true

// This function is called by the scheduler.
// Schedule a call to runtime.scheduler, do not actually sleep.
//export runtime.sleepTicks
func sleepTicks(d timeUnit)

//export runtime.ticks
func ticks() timeUnit

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}
