//go:build tinygo.unwind.asyncify

package runtime

import (
	"internal/task"
	"unsafe"
)

//go:wasmimport asyncify stop_unwind
func asyncifyStopUnwindImport()

//go:wasmimport asyncify start_rewind
func asyncifyStartRewindImport(state unsafe.Pointer)

func startUnwind(frame *deferFrame) bool {
	frame.PanicState |= panicUnwinding
	setUnwindSignal(true)
	task.PanicUnwind()
	return true
}

// Keep this bookkeeping out of the uninstrumented panic catcher so ordinary
// scheduler unwinds do not retain its stack frame.
//
//go:noinline
func savePanicReplay(replay, target uintptr) {
	task.StopPanicUnwind(replay, target)
}

func clearPanicReplay() {
	task.ClearPanicReplay()
}

func panicRewindData() unsafe.Pointer {
	return task.PanicRewindData()
}

func panicRewindStackPointer() unsafe.Pointer {
	return task.PanicRewindStackPointer()
}

//go:linkname setPanicRewindStackPointer tinygo_set_panic_rewind_stack_pointer
func setPanicRewindStackPointer(stackPointer unsafe.Pointer)

func rewindPanic() bool {
	return task.RewindPanic()
}
