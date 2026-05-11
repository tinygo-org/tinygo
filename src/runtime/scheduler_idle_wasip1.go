//go:build wasip1 && (scheduler.tasks || scheduler.asyncify)

package runtime

// schedulerIdleWait is called by the cooperative scheduler when the run queue
// is empty. timeoutTicks is the number of monotonic ticks until the next
// sleep / timer is due, or 0 to indicate that no timer is pending and we
// should wait until something external happens (an FD becoming ready).
//
// On wasip1 this routes through pollIO so that any FD waiters registered
// via netpollAddWait participate in the same poll_oneoff call as the
// timer. When there are no FD waiters the cheap legacy paths
// (sleepTicks / waitForEvents) are used.
func schedulerIdleWait(timeoutTicks timeUnit) {
	if pollCount == 0 {
		if timeoutTicks > 0 {
			sleepTicks(timeoutTicks)
			return
		}
		// No timer and no FDs to poll on. Genuine deadlock.
		waitForEvents()
		return
	}
	if timeoutTicks > 0 {
		pollIO(ticksToNanoseconds(timeoutTicks))
		return
	}
	// FDs registered but no timer — block until any FD is ready.
	pollIO(-1)
}
