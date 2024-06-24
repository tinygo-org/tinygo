//go:build go1.23

package runtime

import "unsafe"

// Time functions for Go 1.23 and above.

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
	return removeTimer(&tim.timer)
}

//go:linkname resetTimer time.resetTimer
func resetTimer(t *timeTimer, when, period int64) bool {
	t.timer.when = when
	t.timer.period = period
	removed := removeTimer(&t.timer)
	addTimer(&timerNode{
		timer:    &t.timer,
		callback: timerCallback,
	})
	return removed
}
