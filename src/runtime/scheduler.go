package runtime

import "internal/task"

const schedulerDebug = false

// Simple logging, for debugging.
func scheduleLog(msg string) {
	if schedulerDebug {
		println("---", msg)
	}
}

// Simple logging with a task pointer, for debugging.
func scheduleLogTask(msg string, t *task.Task) {
	if schedulerDebug {
		println("---", msg, t)
	}
}

// Simple logging with a channel and task pointer.
func scheduleLogChan(msg string, ch *channel, t *task.Task) {
	if schedulerDebug {
		println("---", msg, ch, t)
	}
}

// timerHeap is a binary min-heap of timer nodes, ordered by timer.when and then
// by seq so that timers with the same deadline keep their insertion order.
type timerHeap struct {
	nodes []*timerNode
	seq   int64
}

func (h *timerHeap) peek() *timerNode {
	if len(h.nodes) == 0 {
		return nil
	}
	return h.nodes[0]
}

// push adds tn to the heap. A front node goes before all nodes with the same
// deadline instead of after them.
func (h *timerHeap) push(tn *timerNode, front bool) {
	h.seq++
	tn.seq = h.seq
	if front {
		tn.seq = -h.seq
	}
	tn.timer.node = tn
	h.nodes = append(h.nodes, tn)
	h.fix(len(h.nodes)-1, tn)
}

func (h *timerHeap) pop() *timerNode {
	return h.removeAt(0)
}

// remove removes the queued node of t from the heap, if there is one.
func (h *timerHeap) remove(t *timer) *timerNode {
	tn := t.node
	if tn == nil || tn.index >= len(h.nodes) || h.nodes[tn.index] != tn {
		return nil
	}
	return h.removeAt(tn.index)
}

func (h *timerHeap) removeAt(i int) *timerNode {
	tn := h.nodes[i]
	last := len(h.nodes) - 1
	moved := h.nodes[last]
	h.nodes[last] = nil
	h.nodes = h.nodes[:last]
	if i != last {
		h.fix(i, moved)
	}
	if tn.timer.node == tn {
		tn.timer.node = nil
	}
	return tn
}

// fix places tn, which is not in the slice yet, at the hole at index i.
// It moves up towards the root first and then down.
func (h *timerHeap) fix(i int, tn *timerNode) {
	nodes := h.nodes
	for i > 0 {
		p := nodes[(i-1)/2]
		if tn.timer.when > p.timer.when || tn.timer.when == p.timer.when && tn.seq > p.seq {
			break
		}
		nodes[i] = p
		p.index = i
		i = (i - 1) / 2
	}
	for c := 2*i + 1; c < len(nodes); c = 2*i + 1 {
		if c+1 < len(nodes) && (nodes[c+1].timer.when < nodes[c].timer.when || nodes[c+1].timer.when == nodes[c].timer.when && nodes[c+1].seq < nodes[c].seq) {
			c++
		}
		if nodes[c].timer.when > tn.timer.when || nodes[c].timer.when == tn.timer.when && nodes[c].seq > tn.seq {
			break
		}
		nodes[i] = nodes[c]
		nodes[i].index = i
		i = c
	}
	nodes[i] = tn
	tn.index = i
}

// timerQueue is allocated by the first timerQueueAdd, so that programs
// without timers can drop all timer code from the scheduler.
var timerQueue *timerHeap

func timerQueuePeek() *timerNode {
	if timerQueue == nil {
		return nil
	}
	return timerQueue.peek()
}

func timerQueueAdd(tn *timerNode) {
	if timerQueue == nil {
		timerQueue = new(timerHeap)
	}
	timerQueue.push(tn, false)
}

func timerQueuePop() *timerNode {
	return timerQueue.pop()
}

func timerQueueRemove(t *timer) *timerNode {
	if timerQueue == nil {
		return nil
	}
	n := timerQueue.remove(t)
	if n != nil {
		scheduleLog("removed timer")
	} else {
		scheduleLog("did not remove timer")
	}
	return n
}

// firingTimers is a list of timer nodes whose callback is currently running.
// It is only used by schedulers that run timer callbacks concurrently with user
// goroutines (the threads and cores schedulers), so that a timer stopped or
// reset while its callback is running is not re-added to the timer queue by a
// periodic timer's callback. Access is protected by the scheduler's timer lock.
var firingTimers *timerNode

// firingTimersAdd marks the given timer node as currently firing. The caller
// must hold the scheduler's timer lock.
func firingTimersAdd(tn *timerNode) {
	tn.stopped = false
	tn.firingNext = firingTimers
	firingTimers = tn
}

// firingTimersRemove removes the given timer node from the firing list. The
// caller must hold the scheduler's timer lock.
func firingTimersRemove(tn *timerNode) {
	for q := &firingTimers; *q != nil; q = &(*q).firingNext {
		if *q == tn {
			*q = tn.firingNext
			tn.firingNext = nil
			return
		}
	}
}

// firingTimerStop marks a currently-firing timer as stopped, so that its
// callback will not re-add it to the queue. It returns whether the timer is
// currently firing. The caller must hold the scheduler's timer lock.
func firingTimerStop(tim *timer) bool {
	for tn := firingTimers; tn != nil; tn = tn.firingNext {
		if tn.timer == tim {
			tn.stopped = true
			return true
		}
	}
	return false
}

// Goexit terminates the currently running goroutine. No other goroutines are affected.
func Goexit() {
	panicOrGoexit(nil, panicGoexit)
}

//go:linkname fips_getIndicator crypto/internal/fips140.getIndicator
func fips_getIndicator() uint8 {
	return task.Current().FipsIndicator
}

//go:linkname fips_setIndicator crypto/internal/fips140.setIndicator
func fips_setIndicator(indicator uint8) {
	// This indicator is stored per goroutine.
	task.Current().FipsIndicator = indicator
}

//go:linkname fips140_setBypass crypto/fips140.setBypass
func fips140_setBypass() {
	task.Current().FipsOnlyBypass = true
}

//go:linkname fips140_unsetBypass crypto/fips140.unsetBypass
func fips140_unsetBypass() {
	task.Current().FipsOnlyBypass = false
}

//go:linkname fips140_isBypassed crypto/fips140.isBypassed
func fips140_isBypassed() bool {
	return task.Current().FipsOnlyBypass
}
