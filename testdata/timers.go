package main

import "time"

func main() {
	// Test ticker.
	ticker := time.NewTicker(time.Millisecond * 160)
	println("waiting on ticker")
	go func() {
		time.Sleep(time.Millisecond * 80)
		println(" - after 80ms")
		time.Sleep(time.Millisecond * 160)
		println(" - after 240ms")
		time.Sleep(time.Millisecond * 160)
		println(" - after 400ms")
	}()
	<-ticker.C
	println("waited on ticker at 160ms")
	<-ticker.C
	println("waited on ticker at 320ms")
	ticker.Stop()
	time.Sleep(time.Millisecond * 400)
	select {
	case <-ticker.C:
		println("fail: ticker should have stopped!")
	default:
		println("ticker was stopped (didn't send anything after 400ms)")
	}

	timer := time.NewTimer(time.Millisecond * 160)
	println("waiting on timer")
	go func() {
		time.Sleep(time.Millisecond * 80)
		println(" - after 80ms")
		time.Sleep(time.Millisecond * 160)
		println(" - after 240ms")
	}()
	<-timer.C
	println("waited on timer at 160ms")
	time.Sleep(time.Millisecond * 160)
}
