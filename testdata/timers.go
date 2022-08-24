package main

import "time"

func main() {
	// Test ticker.
	ticker := time.NewTicker(time.Millisecond * 250)
	println("waiting on ticker")
	go func() {
		time.Sleep(time.Millisecond * 125)
		println(" - after 125ms")
		time.Sleep(time.Millisecond * 250)
		println(" - after 375ms")
		time.Sleep(time.Millisecond * 250)
		println(" - after 625ms")
	}()
	<-ticker.C
	println("waited on ticker at 250ms")
	<-ticker.C
	println("waited on ticker at 500ms")
	ticker.Stop()
	time.Sleep(time.Millisecond * 500)
	select {
	case <-ticker.C:
		println("fail: ticker should have stopped!")
	default:
		println("ticker was stopped (didn't send anything after 500ms)")
	}

	timer := time.NewTimer(time.Millisecond * 250)
	println("waiting on timer")
	go func() {
		time.Sleep(time.Millisecond * 125)
		println(" - after 125ms")
		time.Sleep(time.Millisecond * 250)
		println(" - after 250ms")
	}()
	<-timer.C
	println("waited on timer at 250ms")
	time.Sleep(time.Millisecond * 250)

	reset := timer.Reset(time.Millisecond * 250)
	println("timer reset:", reset)
	println("waiting on timer")
	go func() {
		time.Sleep(time.Millisecond * 125)
		println(" - after 125ms")
		time.Sleep(time.Millisecond * 250)
		println(" - after 250ms")
	}()
	<-timer.C
	println("waited on timer at 250ms")
	time.Sleep(time.Millisecond * 250)
}
