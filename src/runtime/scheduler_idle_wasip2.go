//go:build wasip2 && (scheduler.tasks || scheduler.asyncify)

package runtime

import (
	monotonicclock "internal/wasi/clocks/v0.2.0/monotonic-clock"
)

// sleepTicks is the cooperative scheduler's "wait until the next deadline"
// primitive on wasip2. It is only called by the scheduler when the run queue
// is empty and there's a sleeping task or pending timer due in d ticks.
//
// If any pollables are registered via netpollAddPollable, this routes through
// pollIO so the same wasi:io/poll.Poll call observes both the clock
// subscription and the registered pollables. With no pollables it falls
// back to the cheap monotonic-clock-Block path.
func sleepTicks(d timeUnit) {
	if pollCount > 0 {
		pollIO(ticksToNanoseconds(d))
		return
	}
	p := monotonicclock.SubscribeDuration(monotonicclock.Duration(d))
	p.Block()
	p.ResourceDrop()
}

// waitForEvents is the cooperative scheduler's "wait until something external
// happens" primitive. It is only called when both the run queue and the
// timer/sleep queues are empty. With no pollables registered this is a
// genuine deadlock; with pollables we block until any of them is ready.
func waitForEvents() {
	if pollCount > 0 {
		pollIO(-1)
		return
	}
	runtimePanic("deadlocked: no event source")
}
