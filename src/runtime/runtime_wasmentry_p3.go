//go:build wasip3

package runtime

import "internal/task"

// A task that runs _initialize cannot block, but package initializers may
// sleep or wait for goroutines. Run them from wasi:cli/run instead.
const initInRun = true

var initAllDone bool

// initAllOnce initializes the task system and runs the package initializers
// at the start of the first exported function. That is wasi:cli/run for a
// command, and for any other world it is the first export that is called.
func initAllOnce() {
	if initAllDone {
		return
	}
	initAllDone = true
	task.Init(getCurrentStackPointer())
	initAll()
}

func wasmExportInit() {
	initAllOnce()
}
