package runtime_test

import (
	"runtime"
	"sort"
	"testing"
)

type heapEntry struct {
	t    *runtime.HeapTimer
	when int64
	key  int64
}

// heapModel is a reference model of the timers queued in a heap.
type heapModel struct {
	items []heapEntry
	seq   int64
}

func (m *heapModel) add(t *runtime.HeapTimer, when int64, front bool) {
	m.seq++
	k := m.seq
	if front {
		k = -k
	}
	m.items = append(m.items, heapEntry{t, when, k})
}

// first returns the entry that must fire next.
func (m *heapModel) first() heapEntry {
	best := m.items[0]
	for _, e := range m.items[1:] {
		if e.when < best.when || e.when == best.when && e.key < best.key {
			best = e
		}
	}
	return best
}

func (m *heapModel) remove(t *runtime.HeapTimer) bool {
	for i, e := range m.items {
		if e.t == t {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return true
		}
	}
	return false
}

func push(h *runtime.TimerHeap, when int64, front bool) *runtime.HeapTimer {
	tm := runtime.NewHeapTimer(when)
	h.Push(tm, front)
	return tm
}

func TestTimerHeapEmpty(t *testing.T) {
	var h runtime.TimerHeap
	if h.Peek() != nil || h.Pop() != nil {
		t.Fatal("empty heap returned a timer")
	}
	if h.Remove(runtime.NewHeapTimer(1)) {
		t.Fatal("remove of an unqueued timer succeeded")
	}
}

func TestTimerHeapPopOrder(t *testing.T) {
	var h runtime.TimerHeap
	whens := map[*runtime.HeapTimer]int64{}
	for _, w := range []int64{5, 3, 9, 1, 7, 3, 3, 0, 10} {
		whens[push(&h, w, false)] = w
	}
	if msg := h.Check(); msg != "" {
		t.Fatal(msg)
	}
	prev := int64(-1)
	for h.Len() > 0 {
		tm := h.Pop()
		if whens[tm] < prev {
			t.Fatalf("out of order: %d after %d", whens[tm], prev)
		}
		if runtime.Queued(tm) {
			t.Fatal("popped timer still links its node")
		}
		prev = whens[tm]
	}
}

func TestTimerHeapEqualDeadlineFIFO(t *testing.T) {
	var h runtime.TimerHeap
	var timers []*runtime.HeapTimer
	for i := 0; i < 200; i++ {
		timers = append(timers, push(&h, 42, false))
	}
	for i, want := range timers {
		if h.Pop() != want {
			t.Fatalf("pop %d returned the wrong timer", i)
		}
	}
}

func TestTimerHeapFront(t *testing.T) {
	var h runtime.TimerHeap
	a := push(&h, 7, false)
	b := push(&h, 7, true)
	c := push(&h, 7, false)
	for i, want := range []*runtime.HeapTimer{b, a, c} {
		if h.Pop() != want {
			t.Fatalf("pop %d returned the wrong timer", i)
		}
	}
}

func TestTimerHeapRemove(t *testing.T) {
	var h runtime.TimerHeap
	whens := map[*runtime.HeapTimer]int64{}
	var timers []*runtime.HeapTimer
	for i := 0; i < 64; i++ {
		w := int64((i * 31) % 17)
		tm := push(&h, w, false)
		whens[tm] = w
		timers = append(timers, tm)
	}
	for i, tm := range timers {
		if i%2 != 0 {
			continue
		}
		if !h.Remove(tm) || runtime.Queued(tm) {
			t.Fatalf("remove %d failed", i)
		}
		if h.Remove(tm) {
			t.Fatalf("second remove %d succeeded", i)
		}
		if msg := h.Check(); msg != "" {
			t.Fatal(msg)
		}
	}
	if h.Len() != 32 {
		t.Fatalf("len %d, want 32", h.Len())
	}
	prev := int64(-1)
	for h.Len() > 0 {
		tm := h.Pop()
		if whens[tm] < prev {
			t.Fatal("out of order after removes")
		}
		prev = whens[tm]
	}
}

// A timer queued in another heap must not be removed from this one.
func TestTimerHeapRemoveForeign(t *testing.T) {
	var a, b runtime.TimerHeap
	push(&a, 1, false)
	t2 := push(&b, 2, false)
	push(&b, 3, false)
	if a.Remove(t2) {
		t.Fatal("removed a timer owned by another heap")
	}
	if a.Len() != 1 || b.Len() != 2 {
		t.Fatal("heaps changed")
	}
}

func TestTimerHeapRandom(t *testing.T) {
	for seed := uint32(1); seed <= 100; seed++ {
		state := seed
		rnd := func(n int) int {
			state = state*1664525 + 1013904223
			return int(state>>8) % n
		}
		var h runtime.TimerHeap
		var m heapModel
		var queued []*runtime.HeapTimer
		for op := 0; op < 500; op++ {
			switch r := rnd(10); {
			case r < 5:
				w := int64(rnd(20))
				front := rnd(4) == 0
				tm := push(&h, w, front)
				m.add(tm, w, front)
				queued = append(queued, tm)
			case r < 7 && len(m.items) > 0:
				want := m.first()
				if got := h.Pop(); got != want.t {
					t.Fatalf("seed %d op %d: pop returned the wrong timer, want when=%d", seed, op, want.when)
				}
				m.remove(want.t)
			case len(queued) > 0:
				tm := queued[rnd(len(queued))]
				inModel := m.remove(tm)
				if got := h.Remove(tm); got != inModel {
					t.Fatalf("seed %d op %d: remove returned %v, model has it: %v", seed, op, got, inModel)
				}
			}
			if msg := h.Check(); msg != "" {
				t.Fatalf("seed %d op %d: %s", seed, op, msg)
			}
			if h.Len() != len(m.items) {
				t.Fatalf("seed %d op %d: len %d, model %d", seed, op, h.Len(), len(m.items))
			}
			if len(m.items) > 0 && h.Peek() != m.first().t {
				t.Fatalf("seed %d op %d: peek returned the wrong timer", seed, op)
			}
		}
	}
}

// TestTimerHeapStress runs long random sequences with different deadline
// distributions and bursty fill and drain phases, then drains the heap and
// compares the pop order with a sorted copy of the model.
func TestTimerHeapStress(t *testing.T) {
	distributions := []struct {
		name string
		when func(rnd func(int) int, i int) int64
	}{
		{"random", func(rnd func(int) int, i int) int64 { return int64(rnd(1 << 20)) }},
		{"many ties", func(rnd func(int) int, i int) int64 { return int64(rnd(4)) }},
		{"ascending", func(rnd func(int) int, i int) int64 { return int64(i) }},
		{"descending", func(rnd func(int) int, i int) int64 { return int64(1<<20 - i) }},
		{"same", func(rnd func(int) int, i int) int64 { return 1000 }},
	}
	seeds, ops := uint32(20), 4000
	if testing.Short() {
		seeds, ops = 3, 1000
	}
	for _, d := range distributions {
		for seed := uint32(1); seed <= seeds; seed++ {
			state := seed * 2654435761
			rnd := func(n int) int {
				state = state*1664525 + 1013904223
				return int(state>>8) % n
			}
			var h runtime.TimerHeap
			var m heapModel
			var queued []*runtime.HeapTimer
			// phase decides the chance of a push, so the heap grows and
			// shrinks in bursts.
			pushPercent := 80
			for op := 0; op < ops; op++ {
				if op%250 == 0 {
					pushPercent = 20 + rnd(70)
				}
				switch r := rnd(100); {
				case r < pushPercent:
					w := d.when(rnd, op)
					front := rnd(5) == 0
					tm := push(&h, w, front)
					m.add(tm, w, front)
					queued = append(queued, tm)
				case r < pushPercent+(100-pushPercent)/2 && len(m.items) > 0:
					want := m.first()
					if got := h.Pop(); got != want.t {
						t.Fatalf("%s seed %d op %d: pop returned the wrong timer, want when=%d", d.name, seed, op, want.when)
					}
					m.remove(want.t)
				case len(queued) > 0:
					// Often a timer that is not queued any more.
					tm := queued[rnd(len(queued))]
					inModel := m.remove(tm)
					if got := h.Remove(tm); got != inModel {
						t.Fatalf("%s seed %d op %d: remove returned %v, model has it: %v", d.name, seed, op, got, inModel)
					}
				}
				if op%16 == 0 {
					if msg := h.Check(); msg != "" {
						t.Fatalf("%s seed %d op %d: %s", d.name, seed, op, msg)
					}
					if h.Len() != len(m.items) {
						t.Fatalf("%s seed %d op %d: len %d, model %d", d.name, seed, op, h.Len(), len(m.items))
					}
				}
			}

			want := append([]heapEntry(nil), m.items...)
			sort.Slice(want, func(i, j int) bool {
				if want[i].when != want[j].when {
					return want[i].when < want[j].when
				}
				return want[i].key < want[j].key
			})
			for i, e := range want {
				if got := h.Pop(); got != e.t {
					t.Fatalf("%s seed %d: drain %d of %d returned the wrong timer", d.name, seed, i, len(want))
				}
				if runtime.Queued(e.t) {
					t.Fatalf("%s seed %d: drained timer still links its node", d.name, seed)
				}
			}
			if h.Len() != 0 || h.Pop() != nil {
				t.Fatalf("%s seed %d: heap not empty after drain", d.name, seed)
			}
		}
	}
}
