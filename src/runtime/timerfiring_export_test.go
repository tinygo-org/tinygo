package runtime

import "unsafe"

// TimersFireConcurrently reports whether timer callbacks run concurrently with
// goroutines, so that Stop or Reset can run while a callback is in flight.
const TimersFireConcurrently = hasParallelism

// TimerBeginFiring does what the timer runner does when the channel timer t
// expires while a receiver waits. t is a *time.Timer or *time.Ticker.
func TimerBeginFiring(t unsafe.Pointer) unsafe.Pointer {
	tim := &(*timeTimer)(t).timer
	timerChanBlock(tim)
	mask := lockTimers()
	tn := timerQueueRemove(tim)
	firingTimersAdd(tn)
	unlockTimers(mask)
	return unsafe.Pointer(tn)
}

// TimerFinishFiring runs the callback of a node from TimerBeginFiring, then
// releases the receiver.
func TimerFinishFiring(t, tn unsafe.Pointer) {
	n := (*timerNode)(tn)
	n.callback(n, 0)
	timerChanUnblock(&(*timeTimer)(t).timer)
}
