// +build scheduler.none

package task

import "unsafe"

//go:linkname runtimePanic runtime.runtimePanic
func runtimePanic(str string)

func Pause() {
	runtimePanic("scheduler is disabled")
}

func Current() *Task {
	runtimePanic("scheduler is disabled")
	return nil
}

//go:noinline
func start(fn uintptr, args unsafe.Pointer, stackSize uintptr) {
	// The compiler will error if this is reachable.
	runtimePanic("scheduler is disabled")
}

type state struct{}

func (t *Task) Resume() {
	runtimePanic("scheduler is disabled")
}

// OnSystemStack returns whether the caller is running on the system stack.
func OnSystemStack() bool {
	// This scheduler does not do any stack switching.
	return true
}
