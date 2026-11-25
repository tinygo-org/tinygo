package main

import (
	"sync/atomic"
	"time"
	"unsafe"
	_ "unsafe"
)

func main() {
	testStopWhileFiring()
	testStopPeriodicWhileFiring()
	testResetWhileFiring()
	println("timer stop/reset race tests done")
}

// The callbacks must not block on channels. The cores scheduler runs them on a
// scheduler stack with no current task, so they spin on atomics instead.

func waitFor(flag *atomic.Bool) {
	for !flag.Load() {
		time.Sleep(time.Millisecond)
	}
}

func testStopWhileFiring() {
	var started, release atomic.Bool

	tim := newTimerWithCallback(time.Millisecond, 0, func(any, uintptr, int64) {
		started.Store(true)
		for !release.Load() {
		}
	})

	waitFor(&started)
	if stopTimer(tim) {
		println("fail: Stop returned true for a firing timer")
	} else {
		println("Stop returned false for firing timer")
	}
	release.Store(true)
}

func testStopPeriodicWhileFiring() {
	var started, release atomic.Bool
	var calls atomic.Int32

	tim := newTimerWithCallback(time.Millisecond, 10*time.Millisecond, func(any, uintptr, int64) {
		if calls.Add(1) == 1 {
			started.Store(true)
			for !release.Load() {
			}
		}
	})

	waitFor(&started)
	stopTimer(tim)
	release.Store(true)
	time.Sleep(100 * time.Millisecond)
	if calls.Load() == 1 {
		println("stopped periodic timer did not fire again")
	} else {
		println("fail: stopped periodic timer fired again")
	}
}

func testResetWhileFiring() {
	var started, release, firedAgain atomic.Bool
	var calls atomic.Int32

	tim := newTimerWithCallback(time.Millisecond, 5*time.Second, func(any, uintptr, int64) {
		if calls.Add(1) == 1 {
			started.Store(true)
			for !release.Load() {
			}
			return
		}
		firedAgain.Store(true)
	})

	waitFor(&started)
	resetPeriodicTimer(tim, 20*time.Millisecond, 5*time.Second)
	release.Store(true)
	time.Sleep(200 * time.Millisecond)
	if firedAgain.Load() {
		println("reset firing timer used reset deadline")
	} else {
		println("fail: reset firing timer missed reset deadline")
	}
	stopTimer(tim)
}

func newTimerWithCallback(delay, period time.Duration, f func(any, uintptr, int64)) *time.Timer {
	return newTimer(runtimeNano()+int64(delay), int64(period), f, nil, nil)
}

func resetPeriodicTimer(tim *time.Timer, delay, period time.Duration) bool {
	return resetTimer(tim, runtimeNano()+int64(delay), int64(period))
}

//go:linkname newTimer time.newTimer
func newTimer(when, period int64, f func(any, uintptr, int64), arg any, cp unsafe.Pointer) *time.Timer

//go:linkname stopTimer time.stopTimer
func stopTimer(tim *time.Timer) bool

//go:linkname resetTimer time.resetTimer
func resetTimer(tim *time.Timer, when, period int64) bool

//go:linkname runtimeNano time.runtimeNano
func runtimeNano() int64
