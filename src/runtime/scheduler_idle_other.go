//go:build (scheduler.tasks || scheduler.asyncify) && !wasip1

package runtime

// schedulerIdleWait is called by the cooperative scheduler when the run queue
// is empty. timeoutTicks > 0 means a timer or sleep is due in that many
// monotonic ticks. timeoutTicks == 0 means there is no upcoming timer; on
// targets without an external event source this is a deadlock.
//
// The wasip1 variant in scheduler_idle_wasip1.go integrates with the FD
// poller; on every other cooperative-scheduler target this is a thin
// wrapper around the existing sleepTicks / waitForEvents primitives.
func schedulerIdleWait(timeoutTicks timeUnit) {
	if timeoutTicks > 0 {
		sleepTicks(timeoutTicks)
		return
	}
	waitForEvents()
}
