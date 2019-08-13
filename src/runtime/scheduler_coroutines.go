// +build scheduler.coroutines

package runtime

// This file implements the Go scheduler using coroutines.
// A goroutine contains a whole stack. A coroutine is just a single function.
// How do we use coroutines for goroutines, then?
//   * Every function that contains a blocking call (like sleep) is marked
//     blocking, and all it's parents (callers) are marked blocking as well
//     transitively until the root (main.main or a go statement).
//   * A blocking function that calls a non-blocking function is called as
//     usual.
//   * A blocking function that calls a blocking function passes its own
//     coroutine handle as a parameter to the subroutine. When the subroutine
//     returns, it will re-insert the parent into the scheduler.
// Note that we use the type 'task' to refer to a coroutine, for compatibility
// with the task-based scheduler. A task type here does not represent the whole
// task, but just the topmost coroutine. For most of the scheduler, this
// difference doesn't matter.
//
// For more background on coroutines in LLVM:
// https://llvm.org/docs/Coroutines.html

import "unsafe"

// A coroutine instance, wrapped here to provide some type safety. The value
// must not be used directly, it is meant to be used as an opaque *i8 in LLVM.
type task uint8

//go:export llvm.coro.resume
func (t *task) resume()

//go:export llvm.coro.destroy
func (t *task) destroy()

//go:export llvm.coro.done
func (t *task) done() bool

//go:export llvm.coro.promise
func (t *task) _promise(alignment int32, from bool) unsafe.Pointer

// Get the state belonging to a task.
func (t *task) state() *taskState {
	return (*taskState)(t._promise(int32(unsafe.Alignof(taskState{})), false))
}

func makeGoroutine(uintptr) uintptr

// Compiler stub to get the current goroutine. Calls to this function are
// removed in the goroutine lowering pass.
func getCoroutine() *task

// getTaskStatePtr is a helper function to set the current .ptr field of a
// coroutine promise.
func setTaskStatePtr(t *task, value unsafe.Pointer) {
	t.state().ptr = value
}

// getTaskStatePtr is a helper function to get the current .ptr field from a
// coroutine promise.
func getTaskStatePtr(t *task) unsafe.Pointer {
	if t == nil {
		blockingPanic()
	}
	return t.state().ptr
}

//go:linkname sleep time.Sleep
func sleep(d int64) {
	sleepTicks(timeUnit(d / tickMicros))
}

// deadlock is called when a goroutine cannot proceed any more, but is in theory
// not exited (so deferred calls won't run). This can happen for example in code
// like this, that blocks forever:
//
//     select{}
//
// The coroutine version is implemented directly in the compiler but it needs
// this definition to work.
func deadlock()

// reactivateParent reactivates the parent goroutine. It is necessary in case of
// the coroutine-based scheduler.
func reactivateParent(t *task) {
	activateTask(t)
}

// chanYield exits the current goroutine. Used in the channel implementation, to
// suspend the current goroutine until it is reactivated by a channel operation
// of a different goroutine. It is a no-op in the coroutine implementation.
func chanYield() {
	// Nothing to do here, simply returning from the channel operation also exits
	// the goroutine temporarily.
}
