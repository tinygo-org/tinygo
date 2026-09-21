package runtime

import (
	"internal/task"
	"unsafe"
)

var timerChannelsEnabled task.Uint32

// This is the timer that's used internally inside the runtime.
type timer struct {
	lock timerLock

	// When to call the timer, and the interval for the ticker.
	when   int64
	period int64

	// Callback from the time package.
	f   func(arg any, seq uintptr, delta int64)
	arg any

	synctest           *synctestBubble
	pausedNode         *timerNode
	blocked            uint32
	isChan             bool
	stopped            bool
	suppressedCallback bool

	// node is the node of this timer that is waiting in a timer heap, if any.
	node *timerNode
}

func (tim *timer) callCallback(delta int64) {
	tim.f(tim.arg, 0, delta)
}

func (tim *timer) nextWhen(delta int64) int64 {
	next := tim.when + tim.period*(1+delta/tim.period)
	if next < 0 {
		return 1<<63 - 1
	}
	return next
}

// This is the struct used internally in the runtime. The first two fields are
// the same as time.Timer and time.Ticker so it can be used as-is in the time
// package.
type timeTimer struct {
	c         unsafe.Pointer // <-chan time.Time
	initTimer bool
	timer
}

//go:linkname newTimer time.newTimer
func newTimer(when, period int64, f func(arg any, seq uintptr, delta int64), arg any, c unsafe.Pointer) *timeTimer {
	bubble := currentSynctestBubble()
	tim := &timeTimer{
		c:         c,
		initTimer: true,
		timer: timer{
			when:     when,
			period:   period,
			f:        f,
			arg:      arg,
			synctest: bubble,
			isChan:   c != nil,
		},
	}
	if c != nil {
		timerChannelsEnabled.Store(1)
		(*channel)(c).timer = &tim.timer
	}
	scheduleLog("new timer")
	node := newTimerNode()
	node.timer = &tim.timer
	node.callback = timerCallback
	if c != nil && (hasScheduler || hasParallelism) {
		tim.timer.pausedNode = node
	} else if bubble != nil {
		bubble.addTimer(node)
	} else {
		addTimer(node)
	}
	return tim
}

//go:linkname stopTimer time.stopTimer
func stopTimer(tim *timeTimer) bool {
	if tim.timer.synctest != nil {
		tim.timer.synctest.checkTimerAccess("stop")
	}
	tim.timer.lock.Lock()
	tim.timer.suppressedCallback = false
	tim.timer.stopped = true
	var removed bool
	var n *timerNode
	if tim.timer.synctest != nil {
		removed = tim.timer.synctest.removeTimer(&tim.timer, true) != nil
	} else {
		n = removeTimer(&tim.timer, true)
		removed = n != nil
	}
	if tim.timer.pausedNode != nil {
		n = tim.timer.pausedNode
		tim.timer.pausedNode = nil
		removed = true
	}
	drained := timerChanDrain(tim.c)
	suppressed := tim.timer.suppressedCallback
	tim.timer.lock.Unlock()
	if n != nil {
		// Off the queue and not firing, so nothing else refers to n.
		releaseTimerNode(n)
	}
	return removed || drained || suppressed
}

//go:linkname resetTimer time.resetTimer
func resetTimer(t *timeTimer, when, period int64) bool {
	if t.timer.synctest != nil {
		t.timer.synctest.checkTimerAccess("reset")
	}
	t.timer.lock.Lock()
	t.timer.suppressedCallback = false
	var n *timerNode
	if t.timer.synctest != nil {
		n = t.timer.synctest.removeTimer(&t.timer, true)
	} else {
		n = removeTimer(&t.timer, true)
	}
	if n == nil && t.timer.pausedNode != nil {
		n = t.timer.pausedNode
		t.timer.pausedNode = nil
	}
	removed := n != nil
	drained := timerChanDrain(t.c)
	suppressed := t.timer.suppressedCallback
	if n == nil {
		// Allocation can start GC, so do not hold the cores spin lock.
		t.timer.lock.Unlock()
		replacement := newTimerNode()
		t.timer.lock.Lock()
		// A concurrent reset can queue the timer during allocation.
		// Remove it again so this reset takes effect after that operation.
		if t.timer.synctest != nil {
			n = t.timer.synctest.removeTimer(&t.timer, true)
		} else {
			n = removeTimer(&t.timer, true)
		}
		if n == nil && t.timer.pausedNode != nil {
			n = t.timer.pausedNode
			t.timer.pausedNode = nil
		}
		removed = n != nil
		if n == nil {
			n = replacement
		}
		drained = timerChanDrain(t.c) || drained
		suppressed = t.timer.suppressedCallback || suppressed
	}
	t.timer.when = when
	t.timer.period = period
	t.timer.stopped = false
	n.timer = &t.timer
	n.callback = timerCallback
	var runNow bool
	if t.timer.isChan && t.timer.blocked == 0 {
		t.timer.pausedNode = n
	} else if t.timer.synctest != nil {
		runNow = t.timer.synctest.queueTimer(n)
	} else {
		addTimer(n)
	}
	t.timer.lock.Unlock()
	if runNow {
		n.callback(n, 0)
	}
	return removed || drained || suppressed
}

func timerChanRun(tim *timer) {
	if timerChannelsEnabled.Load() == 0 || tim == nil {
		return
	}
	tim.lock.Lock()
	tn := tim.pausedNode
	if tn == nil || tim.stopped {
		tim.lock.Unlock()
		return
	}
	now := nanotime()
	if tim.synctest != nil {
		now = tim.synctest.time()
	}
	if now >= tim.when {
		var current *task.Task
		var bubble unsafe.Pointer
		if synctestIsEnabled() {
			current = task.Current()
			bubble = current.SynctestBubble
			current.SynctestBubble = unsafe.Pointer(tim.synctest)
		}
		tim.callCallback(now - tim.when)
		if current != nil {
			current.SynctestBubble = bubble
		}
		if tim.period != 0 {
			tim.when = tim.nextWhen(now - tim.when)
		} else {
			tim.pausedNode = nil
		}
	}
	tim.lock.Unlock()
}

func timerChanBlock(tim *timer) {
	if timerChannelsEnabled.Load() == 0 || tim == nil {
		return
	}
	tim.lock.Lock()
	tim.blocked++
	tn := tim.pausedNode
	var runNow bool
	if tn != nil && !tim.stopped {
		tim.pausedNode = nil
		if tim.synctest != nil {
			runNow = tim.synctest.queueTimer(tn)
		} else {
			addTimer(tn)
		}
	}
	tim.lock.Unlock()
	if runNow {
		tn.callback(tn, 0)
	}
}

func timerChanUnblock(tim *timer) {
	if timerChannelsEnabled.Load() == 0 || tim == nil {
		return
	}
	tim.lock.Lock()
	if tim.blocked == 0 {
		runtimeFatal("timer channel receive count underflow")
	}
	tim.blocked--
	if tim.blocked == 0 {
		var tn *timerNode
		if tim.synctest != nil {
			tn = tim.synctest.removeTimer(tim, false)
		} else {
			tn = removeTimer(tim, false)
		}
		if tn != nil {
			tim.pausedNode = tn
		}
	}
	tim.lock.Unlock()
}

//go:linkname time_runtimeNano time.runtimeNano
func time_runtimeNano() int64 {
	if bubble := currentSynctestBubble(); bubble != nil {
		return bubble.time()
	}
	return nanotime()
}

//go:linkname time_runtimeNow time.runtimeNow
func time_runtimeNow() (sec int64, nsec int32, mono int64) {
	if bubble := currentSynctestBubble(); bubble != nil {
		now := bubble.time()
		return now / 1e9, int32(now % 1e9), 0
	}
	return now()
}

// timerNode is an element in a timer heap.
type timerNode struct {
	timer    *timer
	callback func(node *timerNode, delta int64)

	// seq breaks ties between equal deadlines, and index is the position in the heap.
	seq   int64
	index int

	// The following fields are only used by schedulers that run timer
	// callbacks concurrently with user goroutines (the threads and cores
	// schedulers). They make it possible to stop or reset a periodic timer (a
	// ticker) while its callback is running, without the callback re-adding the
	// timer to the queue afterwards. They are protected by the scheduler's
	// timer lock for normal timers and the bubble lock for synctest timers.
	//
	// firingNext links nodes whose callback is currently running into the
	// firingTimers list. stopped is set when the timer was stopped or reset
	// while its callback was running, so that timerCallback does not re-add it.
	// All schedulers also use firingNext to link the timerNodeFree list.
	firingNext *timerNode
	stopped    bool
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
	tim := tn.timer
	tim.lock.Lock()

	// Run timer function (implemented in the time package).
	// The seq parameter to the f function is not used in the time
	// package so is left zero.
	if tim.isChan {
		if !tn.stopped {
			tim.callCallback(delta)
		}
	} else {
		tim.lock.Unlock()
		tim.callCallback(delta)
		tim.lock.Lock()
	}

	// Finish firing the timer and re-add it if it is periodic.
	if tim.synctest != nil {
		tim.synctest.finishTimer(tn)
	} else {
		reAddTimer(tn, delta)
	}
	tim.lock.Unlock()
}

//go:linkname time_runtimeIsBubbled time.runtimeIsBubbled
func time_runtimeIsBubbled() bool {
	return currentSynctestBubble() != nil
}
