package runtime

// TimerNodeFreeLen returns the length of the timer node free list as the next
// NewTimer or Reset would see it.
func TimerNodeFreeLen() int {
	mask := lockTimers()
	ageTimerNodes()
	n := timerNodeFreeN
	unlockTimers(mask)
	return n
}
