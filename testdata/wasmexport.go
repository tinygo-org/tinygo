package main

import (
	"runtime"
	"time"
)

func init() {
	println("called init")
	go adder()
}

//go:wasmimport tester callTestMain
func callTestMain()

func main() {
	// main.main is not used when using -buildmode=c-shared.
	callTestMain()
}

//go:wasmexport hello
func hello() {
	println("hello!")
}

//go:wasmexport add
func add(a, b int32) int32 {
	println("called add:", a, b)
	addInputs <- a
	addInputs <- b
	return <-addOutput
}

var addInputs = make(chan int32)
var addOutput = make(chan int32)

func adder() {
	for {
		a := <-addInputs
		b := <-addInputs
		time.Sleep(time.Millisecond)
		addOutput <- a + b
	}
}

//go:wasmimport tester callOutside
func callOutside(a, b int32) int32

//go:wasmexport reentrantCall
func reentrantCall(a, b int32) int32 {
	println("reentrantCall:", a, b)
	result := callOutside(a, b)
	println("reentrantCall result:", result)
	return result
}

// Test for bug: https://github.com/tinygo-org/tinygo/issues/4874
//
//go:wasmexport goroutineExit
func goroutineExit() {
	go func() {
		time.Sleep(time.Second * 10)
		println("goroutineExit: exiting goroutine")
	}()
	runtime.Gosched()
	println("goroutineExit: exit")
}
