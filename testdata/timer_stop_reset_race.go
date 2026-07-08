package main

import (
	"time"
	"unsafe"
	_ "unsafe"
)

func main() {
	testStopWhileFiring()
	testResetWhileFiring()
	println("timer stop/reset race tests done")
}

func testStopWhileFiring() {
	started := make(chan struct{})
	release := make(chan struct{})

	tim := newTimerWithCallback(time.Millisecond, 0, func(any, uintptr, int64) {
		close(started)
		<-release
	})

	<-started
	if stopTimer(tim) {
		println("fail: Stop returned true for a firing timer")
	} else {
		println("Stop returned false for firing timer")
	}
	close(release)
}

func testResetWhileFiring() {
	started := make(chan struct{})
	release := make(chan struct{})
	firedAgain := make(chan struct{}, 1)
	first := true

	tim := newTimerWithCallback(time.Millisecond, 5*time.Second, func(any, uintptr, int64) {
		if first {
			first = false
			close(started)
			<-release
			return
		}
		select {
		case firedAgain <- struct{}{}:
		default:
		}
	})

	<-started
	resetPeriodicTimer(tim, 20*time.Millisecond, 5*time.Second)
	close(release)
	time.Sleep(200 * time.Millisecond)
	select {
	case <-firedAgain:
		println("reset firing timer used reset deadline")
	default:
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
