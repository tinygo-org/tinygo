package runtime

import "unsafe"

// This is the timer that's used internally inside the runtime.
type timer struct {
	// When to call the timer, and the interval for the ticker.
	when   int64
	period int64

	// Callback from the time package.
	f   func(arg any, seq uintptr, delta int64)
	arg any
}

func (tim *timer) callCallback(delta int64) {
	tim.f(tim.arg, 0, delta)
}

// This is the struct used internally in the runtime. The first two fields are
// the same as time.Timer and time.Ticker so it can be used as-is in the time
// package.
type timeTimer struct {
	c    unsafe.Pointer // <-chan time.Time
	init bool
	timer
}

//go:linkname newTimer time.newTimer
func newTimer(when, period int64, f func(arg any, seq uintptr, delta int64), arg any, c unsafe.Pointer) *timeTimer {
	tim := &timeTimer{
		c:    c,
		init: true,
		timer: timer{
			when:   when,
			period: period,
			f:      f,
			arg:    arg,
		},
	}
	scheduleLog("new timer")
	addTimer(&timerNode{
		timer:    &tim.timer,
		callback: timerCallback,
	})
	return tim
}

//go:linkname stopTimer time.stopTimer
func stopTimer(tim *timeTimer) bool {
	return removeTimer(&tim.timer) != nil
}

//go:linkname resetTimer time.resetTimer
func resetTimer(t *timeTimer, when, period int64) bool {
	t.timer.when = when
	t.timer.period = period
	n := removeTimer(&t.timer)
	removed := n != nil
	if n == nil {
		n = new(timerNode)
	}
	n.timer = &t.timer
	n.callback = timerCallback
	addTimer(n)
	return removed
}

//go:linkname time_runtimeNano time.runtimeNano
func time_runtimeNano() int64 {
	// Note: we're ignoring sync groups here (package testing/synctest).
	// See: https://github.com/golang/go/issues/67434
	return nanotime()
}

//go:linkname time_runtimeNow time.runtimeNow
func time_runtimeNow() (sec int64, nsec int32, mono int64) {
	// Also ignoring the sync group here, like time_runtimeNano above.
	return now()
}

// timerNode is an element in a linked list of timers.
type timerNode struct {
	next     *timerNode
	timer    *timer
	callback func(node *timerNode, delta int64)
}

// whenTicks returns the (absolute) time when this timer should trigger next.
func (t *timerNode) whenTicks() timeUnit {
	return nanosecondsToTicks(t.timer.when)
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
func timerCallback(tn *timerNode, delta int64) {
	// Run timer function (implemented in the time package).
	// The seq parameter to the f function is not used in the time
	// package so is left zero.
	tn.timer.callCallback(delta)

	// If this is a periodic timer (a ticker), re-add it to the queue.
	if tn.timer.period != 0 {
		tn.timer.when += tn.timer.period
		addTimer(tn)
	}
}

//go:linkname time_runtimeIsBubbled time.runtimeIsBubbled
func time_runtimeIsBubbled() bool {
	// We don't currently support bubbles.
	return false
}
