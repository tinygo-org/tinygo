package runtime

// timerNodeFree holds nodes of finished or stopped timers, linked through
// firingNext, so NewTimer and Reset can reuse them. It is dropped after each
// collection, like sync.Pool before Go 1.13. Guarded by lockTimers.
var (
	timerNodeFree   *timerNode
	timerNodeFreeN  int
	timerNodeFreeGC uint32
)

// ageTimerNodes drops the free list if a collection ran since the last call.
func ageTimerNodes() {
	if n := poolGCCount(); n != timerNodeFreeGC {
		timerNodeFree = nil
		timerNodeFreeN = 0
		timerNodeFreeGC = n
	}
}

// freeTimerNode keeps tn for reuse. The caller holds lockTimers and nothing
// else refers to tn.
func freeTimerNode(tn *timerNode) {
	ageTimerNodes()
	max := 256
	if baremetal {
		max = 16
	}
	if timerNodeFreeN >= max {
		return
	}
	*tn = timerNode{firingNext: timerNodeFree}
	timerNodeFree = tn
	timerNodeFreeN++
}

// releaseTimerNode keeps a node that nothing refers to any more.
func releaseTimerNode(tn *timerNode) {
	mask := lockTimers()
	freeTimerNode(tn)
	unlockTimers(mask)
}

// newTimerNode returns a zeroed node, reusing a freed one when it can.
func newTimerNode() *timerNode {
	mask := lockTimers()
	ageTimerNodes()
	tn := timerNodeFree
	if tn != nil {
		timerNodeFree = tn.firingNext
		timerNodeFreeN--
		tn.firingNext = nil
	}
	unlockTimers(mask)
	if tn == nil {
		tn = new(timerNode)
	}
	return tn
}
