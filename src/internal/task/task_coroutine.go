// +build scheduler.coroutines

package task

import (
	"unsafe"
)

// rawState is an underlying coroutine state exposed by llvm.coro.
// This matches *i8 in LLVM.
type rawState uint8

//export llvm.coro.resume
func coroResume(*rawState)

type state struct{ *rawState }

//export llvm.coro.noop
func noopState() *rawState

// Resume the task until it pauses or completes.
func (t *Task) Resume() {
	coroResume(t.state.rawState)
}

// setState is used by the compiler to set the state of the function at the beginning of a function call.
// Returns the state of the caller.
func (t *Task) setState(s *rawState) *rawState {
	caller := t.state
	t.state = state{s}
	return caller.rawState
}

// returnTo is used by the compiler to return to the state of the caller.
func (t *Task) returnTo(parent *rawState) {
	t.state = state{parent}
	t.returnCurrent()
}

// returnCurrent is used by the compiler to return to the state of the caller in a case where the state is not replaced.
func (t *Task) returnCurrent() {
	scheduleTask(t)
}

//go:linkname scheduleTask runtime.runqueuePushBack
func scheduleTask(*Task)

// setReturnPtr is used by the compiler to store the return buffer into the task.
// This buffer is where the return value of a function that is about to be called will be stored.
func (t *Task) setReturnPtr(buf unsafe.Pointer) {
	t.Ptr = buf
}

// getReturnPtr is used by the compiler to get the return buffer stored into the task.
// This is called at the beginning of an async function, and the return is stored into this buffer immediately before resuming the caller.
func (t *Task) getReturnPtr() unsafe.Pointer {
	return t.Ptr
}

// createTask returns a new task struct initialized with a no-op state.
func createTask() *Task {
	return &Task{
		state: state{noopState()},
	}
}

// start invokes a function in a new goroutine. Calls to this are inserted by the compiler.
// The created goroutine starts running immediately.
// This is implemented inside the compiler.
func start(fn uintptr, args unsafe.Pointer, stackSize uintptr)

// Current returns the current active task.
// This is implemented inside the compiler.
func Current() *Task

// Pause suspends the current running task.
// This is implemented inside the compiler.
func Pause()

func fake() {
	// Hack to ensure intrinsics are discovered.
	Current()
	Pause()
}

// OnSystemStack returns whether the caller is running on the system stack.
func OnSystemStack() bool {
	// This scheduler does not do any stack switching.
	return true
}
