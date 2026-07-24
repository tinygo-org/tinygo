//go:build tinygo.unwind.asyncify

package runtime

import "internal/task"

//go:wasmimport asyncify stop_unwind
func asyncifyStopUnwindImport()

func startUnwind(frame *deferFrame) bool {
	frame.PanicState |= panicUnwinding
	setUnwindSignal(true)
	task.PanicUnwind()
	return true
}
