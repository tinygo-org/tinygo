package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
	_ "unsafe"
)

// Baremetal targets report GOARCH arm and are slow.
var small = runtime.GOARCH == "arm"

var (
	workers  = 6
	rounds   = 60
	stopped  = 600
	rearms   = 40
	sharedGs = 6
)

func main() {
	if small {
		workers, rounds, stopped, rearms, sharedGs = 3, 20, 100, 10, 3
	}
	var stop atomic.Bool
	var gcs sync.WaitGroup
	gcs.Add(1)
	go func() {
		defer gcs.Done()
		for !stop.Load() {
			runtime.GC()
			time.Sleep(3 * time.Millisecond)
		}
	}()

	oneShotChurn()
	periodicChurn()
	manyStopped()
	selfRearm()
	sharedTimers()

	stop.Store(true)
	gcs.Wait()
	println("timer node reuse tests done")
}

type rng uint64

func (r *rng) next() uint64 {
	x := uint64(*r)
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	*r = rng(x)
	return x
}

// Each worker arms, resets and stops its own AfterFunc timer. The number of
// callbacks must match the number of arms that were not cancelled. A node
// shared by two timers or freed too early loses or duplicates callbacks.
func oneShotChurn() {
	var wg sync.WaitGroup
	var bad atomic.Int32
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			r := rng(seed*2654435761 + 1)
			var fires atomic.Int64
			expected := int64(1)
			tm := time.AfterFunc(time.Duration(r.next()%2000)*time.Microsecond, func() {
				if fires.Add(1)%8 == 0 {
					runtime.GC()
				}
			})
			for i := 0; i < rounds; i++ {
				if i%25 == 0 {
					runtime.GC()
				}
				d := time.Duration(r.next()%1500) * time.Microsecond
				switch r.next() % 4 {
				case 0:
					if tm.Stop() {
						expected--
					}
				case 1:
					if tm.Reset(d) {
						expected--
					}
					expected++
				case 2:
					if tm.Stop() {
						expected--
					}
					if tm.Reset(d) {
						expected--
					}
					expected++
				case 3:
					time.Sleep(d)
				}
			}
			if tm.Stop() {
				expected--
			}
			deadline := time.Now().Add(30 * time.Second)
			for fires.Load() < expected && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			time.Sleep(10 * time.Millisecond)
			if got := fires.Load(); got != expected {
				println("one-shot worker", seed, "fires", got, "want", expected)
				bad.Add(1)
			}
		}(uint64(w))
	}
	wg.Wait()
	if bad.Load() == 0 {
		println("one-shot churn ok")
	}
}

type counter struct {
	fires    atomic.Int64
	inflight atomic.Int32
}

func waitFires(c *counter, n int64) {
	deadline := time.Now().Add(30 * time.Second)
	for c.fires.Load() < int64(n) && time.Now().Before(deadline) {
		time.Sleep(200 * time.Microsecond)
	}
}

// A stopped periodic timer must not fire again, even after its node has been
// reused by another timer.
func periodicChurn() {
	var wg sync.WaitGroup
	var bad atomic.Int32
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			r := rng(seed*40503 + 7)
			for i := 0; i < rounds/10; i++ {
				c := new(counter)
				period := time.Duration(200+r.next()%800) * time.Microsecond
				tm := newTimer(runtimeNano()+int64(period), int64(period), func(any, uintptr, int64) {
					c.inflight.Add(1)
					// Run long enough that Stop and Reset often land mid-callback.
					if c.fires.Add(1)%3 == 0 {
						for end := time.Now().Add(400 * time.Microsecond); time.Now().Before(end); {
						}
					}
					c.inflight.Add(-1)
				}, nil, nil)
				waitFires(c, 2)
				if r.next()%2 == 0 {
					base := c.fires.Load()
					resetTimer(tm, runtimeNano()+int64(period), int64(period))
					waitFires(c, base+2)
				}
				if c.fires.Load() < 2 {
					println("periodic timer did not keep firing", seed, i)
					bad.Add(1)
					return
				}
				if i%4 == 0 {
					runtime.GC()
				}
				cancelled := stopTimer(tm)
				for c.inflight.Load() != 0 {
					runtime.Gosched()
				}
				seen := c.fires.Load()

				// Churn nodes so that a stale reference would hit a reused one.
				var others [4]*time.Timer
				for j := range others {
					others[j] = time.NewTimer(time.Hour)
				}
				for _, o := range others {
					o.Stop()
				}
				runtime.GC()
				time.Sleep(3 * time.Millisecond)
				// A callback that was popped just before Stop may still run once.
				slack := int64(1)
				if cancelled {
					slack = 0
				}
				if c.fires.Load() > seen+slack {
					println("periodic timer fired after stop", seed, i)
					bad.Add(1)
					return
				}
			}
		}(uint64(w))
	}
	wg.Wait()
	if bad.Load() == 0 {
		println("periodic churn ok")
	}
}

// Stop more timers than the free list holds, then create them again.
func manyStopped() {
	for round := 0; round < 3; round++ {
		n := stopped
		var fires atomic.Int64
		tms := make([]*time.Timer, n)
		for i := range tms {
			tms[i] = time.AfterFunc(time.Hour, func() { fires.Add(1) })
		}
		for _, tm := range tms {
			if !tm.Stop() {
				println("fail: Stop of pending timer returned false")
			}
		}
		for i := range tms {
			tms[i] = time.AfterFunc(time.Duration(i%5)*time.Millisecond, func() { fires.Add(1) })
		}
		deadline := time.Now().Add(30 * time.Second)
		for fires.Load() < int64(n) && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		time.Sleep(5 * time.Millisecond)
		if got := fires.Load(); got != int64(n) {
			println("many stopped round", round, "fires", got, "want", n)
			return
		}
	}
	println("many stopped ok")
}

// An AfterFunc that resets itself from its own callback. Reset finds the timer
// firing, so the old node is released after the callback while the new one is
// already queued.
func selfRearm() {
	k := int64(rearms)
	var wg sync.WaitGroup
	var bad atomic.Int32
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var n atomic.Int64
			var tm *time.Timer
			done := make(chan struct{})
			tm = time.AfterFunc(time.Hour, func() {
				if n.Add(1)%16 == 0 {
					runtime.GC()
				}
				if n.Load() < k {
					tm.Reset(time.Duration(n.Load()%3) * 100 * time.Microsecond)
				} else {
					close(done)
				}
			})
			tm.Reset(0)
			select {
			case <-done:
			case <-time.After(60 * time.Second):
				println("self re-arming timer stalled at", n.Load())
				bad.Add(1)
			}
		}()
	}
	wg.Wait()
	if bad.Load() == 0 {
		println("self re-arm ok")
	}
}

// sharedTimer counts what happens to one timer. Every Reset is an arm. A Stop or
// Reset that returns true cancelled a pending arm. Each arm that is not
// cancelled must fire exactly once, even if the callback runs late.
type sharedTimer struct {
	tm      *time.Timer
	arms    atomic.Int64
	cancels atomic.Int64
	fires   atomic.Int64
}

func (s *sharedTimer) stop() {
	if stopTimer(s.tm) {
		s.cancels.Add(1)
	}
}

func (s *sharedTimer) reset(d time.Duration) {
	s.arms.Add(1)
	if resetTimer(s.tm, runtimeNano()+int64(d), 0) {
		s.cancels.Add(1)
	}
}

// checkShared waits until every timer fired as often as it was armed and not
// cancelled. It then checks that no extra callback arrives.
func checkShared(label string, tms []*sharedTimer) bool {
	want := func(s *sharedTimer) int64 { return s.arms.Load() - s.cancels.Load() }
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		done := true
		for _, s := range tms {
			if s.fires.Load() < want(s) {
				done = false
			}
		}
		if done {
			break
		}
		time.Sleep(time.Millisecond)
	}
	runtime.GC()
	time.Sleep(20 * time.Millisecond)
	ok := true
	for i, s := range tms {
		if got := s.fires.Load(); got != want(s) {
			println(label, "timer", i, "fires", got, "want", want(s))
			ok = false
		}
	}
	return ok
}

// Several goroutines Stop and Reset the same timers. The callbacks run in the
// scheduler, so a timer node that is shared or reused too early shows up as a
// lost or extra fire for some timer.
func sharedTimers() {
	const nt = 8
	tms := make([]*sharedTimer, nt)
	for i := range tms {
		s := new(sharedTimer)
		s.arms.Store(1)
		s.tm = newTimer(runtimeNano()+int64(time.Hour), 0, func(arg any, _ uintptr, _ int64) {
			arg.(*sharedTimer).fires.Add(1)
		}, s, nil)
		tms[i] = s
	}
	var wg sync.WaitGroup
	for g := 0; g < sharedGs; g++ {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			r := rng(seed*7919 + 3)
			for i := 0; i < rounds*2; i++ {
				s := tms[r.next()%nt]
				switch r.next() % 3 {
				case 0:
					s.stop()
				case 1:
					s.reset(time.Duration(r.next()%500) * time.Microsecond)
				case 2:
					if i%20 == 0 {
						runtime.GC()
					}
					runtime.Gosched()
				}
			}
		}(uint64(g))
	}
	wg.Wait()
	for _, s := range tms {
		s.stop()
	}
	ok := checkShared("after stop", tms)

	// Arm every timer once more. Each must fire exactly once.
	for _, s := range tms {
		s.reset(time.Millisecond)
	}
	ok = checkShared("after reset", tms) && ok
	if ok {
		println("shared timers ok")
	}
}

//go:linkname newTimer time.newTimer
func newTimer(when, period int64, f func(any, uintptr, int64), arg any, cp unsafe.Pointer) *time.Timer

//go:linkname stopTimer time.stopTimer
func stopTimer(tim *time.Timer) bool

//go:linkname resetTimer time.resetTimer
func resetTimer(tim *time.Timer, when, period int64) bool

//go:linkname runtimeNano time.runtimeNano
func runtimeNano() int64
