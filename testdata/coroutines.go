package main

import (
	"sync"
	"time"
)

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

	value := delayedValue()
	println("value produced after some time:", value)

	// Run a non-blocking call in a goroutine. This should be turned into a
	// regular call, so should be equivalent to calling nowait() without 'go'
	// prefix.
	go nowait()
	time.Sleep(time.Millisecond)
	println("done with non-blocking goroutine")

	var printer Printer
	printer = &myPrinter{}
	printer.Print()

	sleepFuncValue(func(x int) {
		time.Sleep(1 * time.Millisecond)
		println("slept inside func pointer", x)
	})
	time.Sleep(1 * time.Millisecond)
	n := 20
	sleepFuncValue(func(x int) {
		time.Sleep(1 * time.Millisecond)
		println("slept inside closure, with value:", n, x)
	})

	time.Sleep(2 * time.Millisecond)

	var x int
	go func() {
		time.Sleep(2 * time.Millisecond)
		x = 1
	}()
	time.Sleep(time.Second / 2)
	println("closure go call result:", x)

	time.Sleep(2 * time.Millisecond)

	var m sync.Mutex
	m.Lock()
	println("pre-acquired mutex")
	go acquire(&m)
	time.Sleep(2 * time.Millisecond)
	println("releasing mutex")
	m.Unlock()
	time.Sleep(2 * time.Millisecond)
	m.Lock()
	println("re-acquired mutex")
	m.Unlock()
	println("done")

	startSimpleFunc(emptyFunc)

	time.Sleep(2 * time.Millisecond)

	// Test ticker.
	ticker := time.NewTicker(time.Millisecond * 10)
	println("waiting on ticker")
	go func() {
		time.Sleep(time.Millisecond * 5)
		println(" - after 5ms")
		time.Sleep(time.Millisecond * 10)
		println(" - after 15ms")
		time.Sleep(time.Millisecond * 10)
		println(" - after 25ms")
	}()
	<-ticker.C
	println("waited on ticker at 10ms")
	<-ticker.C
	println("waited on ticker at 20ms")
	ticker.Stop()
	time.Sleep(time.Millisecond * 20)
	select {
	case <-ticker.C:
		println("fail: ticker should have stopped!")
	default:
		println("ticker was stopped (didn't send anything after 50ms)")
	}

	timer := time.NewTimer(time.Millisecond * 10)
	println("waiting on timer")
	go func() {
		time.Sleep(time.Millisecond * 5)
		println(" - after 5ms")
		time.Sleep(time.Millisecond * 10)
		println(" - after 15ms")
	}()
	<-timer.C
	println("waited on timer at 10ms")
	time.Sleep(time.Millisecond * 10)
}

func acquire(m *sync.Mutex) {
	m.Lock()
	println("acquired mutex from goroutine")
	time.Sleep(2 * time.Millisecond)
	m.Unlock()
	println("released mutex from goroutine")
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

func delayedValue() int {
	time.Sleep(time.Millisecond)
	return 42
}

func sleepFuncValue(fn func(int)) {
	go fn(8)
}

func startSimpleFunc(fn simpleFunc) {
	// Test that named function types don't crash the compiler.
	go fn()
}

func nowait() {
	println("non-blocking goroutine")
}

type Printer interface {
	Print()
}

type myPrinter struct {
}

func (i *myPrinter) Print() {
	time.Sleep(time.Millisecond)
	println("async interface method call")
}

type simpleFunc func()

func emptyFunc() {
}
