//go:build wasip1 && (scheduler.tasks || scheduler.asyncify)

package runtime

// sleepTicks is the cooperative scheduler's "wait until the next deadline"
// primitive on wasip1. It is only called by the scheduler when the run queue
// is empty and there's a sleeping task or pending timer due in d ticks.
//
// If any FD waiters are registered via netpollAddWait, this routes through
// pollIO so the same poll_oneoff call observes both the clock subscription
// and the FD subscriptions. With no FD waiters it falls back to the cheap
// single-clock-subscription path.
func sleepTicks(d timeUnit) {
	if pollCount > 0 {
		pollIO(ticksToNanoseconds(d))
		return
	}
	sleepTicksSubscription.u.u.timeout = uint64(d)
	poll_oneoff(&sleepTicksSubscription, &sleepTicksResult, 1, &sleepTicksNEvents)
}

// waitForEvents is the cooperative scheduler's "wait until something external
// happens" primitive. It is only called when both the run queue and the
// timer/sleep queues are empty. With no FD waiters this is a genuine
// deadlock; with FD waiters we block until any of them is ready.
func waitForEvents() {
	if pollCount > 0 {
		pollIO(-1)
		return
	}
	runtimePanic("deadlocked: no event source")
}
