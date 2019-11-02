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
	runqueueFront *task
	runqueueBack  *task
	nextTasks     *task
)

// variable set to true after main returns, to indicate that the scheduler should exit
var schedulerDone bool

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

// deadlock is called when a goroutine cannot proceed any more, but is in theory
// not exited (so deferred calls won't run). This can happen for example in code
// like this, that blocks forever:
//
//     select{}
//go:noinline
func deadlock() {
	// call yield without requesting a wakeup
	yield()
	panic("unreachable")
}

// Goexit terminates the currently running goroutine. No other goroutines are affected.
//
// Unlike the main Go implementation, no deffered calls will be run.
//go:inline
func Goexit() {
	// its really just a deadlock
	deadlock()
}

// unblock unblocks a task and returns the next value
func unblock(t *task) *task {
	state := t.state()
	next := state.next
	state.next = nil
	activateTask(t)
	return next
}

// unblockChain unblocks the next task on the stack/queue, returning it
// also updates the chain, putting the next element into the chain pointer
// if the chain is used as a queue, tail is used as a pointer to the final insertion point
// if the chain is used as a stack, tail should be nil
func unblockChain(chain **task, tail ***task) *task {
	t := *chain
	if t == nil {
		return nil
	}
	*chain = unblock(t)
	if tail != nil && *chain == nil {
		*tail = chain
	}
	return t
}

// dropChain drops a task from the given stack or queue
// if the chain is used as a queue, tail is used as a pointer to the field containing a pointer to the next insertion point
// if the chain is used as a stack, tail should be nil
func dropChain(t *task, chain **task, tail ***task) {
	for c := chain; *c != nil; c = &((*c).state().next) {
		if *c == t {
			next := (*c).state().next
			if next == nil && tail != nil {
				*tail = c
			}
			*c = next
			return
		}
	}
	panic("runtime: task not in chain")
}

// Pause the current task for a given time.
//go:linkname sleep time.Sleep
func sleep(duration int64) {
	addSleepTask(getCoroutine(), duration)
	yield()
}

func avrSleep(duration int64) {
	sleepTicks(timeUnit(duration / tickMicros))
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
//go:inline
func getTaskStateData(t *task) uint {
	return t.state().data
}

// Add this task to the end of the run queue. May also destroy the task if it's
// done.
func runqueuePushBack(t *task) {
	if schedulerDebug {
		scheduleLogTask("  pushing back:", t)
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

// Run the scheduler until all tasks have finished.
func scheduler() {
	// Main scheduler loop.
	for {
		if schedulerDone {
			scheduleLog("  done")
			return
		}

		scheduleLog("")
		scheduleLog("  schedule")

		// run active tasks
		for nextTasks != nil {
			t := nextTasks
			nextTasks, t.state().next = t.state().next, nil
			scheduleLogTask("  run:", t)
			t.resume()
		}

		// check for events
		evs := poll()
		for evs != nil {
			t := evs
			evs = t.state().next
			t.state().next = nil
			scheduleLogTask("  resumed by event:", t)
			runqueuePushBack(t)
		}

		// copy runqueue to nextTasks
		nextTasks = runqueueFront
		runqueueFront, runqueueBack = nil, nil

		if schedulerDone {
			// main returned, immediately exit
			scheduleLog("  done")
			return
		}

		if nextTasks == nil {
			// wait for something to do
			wait()
			if asyncScheduler {
				return
			}
		}
	}
}

func Gosched() {
	runqueuePushBack(getCoroutine())
	yield()
}
