package main

import (
	"math"
	"time"
)

var timer = time.NewTimer(time.Millisecond)

func main() {
	// Test ticker.
	ticker := time.NewTicker(time.Millisecond * 500)
	println("waiting on ticker")
	go func() {
		time.Sleep(time.Millisecond * 150)
		println(" - after 150ms")
		time.Sleep(time.Millisecond * 200)
		println(" - after 200ms")
		time.Sleep(time.Millisecond * 300)
		println(" - after 300ms")
	}()
	<-ticker.C
	println("waited on ticker at 500ms")
	<-ticker.C
	println("waited on ticker at 1000ms")
	ticker.Stop()
	// Ticker.Stop does not drain already-buffered ticks from the channel, and
	// on a slow CI runner a final tick may be delivered concurrently right as
	// Stop is called. Give any such in-flight tick a chance to arrive, then
	// drain the channel before checking that no further ticks are sent.
	time.Sleep(time.Millisecond * 100)
	select {
	case <-ticker.C:
	default:
	}
	time.Sleep(time.Millisecond * 750)
	select {
	case <-ticker.C:
		println("fail: ticker should have stopped!")
	default:
		println("ticker was stopped (didn't send anything after 750ms)")
	}

	timer := time.NewTimer(time.Millisecond * 750)
	println("waiting on timer")
	go func() {
		time.Sleep(time.Millisecond * 200)
		println(" - after 200ms")
		time.Sleep(time.Millisecond * 400)
		println(" - after 400ms")
	}()
	<-timer.C
	println("waited on timer at 750ms")
	time.Sleep(time.Millisecond * 500)

	reset := timer.Reset(time.Millisecond * 750)
	println("timer reset:", reset)
	println("waiting on timer")
	go func() {
		time.Sleep(time.Millisecond * 200)
		println(" - after 200ms")
		time.Sleep(time.Millisecond * 400)
		println(" - after 400ms")
	}()
	<-timer.C
	println("waited on timer at 750ms")
	time.Sleep(time.Millisecond * 500)
	testHugeTicker()
	testTimerStopReset()
	testLazyTimers()
}

func testTimerStopReset() {
	stopped := time.NewTimer(time.Millisecond)
	if len(stopped.C) != 0 || cap(stopped.C) != 0 {
		panic("timer channel was not synchronous")
	}
	time.Sleep(10 * time.Millisecond)
	if !stopped.Stop() {
		panic("Stop returned false for an unread fired timer")
	}
	select {
	case <-stopped.C:
		panic("Stop left a stale timer value")
	default:
	}
	if stopped.Stop() {
		panic("Stop returned true for a stopped timer")
	}

	reset := time.NewTimer(time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	if !reset.Reset(time.Hour) {
		panic("Reset returned false for an unread fired timer")
	}
	select {
	case <-reset.C:
		panic("Reset left a stale timer value")
	default:
	}
	reset.Reset(time.Millisecond)
	<-reset.C
	if reset.Stop() {
		panic("Stop returned true for a received timer")
	}
}

func testHugeTicker() {
	ticker := time.NewTicker(math.MaxInt64)
	defer ticker.Stop()
	time.Sleep(10 * time.Millisecond)
	select {
	case <-ticker.C:
		panic("maximal-period ticker fired early")
	default:
	}
}

func testLazyTimers() {
	for i := 0; i < 100; i++ {
		select {
		case <-time.After(-time.Nanosecond):
		default:
			panic("expired timer was not ready")
		}
	}
	ticker := time.NewTicker(time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	select {
	case <-ticker.C:
	default:
		panic("expired ticker was not ready")
	}
	ticker.Stop()

	timer := time.NewTimer(time.Hour)
	started := make(chan struct{}, 2)
	done := make(chan struct{}, 2)
	for i := 0; i < 2; i++ {
		go func() {
			started <- struct{}{}
			<-timer.C
			done <- struct{}{}
		}()
	}
	<-started
	<-started
	timer.Reset(time.Millisecond)
	<-done
	timer.Reset(time.Millisecond)
	<-done
}
