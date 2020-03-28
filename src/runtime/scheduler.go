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

import (
	"internal/task"
)

const schedulerDebug = false

// runqueue is a queue of tasks that are ready to run.
var runqueue task.Queue

// Simple logging, for debugging.
func scheduleLog(msg string) {
	if schedulerDebug {
		println("---", msg)
	}
}

// Simple logging with a task pointer, for debugging.
func scheduleLogTask(msg string, t *task.Task) {
	if schedulerDebug {
		println("---", msg, t)
	}
}

// Simple logging with a channel and task pointer.
func scheduleLogChan(msg string, ch *channel, t *task.Task) {
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
	task.Pause()
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

// Add this task to the end of the run queue.
func runqueuePushBack(t *task.Task) {
	runqueue.Push(t)
}

var schedulerDone bool

const pollInterval = 256

// Run the scheduler until all tasks have finished.
func scheduler() {
	var n uint
	for !schedulerDone {
		// Get the next available task.
		t := runqueue.Pop()
		if t == nil {
			scheduleLog("  runqueue empty")

			// Check for any available tasks.
			if poll() {
				// A task was found and pushed onto the runqueue.
				continue
			}

			// Sleep until another task is available.
			wait()
			if asyncScheduler {
				// This platform (WebAssembly) requires us to return control to the host while waiting.
				// The host will eventually re-invoke the scheduler when there is work available.
				return
			}

			// Try again.
			continue
		}

		// Run task.
		scheduleLogTask("resuming:", t)
		t.Resume()

		// Periodically poll for additional events.
		if !asyncScheduler && n > pollInterval {
			poll()
			n = 0
		} else {
			n++
		}
	}

	scheduleLog("  program complete!")
}

func Gosched() {
	runqueue.Push(task.Current())
	task.Pause()
}
