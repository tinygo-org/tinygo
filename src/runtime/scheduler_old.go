package runtime

// This file implements a compatibility layer for the old scheduler layout as we migrate to an interrupt-driven scheduler.

// queue of sleeping goroutines.
var (
	sleepQueue         *task
	sleepQueueBaseTime timeUnit
)

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

// pollTime is the last time value read by poll.
var pollTime timeUnit

// poll checks for goroutines that are ready to wake up.
func poll() *task {
	if sleepQueue == nil {
		return nil
	}

	pollTime = ticks()
	if pollTime-sleepQueueBaseTime >= timeUnit(sleepQueue.state().data) {
		t := sleepQueue
		scheduleLogTask("  awake:", t)
		state := t.state()
		sleepQueueBaseTime += timeUnit(state.data)
		sleepQueue = state.next
		state.next = nil
		return t
	}

	return nil
}

// wait sleeps until a goroutine is ready to wake up.
// This must only be called after poll returns nil.
func wait() {
	if sleepQueue == nil {
		runtimePanic("deadlock")
	}
	sleepTicks(timeUnit(sleepQueue.state().data) - (pollTime - sleepQueueBaseTime))
}
