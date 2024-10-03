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
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	wasmExportState = wasmExportStateInMain
	run()
	wasmExportState = wasmExportStateExited
	beforeExit()
}

// This is the _initialize entry point, when using -buildmode=c-shared.
func wasmEntryReactor() {
	// This function is called before any //go:wasmexport functions are called
	// to initialize everything. It must not block.

	// Initialize the heap.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	initHeap()

	if hasScheduler {
		// A package initializer might do funky stuff like start a goroutine and
		// wait until it completes, so we have to run package initializers in a
		// goroutine.
		go func() {
			initAll()
			wasmExportState = wasmExportStateReactor
		}()
		scheduler(true)
		if wasmExportState != wasmExportStateReactor {
			// Unlikely, but if package initializers do something blocking (like
			// time.Sleep()), that's a bug.
			runtimePanic("package initializer blocks")
		}
	} else {
		// There are no goroutines (except for the main one, if you can call it
		// that), so we can just run all the package initializers.
		initAll()
		wasmExportState = wasmExportStateReactor
	}
}

// Track which state we're in: before (or during) init, running inside
// main.main, after main.main returned, or reactor mode (after init).
var wasmExportState uint8

const (
	wasmExportStateInit = iota
	wasmExportStateInMain
	wasmExportStateExited
	wasmExportStateReactor
)

func wasmExportCheckRun() {
	switch wasmExportState {
	case wasmExportStateInit:
		runtimePanic("//go:wasmexport function called before runtime initialization")
	case wasmExportStateExited:
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
	task.Pause()

	// TODO: we could cache the allocated stack so we don't have to keep
	// allocating a new stack on every //go:wasmexport call.
}
