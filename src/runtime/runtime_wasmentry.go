//go:build tinygo.wasm

package runtime

// Entry points for WebAssembly modules, and runtime support for
// //go:wasmexport: runtime.wasmExport* function calls are inserted by the
// compiler for //go:wasmexport support.

import (
	"internal/task"
	"unsafe"
)

// This is the _start entry point, when using -buildmode=default.
func wasmEntryCommand() {
	// These need to be initialized early so that the heap can be initialized.
	initializeCalled = true
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	run()
	if mainExited {
		// To make sure wasm_exec.js knows that we've exited, exit explicitly.
		mainReturnExit()
	}
}

// This is the _initialize entry point, when using -buildmode=c-shared.
func wasmEntryReactor() {
	// This function is called before any //go:wasmexport functions are called
	// to initialize everything. It must not block.

	initializeCalled = true

	// Initialize the heap.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	initRand()
	initHeap()

	if hasScheduler {
		// A package initializer might do funky stuff like start a goroutine and
		// wait until it completes, so we have to run package initializers in a
		// goroutine.
		go func() {
			initAll()
		}()
		scheduler(true)
	} else {
		// There are no goroutines (except for the main one, if you can call it
		// that), so we can just run all the package initializers.
		initAll()
	}
}

// This is the _start entry point, when using -buildmode=wasi-legacy.
func wasmEntryLegacy() {
	// These need to be initialized early so that the heap can be initialized.
	initializeCalled = true
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	run()
}

// Whether the runtime was initialized by a call to _initialize or _start.
var initializeCalled bool

func wasmExportCheckRun() {
	switch {
	case !initializeCalled:
		runtimePanic("//go:wasmexport function called before runtime initialization")
	case mainExited:
		runtimePanic("//go:wasmexport function called after main.main returned")
	}
}

// Called from within a //go:wasmexport wrapper (the one that's exported from
// the wasm module) after the goroutine has been queued. Just run the scheduler,
// and check that the goroutine finished when the scheduler is idle (as required
// by the //go:wasmexport proposal).
//
// This function is not called when the scheduler is disabled.
func wasmExportRun(done *bool) {
	scheduler(true)
	if !*done {
		runtimePanic("//go:wasmexport function did not finish")
	}
}

// Called from the goroutine wrapper for the //go:wasmexport function. It just
// signals to the runtime that the //go:wasmexport call has finished, and can
// switch back to the wasmExportRun function.
//
// This function is not called when the scheduler is disabled.
func wasmExportExit() {
	// Signal to the scheduler that it should return, since this call to a
	// //go:wasmexport function has exited.
	schedulerExit = true

	task.Pause()

	// TODO: we could cache the allocated stack so we don't have to keep
	// allocating a new stack on every //go:wasmexport call.
}
