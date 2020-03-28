package runtime

import "internal/task"

var (
	sleepQueue         *task.Task
	sleepQueueBaseTime timeUnit
)

// Add this task to the sleep queue, assuming its state is set to sleeping.
func addSleepTask(t *task.Task, duration int64) {
	if schedulerDebug {
		println("  set sleep:", t, uint(duration/tickMicros))
		if t.Next != nil {
			panic("runtime: addSleepTask: expected next task to be nil")
		}
	}
	t.Data = uint(duration / tickMicros) // TODO: longer durations
	now := ticks()
	if sleepQueue == nil {
		scheduleLog("  -> sleep new queue")

		// set new base time
		sleepQueueBaseTime = now
	}

	// Add to sleep queue.
	q := &sleepQueue
	for ; *q != nil; q = &(*q).Next {
		if t.Data < (*q).Data {
			// this will finish earlier than the next - insert here
			break
		} else {
			// this will finish later - adjust delay
			t.Data -= (*q).Data
		}
	}
	if *q != nil {
		// cut delay time between this sleep task and the next
		(*q).Data -= t.Data
	}
	t.Next = *q
	*q = t
}

// pollSleepQueue checks the sleep queue to see if any tasks are now ready to run.
func pollSleepQueue(now timeUnit) bool {
	// Add tasks that are done sleeping to the end of the runqueue so they
	// will be executed soon.
	var awoke bool
	for sleepQueue != nil && now-sleepQueueBaseTime >= timeUnit(sleepQueue.Data) {
		t := sleepQueue
		scheduleLogTask("  awake:", t)
		sleepQueueBaseTime += timeUnit(t.Data)
		sleepQueue = t.Next
		t.Next = nil
		runqueue.Push(t)
		awoke = true
	}

	return awoke
}

var curTime timeUnit

// poll checks for any events that are ready and pushes them onto the runqueue.
// It returns true if any tasks were ready.
func poll() bool {
	scheduleLog("  polling for events")
	if sleepQueue == nil {
		return false
	}
	curTime = ticks()
	return pollSleepQueue(curTime)
}

// wait sleeps until any tasks are awoken by external events.
func wait() {
	scheduleLog("  waiting for events")
	if sleepQueue == nil {
		// There are no timers, so timer wakeup is impossible.
		if asyncScheduler {
			// On WASM, sometimes callbacks are used to process wakeups from JS events.
			// Therefore, we do not actually know if we deadlocked or not.
			scheduleLog("  no tasks left!")
			return
		}
		runtimePanic("deadlock")
	}

	// Sleep until the next timer hits.
	timeLeft := timeUnit(sleepQueue.Data) - (curTime - sleepQueueBaseTime)
	if schedulerDebug {
		println("  sleeping...", sleepQueue, uint(timeLeft))
		for t := sleepQueue; t != nil; t = t.Next {
			println("    task sleeping:", t, timeUnit(t.Data))
		}
	}
	sleepTicks(timeLeft)
}
