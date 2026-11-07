package main

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
	_ "unsafe"
)

// A one-shot callback is held in flight while main adds a timer, which takes
// the slot the callback left in the queue. Finishing the callback must not
// allocate, because a one-shot timer is not added back. Main only sleeps while
// the callback finishes, so any allocation in that window comes from the
// runtime. The number of queued timers is swept so the queue length crosses
// every capacity boundary.

func main() {
	const steps = 200
	bad := 0
	for step := 0; step < steps; step++ {
		var started, release atomic.Bool
		newTimer(runtimeNano()+int64(time.Millisecond), 0, func(any, uintptr, int64) {
			started.Store(true)
			for !release.Load() {
			}
		}, nil, nil)
		for !started.Load() {
			time.Sleep(100 * time.Microsecond)
		}

		// Fill the slot freed by the in-flight timer. This stays queued.
		newTimer(runtimeNano()+int64(time.Hour), 0, func(any, uintptr, int64) {}, nil, nil)

		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		release.Store(true)
		time.Sleep(2 * time.Millisecond)
		runtime.ReadMemStats(&after)

		if d := after.Mallocs - before.Mallocs; d != 0 {
			bad++
			println("queued", step+1, "allocations while a one-shot timer finished:", d)
		}
	}
	if bad != 0 {
		println("FAIL: finishing a one-shot timer allocated in", bad, "of", steps, "steps")
		return
	}
	println("timer reserve alloc ok")
}

//go:linkname newTimer time.newTimer
func newTimer(when, period int64, f func(any, uintptr, int64), arg any, cp unsafe.Pointer) *time.Timer

//go:linkname runtimeNano time.runtimeNano
func runtimeNano() int64
