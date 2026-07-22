//go:build wasip2 && !scheduler.tasks && !scheduler.asyncify

package runtime

import (
	monotonicclock "internal/wasi/clocks/v0.2.0/monotonic-clock"
)

// sleepTicks blocks the current execution context for d ticks. This is the
// fallback used when no cooperative scheduler is configured on wasip2 — it
// has no pollable-polling integration, see scheduler_idle_wasip2.go for the
// cooperative variant.
func sleepTicks(d timeUnit) {
	p := monotonicclock.SubscribeDuration(monotonicclock.Duration(d))
	p.Block()
	p.ResourceDrop()
}

// waitForEvents is only meaningful when there's an event source available.
// Without the cooperative scheduler running poll on registered pollables,
// wasip2 has nothing to wake on, so this is a hard deadlock.
func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
