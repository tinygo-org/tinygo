package runtime

// HeapTimer and TimerHeap expose the timer heap to the runtime_test package.
type HeapTimer = timer

type TimerHeap struct {
	h timerHeap
}

func NewHeapTimer(when int64) *HeapTimer {
	return &timer{when: when}
}

func (h *TimerHeap) Len() int {
	return len(h.h.nodes)
}

func (h *TimerHeap) Push(t *HeapTimer, front bool) {
	h.h.push(&timerNode{timer: t}, front)
}

// Peek returns the timer that fires next, or nil.
func (h *TimerHeap) Peek() *HeapTimer {
	if n := h.h.peek(); n != nil {
		return n.timer
	}
	return nil
}

// Pop removes and returns the timer that fires next, or nil.
func (h *TimerHeap) Pop() *HeapTimer {
	if h.h.peek() == nil {
		return nil
	}
	return h.h.pop().timer
}

// Remove reports whether t was queued in this heap.
func (h *TimerHeap) Remove(t *HeapTimer) bool {
	return h.h.remove(t) != nil
}

// Queued reports whether t still links a heap node.
func Queued(t *HeapTimer) bool {
	return t.node != nil
}

// Check returns a description of the first broken heap invariant, or "".
func (h *TimerHeap) Check() string {
	for i, n := range h.h.nodes {
		if n.index != i {
			return "node index does not match its position"
		}
		if n.timer.node != n {
			return "timer does not link its node"
		}
		if i > 0 {
			p := h.h.nodes[(i-1)/2]
			if n.timer.when < p.timer.when || n.timer.when == p.timer.when && n.seq < p.seq {
				return "heap property broken"
			}
		}
	}
	return ""
}
