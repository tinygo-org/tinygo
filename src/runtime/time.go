package runtime

// timerNode is an element in a linked list of timers.
type timerNode struct {
	next     *timerNode
	timer    *timer
	callback func(*timerNode)
}

// whenTicks returns the (absolute) time when this timer should trigger next.
func (t *timerNode) whenTicks() timeUnit {
	return nanosecondsToTicks(t.timer.when)
}

// Defined in the time package, implemented here in the runtime.
//go:linkname startTimer time.startTimer
func startTimer(tim *timer) {
	addTimer(&timerNode{
		timer:    tim,
		callback: timerCallback,
	})
	scheduleLog("adding timer")
}

// timerCallback is called when a timer expires. It makes sure to call the
// callback in the time package and to re-add the timer to the queue if this is
// a ticker (repeating timer).
// This is intentionally used as a callback and not a direct call (even though a
// direct call would be trivial), because otherwise a circular dependency
// between scheduler, addTimer and timerQueue would form. Such a circular
// dependency causes timerQueue not to get optimized away.
// If timerQueue doesn't get optimized away, small programs (that don't call
// time.NewTimer etc) would still pay the cost of these timers.
func timerCallback(tn *timerNode) {
	// Run timer function (implemented in the time package).
	// The seq parameter to the f function is not used in the time
	// package so is left zero.
	tn.timer.f(tn.timer.arg, 0)

	// If this is a periodic timer (a ticker), re-add it to the queue.
	if tn.timer.period != 0 {
		tn.timer.when += tn.timer.period
		addTimer(tn)
	}
}

//go:linkname stopTimer time.stopTimer
func stopTimer(tim *timer) bool {
	return removeTimer(tim)
}
