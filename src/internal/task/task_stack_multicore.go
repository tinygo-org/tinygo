//go:build scheduler.cores

package task

import "runtime/interrupt"

// Current returns the current active task.
//
//go:linkname Current runtime.currentTask
func Current() *Task

// Pause suspends the current task and returns to the scheduler.
// This function may only be called when running on a goroutine stack, not when running on the system stack or in an interrupt.
func Pause() {
	lockScheduler()
	PauseLocked()
}

// PauseLocked is the same as Pause, but must be called with the scheduler lock
// already taken.
func PauseLocked() {
	// Check whether the canary (the lowest address of the stack) is still
	// valid. If it is not, a stack overflow has occurred.
	current := Current()
	if *current.state.canaryPtr != stackCanary {
		runtimePanic("goroutine stack overflow")
	}
	if interrupt.In() {
		runtimePanic("blocked inside interrupt")
	}
	if current.RunState == RunStateResuming {
		// Another core already marked this goroutine as ready to resume.
		current.RunState = RunStateRunning
		unlockScheduler()
		return
	}
	current.RunState = RunStatePaused
	current.state.pause()
}

// Resume the task until it pauses or completes.
// This may only be called from the scheduler.
func (t *Task) Resume() {
	t.gcData.swap()
	t.state.resume()
	t.gcData.swap()
}

//go:linkname lockScheduler runtime.lockScheduler
func lockScheduler()

//go:linkname unlockScheduler runtime.unlockScheduler
func unlockScheduler()
