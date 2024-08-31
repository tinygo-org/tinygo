package main

import (
	"runtime"
	"sync"
	"time"
)

func init() {
	println("init")
	go println("goroutine in init")
	time.Sleep(1 * time.Millisecond)
}

func main() {
	println("main 1")
	go sub()
	time.Sleep(1 * time.Millisecond)
	println("main 2")
	time.Sleep(2 * time.Millisecond)
	println("main 3")

	// Await a blocking call.
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

	testGoOnBuiltins()

	testGoOnInterface(Foo(0))

	testCond()

	testIssue1790()

	done := make(chan int)
	go testPaddedParameters(paddedStruct{x: 5, y: 7}, done)
	<-done
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

func testGoOnBuiltins() {
	// Test copy builtin (there is no non-racy practical use of this).
	go copy(make([]int, 8), []int{2, 5, 8, 4})

	// Test recover builtin (no-op).
	go recover()

	// Test close builtin.
	ch := make(chan int)
	go close(ch)
	n, ok := <-ch
	if n != 0 || ok != false {
		println("error: expected closed channel to return 0, false")
	}

	// Test delete builtin.
	m := map[string]int{"foo": 3}
	go delete(m, "foo")
	time.Sleep(time.Millisecond)
	v, ok := m["foo"]
	if v != 0 || ok != false {
		println("error: expected deleted map entry to be 0, false")
	}
}

func testCond() {
	var cond runtime.Cond
	go func() {
		// Wait for the caller to wait on the cond.
		time.Sleep(time.Millisecond)

		// Notify the caller.
		ok := cond.Notify()
		if !ok {
			panic("notification not sent")
		}

		// This notification will be buffered inside the cond.
		ok = cond.Notify()
		if !ok {
			panic("notification not queued")
		}

		// This notification should fail, since there is already one buffered.
		ok = cond.Notify()
		if ok {
			panic("notification double-sent")
		}
	}()

	// Verify that the cond has no pending notifications.
	ok := cond.Poll()
	if ok {
		panic("unexpected early notification")
	}

	// Wait for the goroutine spawned earlier to send a notification.
	cond.Wait()

	// The goroutine should have also queued a notification in the cond.
	ok = cond.Poll()
	if !ok {
		panic("missing queued notification")
	}
}

var once sync.Once

func testGoOnInterface(f Itf) {
	go f.Nowait()
	time.Sleep(time.Millisecond)
	go f.Wait()
	time.Sleep(time.Millisecond * 2)
	println("done with 'go on interface'")
}

// This tests a fix for issue 1790:
// https://github.com/tinygo-org/tinygo/issues/1790
func testIssue1790() *int {
	once.Do(func() {})
	i := 0
	return &i
}

type Itf interface {
	Nowait()
	Wait()
}

type Foo string

func (f Foo) Nowait() {
	println("called: Foo.Nowait")
}

func (f Foo) Wait() {
	println("called: Foo.Wait")
	time.Sleep(time.Microsecond)
	println("  ...waited")
}

type paddedStruct struct {
	x uint8
	_ [0]int64
	y uint8
}

// Structs with interesting padding used to crash.
func testPaddedParameters(s paddedStruct, done chan int) {
	println("paddedStruct:", s.x, s.y)
	close(done)
}
