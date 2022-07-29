//go:build scheduler.tasks
// +build scheduler.tasks

package runtime

import "internal/task"

// getSystemStackPointer returns the current stack pointer of the system stack.
// This is not necessarily the same as the current stack pointer.
func getSystemStackPointer() uintptr {
	// TODO: this always returns the correct stack on Cortex-M, so don't bother
	// comparing against 0.
	sp := task.SystemStack()
	if sp == 0 {
		sp = getCurrentStackPointer()
	}
	return sp
}

// Pause the current task for a given time.
//
//go:linkname sleep time.Sleep
func sleep(duration int64) {
	if duration <= 0 {
		return
	}

	addSleepTask(task.Current(), nanosecondsToTicks(duration))
	task.Pause()
}

// run is called by the program entry point to execute the go program.
// The main goroutine runs on the system stack, so initAll and callMain can be
// called directly.
func run() {
	initHeap()
	initAll()
	callMain()
	schedulerDone = true
}

const hasScheduler = true
