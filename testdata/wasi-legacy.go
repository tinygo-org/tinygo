package main

import "time"

var result int32

func main() {
	done := make(chan int32)
	go func() {
		time.Sleep(time.Millisecond)
		done <- 42
	}()
	result = <-done
	println("main done")
}

// Unlike a //go:wasmexport function, an //export function can still be called
// after main returns.
//
//export getResult
func getResult() int32 {
	return result
}
