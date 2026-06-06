//go:build wasip1 && !scheduler.tasks && !scheduler.asyncify

package runtime

// sleepTicks blocks the current execution context for d ticks. This is the
// fallback used when no cooperative scheduler is configured (-scheduler=none
// or -scheduler=threads on wasip1) and it has no FD-polling integration —
// see scheduler_idle_wasip1.go for the cooperative variant.
func sleepTicks(d timeUnit) {
	sleepTicksSubscription.u.u.timeout = uint64(d)
	poll_oneoff(&sleepTicksSubscription, &sleepTicksResult, 1, &sleepTicksNEvents)
}

// waitForEvents is only meaningful when there's an event source available.
// Without the cooperative scheduler running poll_oneoff on FDs, wasip1 has
// nothing to wake on, so this is a hard deadlock.
func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
