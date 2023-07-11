package main

import "time"

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
