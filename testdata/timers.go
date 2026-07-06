package main

import "time"

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
}
