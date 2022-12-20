//go:build scheduler.tasks

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
