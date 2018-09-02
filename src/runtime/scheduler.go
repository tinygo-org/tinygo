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
//     coroutine handle as a parameter to the subroutine and will make sure it's
//     own coroutine is removed from the scheduler. When the subroutine returns,
//     it will re-insert the parent into the scheduler.
// Note that a goroutine is generally called a 'task' for brevity and because
// that's the more common term among RTOSes. But a goroutine and a task are
// basically the same thing. Although, the code often uses the word 'task' to
// refer to both a coroutine and a goroutine, as most of the scheduler isn't
// aware of the difference.
//
// For more background on coroutines in LLVM:
// https://llvm.org/docs/Coroutines.html

import (
	"unsafe"
)

// A coroutine instance, wrapped here to provide some type safety. The value
// must not be used directly, it is meant to be used as an opaque *i8 in LLVM.
type coroutine uint8

//go:linkname resume llvm.coro.resume
func (t *coroutine) resume()

//go:linkname destroy llvm.coro.destroy
func (t *coroutine) destroy()

//go:linkname done llvm.coro.done
func (t *coroutine) done() bool

//go:linkname _promise llvm.coro.promise
func (t *coroutine) _promise(alignment int32, from bool) unsafe.Pointer

// Get the promise belonging to a task.
func (t *coroutine) promise() *taskState {
	return (*taskState)(t._promise(4, false))
}

// State/promise of a task. Internally represented as:
//
//     {i8 state, i32 data, i8* next}
type taskState struct {
	state uint8
	data  uint32
	next  *coroutine
}

// Various states a task can be in.
const (
	TASK_STATE_RUNNABLE = iota
	TASK_STATE_SLEEP
	TASK_STATE_CALL // waiting for a sub-coroutine
)

// Queues used by the scheduler.
//
// TODO: runqueueFront can be removed by making the run queue a circular linked
// list. The runqueueBack will simply refer to the front in the 'next' pointer.
var (
	runqueueFront      *coroutine
	runqueueBack       *coroutine
	sleepQueue         *coroutine
	sleepQueueBaseTime uint64
)

// Simple logging, for debugging.
func scheduleLog(msg string) {
	//println(msg)
}

// Simple logging with a task pointer, for debugging.
func scheduleLogTask(msg string, t *coroutine) {
	//println(msg, t)
}

// Set the task state to sleep for a given time.
//
// This is a compiler intrinsic.
func sleepTask(caller *coroutine, duration Duration) {
	promise := caller.promise()
	promise.state = TASK_STATE_SLEEP
	promise.data = uint32(duration) // TODO: longer durations
}

// Wait for the result of an async call. This means that the parent goroutine
// will be removed from the runqueue and be rescheduled by the callee.
//
// This is a compiler intrinsic.
func waitForAsyncCall(caller *coroutine) {
	promise := caller.promise()
	promise.state = TASK_STATE_CALL
}

// Add a task to the runnable or sleep queue, depending on the state.
//
// This is a compiler intrinsic.
func scheduleTask(t *coroutine) {
	if t == nil {
		return
	}
	scheduleLogTask("  schedule task:", t)
	// See what we should do with this task: try to execute it directly
	// again or let it sleep for a bit.
	promise := t.promise()
	if promise.state == TASK_STATE_CALL {
		return // calling an async task, the subroutine will re-active the parent
	} else if promise.state == TASK_STATE_SLEEP && promise.data != 0 {
		addSleepTask(t)
	} else {
		pushTask(t)
	}
}

// Add this task to the end of the run queue. May also destroy the task if it's
// done.
func pushTask(t *coroutine) {
	if t.done() {
		scheduleLogTask("  destroy task:", t)
		t.destroy()
		return
	}
	if runqueueBack == nil { // empty runqueue
		runqueueBack = t
		runqueueFront = t
	} else {
		lastTaskPromise := runqueueBack.promise()
		lastTaskPromise.next = t
		runqueueBack = t
	}
}

// Get a task from the front of the run queue. May return nil if there is none.
func popTask() *coroutine {
	t := runqueueFront
	if t == nil {
		return nil
	}
	scheduleLogTask("    popTask:", t)
	promise := t.promise()
	runqueueFront = promise.next
	if runqueueFront == nil {
		runqueueBack = nil
	}
	promise.next = nil
	return t
}

// Add this task to the sleep queue, assuming its state is set to sleeping.
func addSleepTask(t *coroutine) {
	now := monotime()
	if sleepQueue == nil {
		scheduleLog("  -> sleep new queue")
		// Create new linked list for the sleep queue.
		sleepQueue = t
		sleepQueueBaseTime = now
		return
	}

	// Make sure promise.data is relative to the queue time base.
	promise := t.promise()

	// Insert at front of sleep queue.
	if promise.data < sleepQueue.promise().data {
		scheduleLog("  -> sleep at start")
		sleepQueue.promise().data -= promise.data
		promise.next = sleepQueue
		sleepQueue = t
		return
	}

	// Add to sleep queue (in the middle or at the end).
	queueIndex := sleepQueue
	for {
		promise.data -= queueIndex.promise().data
		if queueIndex.promise().next == nil || queueIndex.promise().data > promise.data {
			if queueIndex.promise().next == nil {
				scheduleLog("  -> sleep at end")
				promise.next = nil
			} else {
				scheduleLog("  -> sleep in middle")
				promise.next = queueIndex.promise().next
				promise.next.promise().data -= promise.data
			}
			queueIndex.promise().next = t
			break
		}
		queueIndex = queueIndex.promise().next
	}
}

// Run the scheduler until all tasks have finished.
// It takes an initial task (main.main) to bootstrap.
func scheduler(main *coroutine) {
	// Initial task.
	scheduleTask(main)

	// Main scheduler loop.
	for {
		scheduleLog("\n  schedule")
		now := monotime()

		// Add tasks that are done sleeping to the end of the runqueue so they
		// will be executed soon.
		if sleepQueue != nil && now-sleepQueueBaseTime >= uint64(sleepQueue.promise().data) {
			scheduleLog("  run <- sleep")
			t := sleepQueue
			promise := t.promise()
			sleepQueueBaseTime += uint64(promise.data)
			sleepQueue = promise.next
			promise.state = TASK_STATE_RUNNABLE
			promise.next = nil
			pushTask(t)
		}

		scheduleLog("  <- popTask")
		t := popTask()
		if t == nil {
			if sleepQueue == nil {
				// No more tasks to execute.
				// It would be nice if we could detect deadlocks here, because
				// there might still be functions waiting on each other in a
				// deadlock.
				scheduleLog("  no tasks left!")
				return
			}
			scheduleLog("  sleeping...")
			timeLeft := uint64(sleepQueue.promise().data) - (now - sleepQueueBaseTime)
			sleep(Duration(timeLeft))
			continue
		}

		// Run the given task.
		scheduleLogTask("  run:", t)
		t.resume()

		// Add the just resumed task to the run queue or the sleep queue.
		scheduleTask(t)
	}
}
