package main

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
	_ "unsafe"
)

// One periodic timer callback per spare core is held in flight at once while
// main adds timers, which take the slots the callbacks left in the queue. The
// callbacks then finish and re-add their timers. Re-adding must not allocate.
// Main only sleeps while the callbacks finish, so any allocation in that window
// comes from the runtime. Main adds a varying number of timers each step so the
// queue length crosses every capacity boundary.

func main() {
	k := runtime.NumCPU() - 1
	if k < 2 {
		println("need at least 3 cores")
		return
	}
	const steps = 300
	bad := 0
	for step := 0; step < steps; step++ {
		var started atomic.Int32
		var release atomic.Bool
		tims := make([]*time.Timer, k)
		for i := range tims {
			tims[i] = newTimer(runtimeNano()+int64(time.Millisecond), int64(time.Hour), func(any, uintptr, int64) {
				if release.Load() {
					return
				}
				started.Add(1)
				for !release.Load() {
				}
			}, nil, nil)
		}
		for started.Load() < int32(k) {
			time.Sleep(100 * time.Microsecond)
		}

		// Fill slots freed by the in-flight timers. These stay queued.
		for i := 0; i < 1+step%k; i++ {
			newTimer(runtimeNano()+int64(time.Hour), 0, func(any, uintptr, int64) {}, nil, nil)
		}

		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		release.Store(true)
		time.Sleep(2 * time.Millisecond)
		runtime.ReadMemStats(&after)

		if d := after.Mallocs - before.Mallocs; d != 0 {
			bad++
			println("step", step, "allocations while", k, "timers re-added:", d)
		}
		for _, t := range tims {
			stopTimer(t)
		}
	}
	if bad != 0 {
		println("FAIL: re-adding timers allocated in", bad, "of", steps, "steps")
		return
	}
	println("timer reserve alloc multicore ok")
}

//go:linkname newTimer time.newTimer
func newTimer(when, period int64, f func(any, uintptr, int64), arg any, cp unsafe.Pointer) *time.Timer

//go:linkname stopTimer time.stopTimer
func stopTimer(tim *time.Timer) bool

//go:linkname runtimeNano time.runtimeNano
func runtimeNano() int64
