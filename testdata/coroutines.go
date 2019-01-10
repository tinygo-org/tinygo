package main

import "time"

func main() {
	println("main 1")
	go sub()
	time.Sleep(1 * time.Millisecond)
	println("main 2")
	time.Sleep(2 * time.Millisecond)
	println("main 3")

	// Await a blocking call. This must create a new coroutine.
	println("wait:")
	wait()
	println("end waiting")

	// Run a non-blocking call in a goroutine. This should be turned into a
	// regular call, so should be equivalent to calling nowait() without 'go'
	// prefix.
	go nowait()
	time.Sleep(time.Millisecond)
	println("done with non-blocking goroutine")
}

func sub() {
	println("sub 1")
	time.Sleep(2 * time.Millisecond)
	println("sub 2")
}

func wait() {
	println("  wait start")
	time.Sleep(time.Millisecond)
	println("  wait end")
}

func nowait() {
	println("non-blocking goroutine")
}
