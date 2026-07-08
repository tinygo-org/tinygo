package runtime

import "internal/task"

const schedulerDebug = false

var timerQueue *timerNode

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

func timerQueueAdd(tn *timerNode) {
	q := &timerQueue
	for ; *q != nil; q = &(*q).next {
		if tn.whenTicks() < (*q).whenTicks() {
			// this will finish earlier than the next - insert here
			break
		}
	}
	tn.next = *q
	*q = tn
}

func timerQueueRemove(t *timer) *timerNode {
	for q := &timerQueue; *q != nil; q = &(*q).next {
		if (*q).timer == t {
			scheduleLog("removed timer")
			n := *q
			*q = (*q).next
			return n
		}
	}
	scheduleLog("did not remove timer")
	return nil
}

// firingTimers is a list of timer nodes whose callback is currently running.
// It is only used by schedulers that run timer callbacks concurrently with user
// goroutines (the threads and cores schedulers), so that a timer stopped or
// reset while its callback is running is not re-added to the timer queue by a
// periodic timer's callback. Access is protected by the scheduler's timer lock.
var firingTimers *timerNode

// firingTimersAdd marks the given timer node as currently firing. The caller
// must hold the scheduler's timer lock.
func firingTimersAdd(tn *timerNode) {
	tn.stopped = false
	tn.firingNext = firingTimers
	firingTimers = tn
}

// firingTimersRemove removes the given timer node from the firing list. The
// caller must hold the scheduler's timer lock.
func firingTimersRemove(tn *timerNode) {
	for q := &firingTimers; *q != nil; q = &(*q).firingNext {
		if *q == tn {
			*q = tn.firingNext
			tn.firingNext = nil
			return
		}
	}
}

// firingTimerStop marks a currently-firing timer as stopped, so that its
// callback will not re-add it to the queue. It returns whether the timer is
// currently firing. The caller must hold the scheduler's timer lock.
func firingTimerStop(tim *timer) bool {
	for tn := firingTimers; tn != nil; tn = tn.firingNext {
		if tn.timer == tim {
			tn.stopped = true
			return true
		}
	}
	return false
}

// Goexit terminates the currently running goroutine. No other goroutines are affected.
func Goexit() {
	panicOrGoexit(nil, panicGoexit)
}

//go:linkname fips_getIndicator crypto/internal/fips140.getIndicator
func fips_getIndicator() uint8 {
	return task.Current().FipsIndicator
}

//go:linkname fips_setIndicator crypto/internal/fips140.setIndicator
func fips_setIndicator(indicator uint8) {
	// This indicator is stored per goroutine.
	task.Current().FipsIndicator = indicator
}

//go:linkname fips140_setBypass crypto/fips140.setBypass
func fips140_setBypass() {
	task.Current().FipsOnlyBypass = true
}

//go:linkname fips140_unsetBypass crypto/fips140.unsetBypass
func fips140_unsetBypass() {
	task.Current().FipsOnlyBypass = false
}

//go:linkname fips140_isBypassed crypto/fips140.isBypassed
func fips140_isBypassed() bool {
	return task.Current().FipsOnlyBypass
}
