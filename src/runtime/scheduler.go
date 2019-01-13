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
// Note that a goroutine is generally called a 'task' for brevity and because
// that's the more common term among RTOSes. But a goroutine and a task are
// basically the same thing. Although, the code often uses the word 'task' to
// refer to both a coroutine and a goroutine, as most of the scheduler doesn't
// care about the difference.
//
// For more background on coroutines in LLVM:
// https://llvm.org/docs/Coroutines.html

import (
	"unsafe"
)

const schedulerDebug = false

// A coroutine instance, wrapped here to provide some type safety. The value
// must not be used directly, it is meant to be used as an opaque *i8 in LLVM.
type coroutine uint8

//go:export llvm.coro.resume
func (t *coroutine) resume()

//go:export llvm.coro.destroy
func (t *coroutine) destroy()

//go:export llvm.coro.done
func (t *coroutine) done() bool

//go:export llvm.coro.promise
func (t *coroutine) _promise(alignment int32, from bool) unsafe.Pointer

// Get the promise belonging to a task.
func (t *coroutine) promise() *taskState {
	return (*taskState)(t._promise(int32(unsafe.Alignof(taskState{})), false))
}

func makeGoroutine(*uint8) *uint8

// State/promise of a task. Internally represented as:
//
//     {i8* next, i1 commaOk, i32/i64 data}
type taskState struct {
	next    *coroutine
	commaOk bool // 'comma-ok' flag for channel receive operation
	data    uint
}

// Queues used by the scheduler.
//
// TODO: runqueueFront can be removed by making the run queue a circular linked
// list. The runqueueBack will simply refer to the front in the 'next' pointer.
var (
	runqueueFront      *coroutine
	runqueueBack       *coroutine
	sleepQueue         *coroutine
	sleepQueueBaseTime timeUnit
)

// Simple logging, for debugging.
func scheduleLog(msg string) {
	if schedulerDebug {
		println(msg)
	}
}

// Simple logging with a task pointer, for debugging.
func scheduleLogTask(msg string, t *coroutine) {
	if schedulerDebug {
		println(msg, t)
	}
}

// Set the task to sleep for a given time.
//
// This is a compiler intrinsic.
func sleepTask(caller *coroutine, duration int64) {
	if schedulerDebug {
		println("  set sleep:", caller, uint(duration/tickMicros))
	}
	promise := caller.promise()
	promise.data = uint(duration / tickMicros) // TODO: longer durations
	addSleepTask(caller)
}

// Add a non-queued task to the run queue.
//
// This is a compiler intrinsic, and is called from a callee to reactivate the
// caller.
func activateTask(task *coroutine) {
	if task == nil {
		return
	}
	scheduleLogTask("  set runnable:", task)
	runqueuePushBack(task)
}

// Add this task to the end of the run queue. May also destroy the task if it's
// done.
func runqueuePushBack(t *coroutine) {
	if t.done() {
		scheduleLogTask("  destroy task:", t)
		t.destroy()
		return
	}
	if schedulerDebug {
		if t.promise().next != nil {
			panic("runtime: runqueuePushBack: expected next task to be nil")
		}
	}
	if runqueueBack == nil { // empty runqueue
		scheduleLogTask("  add to runqueue front:", t)
		runqueueBack = t
		runqueueFront = t
	} else {
		scheduleLogTask("  add to runqueue back:", t)
		lastTaskPromise := runqueueBack.promise()
		lastTaskPromise.next = t
		runqueueBack = t
	}
}

// Get a task from the front of the run queue. Returns nil if there is none.
func runqueuePopFront() *coroutine {
	t := runqueueFront
	if t == nil {
		return nil
	}
	if schedulerDebug {
		println("    runqueuePopFront:", t)
	}
	promise := t.promise()
	runqueueFront = promise.next
	if runqueueFront == nil {
		// Runqueue is empty now.
		runqueueBack = nil
	}
	promise.next = nil
	return t
}

// Add this task to the sleep queue, assuming its state is set to sleeping.
func addSleepTask(t *coroutine) {
	if schedulerDebug {
		if t.promise().next != nil {
			panic("runtime: addSleepTask: expected next task to be nil")
		}
	}
	now := ticks()
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
func scheduler() {
	// Main scheduler loop.
	for {
		scheduleLog("\n  schedule")
		now := ticks()

		// Add tasks that are done sleeping to the end of the runqueue so they
		// will be executed soon.
		if sleepQueue != nil && now-sleepQueueBaseTime >= timeUnit(sleepQueue.promise().data) {
			t := sleepQueue
			scheduleLogTask("  awake:", t)
			promise := t.promise()
			sleepQueueBaseTime += timeUnit(promise.data)
			sleepQueue = promise.next
			promise.next = nil
			runqueuePushBack(t)
		}

		t := runqueuePopFront()
		if t == nil {
			if sleepQueue == nil {
				// No more tasks to execute.
				// It would be nice if we could detect deadlocks here, because
				// there might still be functions waiting on each other in a
				// deadlock.
				scheduleLog("  no tasks left!")
				return
			}
			timeLeft := timeUnit(sleepQueue.promise().data) - (now - sleepQueueBaseTime)
			if schedulerDebug {
				println("  sleeping...", sleepQueue, uint(timeLeft))
			}
			sleepTicks(timeUnit(timeLeft))
			if asyncScheduler {
				// The sleepTicks function above only sets a timeout at which
				// point the scheduler will be called again. It does not really
				// sleep.
				break
			}
			continue
		}

		// Run the given task.
		scheduleLog("  <- runqueuePopFront")
		scheduleLogTask("  run:", t)
		t.resume()
	}
}
