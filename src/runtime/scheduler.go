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

// variable set to true after main returns, to indicate that the scheduler should exit
var schedDone bool

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

// blockTask is a helper for task resume stacks which can be accessed via linkname for use in sync primitives
func blockTask(t *task, chain **task) {
	*chain, t.state().next = t, *chain
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

// Add this task to the sleep queue, assuming its state is set to sleeping.
func addSleepTask(t *task, duration int64) {
	if schedulerDebug {
		println("  set sleep:", t, uint(duration/tickMicros))
		if t.state().next != nil {
			panic("runtime: addSleepTask: expected next task to be nil")
		}
	}
	t.state().data = uint(duration / tickMicros) // TODO: longer durations
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
	var now timeUnit
	var intCycle uint8
	for {
		if schedDone {
			scheduleLog("  done")
			return
		}

		scheduleLog("")
		scheduleLog("  schedule")
		if sleepQueue != nil {
			now = ticks()
		}

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
		switch {
		case t != nil:
			intCycle = 0
			// Run the given task.
			scheduleLogTask("  run:", t)
			t.resume()
		case intCycle < 2:
			intCycle++

			// get interrupt wakeup chain
			head := popInt()

			if head == nil {
				// no interrupts found, skip
				break
			}

			// find tail of interrupt chain
			tail := head
			for ; tail.state().next != nil; tail = tail.state().next {}

			// glue interrupt chain into scheduler queue
			runqueueFront, runqueueBack = head, tail
		case sleepQueue != nil:
			timeLeft := timeUnit(sleepQueue.state().data) - (now - sleepQueueBaseTime)
			if schedulerDebug {
				println("  sleeping...", sleepQueue, uint(timeLeft))
				for t := sleepQueue; t != nil; t = t.state().next {
					println("    task sleeping:", t, timeUnit(t.state().data))
				}
			}
			intCycle = 0
			sleepTicks(timeLeft)
			if asyncScheduler {
				// The sleepTicks function above only sets a timeout at which
				// point the scheduler will be called again. It does not really
				// sleep.
				return
			}
		default:
			// No more tasks to execute.
			// It would be nice if we could detect deadlocks here, because
			// there might still be functions waiting on each other in a
			// deadlock.
			// Now, due to the addition of interrupts, we have to sleep here.
			scheduleLog("  no tasks left!")
			sleepTicks(0)
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
