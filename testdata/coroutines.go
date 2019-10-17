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

	// should not print anything, because we should exit as soon as we return
	go delayedPrint()
}

func delayedPrint() {
	time.Sleep(time.Second)
	println("should have already exited")
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
