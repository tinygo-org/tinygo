package runtime

// This file implements the TinyGo scheduler. This scheduler is a very simple
// cooperative round robin scheduler, with a runqueue that contains a linked
// list of goroutines (tasks) that should be run next, in order of when they
// were added to the queue (first-in, first-out). It also contains a sleep queue
// with sleeping goroutines in order of when they should be re-activated.
//
// The scheduler is used both for the coroutine based scheduler and for the task
// based scheduler (see compiler/goroutine-lowering.go for a description). In
// both cases, the 'task' type is used to represent one goroutine. In the case
// of the task based scheduler, it literally is the goroutine itself: a pointer
// to the bottom of the stack where some important fields are kept. In the case
// of the coroutine-based scheduler, it is the coroutine pointer (a *i8 in
// LLVM).

import "unsafe"

const schedulerDebug = false

// State of a task. Internally represented as:
//
//     {i8* next, i8* ptr, i32/i64 data}
type taskState struct {
	next *task
	ptr  unsafe.Pointer
	data uint
}

// Queues used by the scheduler.
//
// TODO: runqueueFront can be removed by making the run queue a circular linked
// list. The runqueueBack will simply refer to the front in the 'next' pointer.
var (
	runqueueFront      *task
	runqueueBack       *task
	sleepQueue         *task
	sleepQueueBaseTime timeUnit
)

// Simple logging, for debugging.
func scheduleLog(msg string) {
	if schedulerDebug {
		println("---", msg)
	}
}

// Simple logging with a task pointer, for debugging.
func scheduleLogTask(msg string, t *task) {
	if schedulerDebug {
		println("---", msg, t)
	}
}

// Simple logging with a channel and task pointer.
func scheduleLogChan(msg string, ch *channel, t *task) {
	if schedulerDebug {
		println("---", msg, ch, t)
	}
}

// Set the task to sleep for a given time.
//
// This is a compiler intrinsic.
func sleepTask(caller *task, duration int64) {
	if schedulerDebug {
		println("  set sleep:", caller, uint(duration/tickMicros))
	}
	state := caller.state()
	state.data = uint(duration / tickMicros) // TODO: longer durations
	addSleepTask(caller)
}

// Add a non-queued task to the run queue.
//
// This is a compiler intrinsic, and is called from a callee to reactivate the
// caller.
func activateTask(t *task) {
	if t == nil {
		return
	}
	scheduleLogTask("  set runnable:", t)
	runqueuePushBack(t)
}

// getTaskStateData is a helper function to get the current .data field of the
// goroutine state.
func getTaskStateData(t *task) uint {
	return t.state().data
}

// Add this task to the end of the run queue. May also destroy the task if it's
// done.
func runqueuePushBack(t *task) {
	if schedulerDebug {
		if t.state().next != nil {
			panic("runtime: runqueuePushBack: expected next task to be nil")
		}
	}
	if runqueueBack == nil { // empty runqueue
		runqueueBack = t
		runqueueFront = t
	} else {
		lastTaskState := runqueueBack.state()
		lastTaskState.next = t
		runqueueBack = t
	}
}

// Get a task from the front of the run queue. Returns nil if there is none.
func runqueuePopFront() *task {
	t := runqueueFront
	if t == nil {
		return nil
	}
	state := t.state()
	runqueueFront = state.next
	if runqueueFront == nil {
		// Runqueue is empty now.
		runqueueBack = nil
	}
	state.next = nil
	return t
}

// Add this task to the sleep queue, assuming its state is set to sleeping.
func addSleepTask(t *task) {
	if schedulerDebug {
		if t.state().next != nil {
			panic("runtime: addSleepTask: expected next task to be nil")
		}
	}
	now := ticks()
	if sleepQueue == nil {
		scheduleLog("  -> sleep new queue")

		// set new base time
		sleepQueueBaseTime = now
	}

	// Add to sleep queue.
	q := &sleepQueue
	for ; *q != nil; q = &((*q).state()).next {
		if t.state().data < (*q).state().data {
			// this will finish earlier than the next - insert here
			break
		} else {
			// this will finish later - adjust delay
			t.state().data -= (*q).state().data
		}
	}
	if *q != nil {
		// cut delay time between this sleep task and the next
		(*q).state().data -= t.state().data
	}
	t.state().next = *q
	*q = t
}

// Run the scheduler until all tasks have finished.
func scheduler() {
	// Main scheduler loop.
	for {
		scheduleLog("")
		scheduleLog("  schedule")
		now := ticks()

		// Add tasks that are done sleeping to the end of the runqueue so they
		// will be executed soon.
		if sleepQueue != nil && now-sleepQueueBaseTime >= timeUnit(sleepQueue.state().data) {
			t := sleepQueue
			scheduleLogTask("  awake:", t)
			state := t.state()
			sleepQueueBaseTime += timeUnit(state.data)
			sleepQueue = state.next
			state.next = nil
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
			timeLeft := timeUnit(sleepQueue.state().data) - (now - sleepQueueBaseTime)
			if schedulerDebug {
				println("  sleeping...", sleepQueue, uint(timeLeft))
				for t := sleepQueue; t != nil; t = t.state().next {
					println("    task sleeping:", t, timeUnit(t.state().data))
				}
			}
			sleepTicks(timeLeft)
			if asyncScheduler {
				// The sleepTicks function above only sets a timeout at which
				// point the scheduler will be called again. It does not really
				// sleep.
				break
			}
			continue
		}

		// Run the given task.
		scheduleLogTask("  run:", t)
		t.resume()
	}
}
