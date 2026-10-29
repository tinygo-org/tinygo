package main

import (
	"runtime"
	"sync"
	"time"
)

// Goroutines keep the only reference to heap objects in locals while parked. A
// missed stack root lets churn overwrite the object, and check fails.

// Baremetal targets report GOARCH arm and have little RAM.
var small = runtime.GOARCH == "arm"

// size picks the count for the full run or for small targets.
func size(full, little int) int {
	if small {
		return little
	}
	return full
}

type object struct {
	id   int
	data [14]int
}

//go:noinline
func newObject(id int) *object {
	o := &object{id: id}
	for i := range o.data {
		o.data[i] = id*31 + i
	}
	return o
}

func (o *object) check(what string) {
	if o.id < 0 {
		panic(what + ": object was freed and reused")
	}
	for i, v := range o.data {
		if v != o.id*31+i {
			panic(what + ": object was corrupted")
		}
	}
}

// churn allocates objects that reuse freed memory, then drops them.
func churn() {
	var all [][]*object
	for i := 0; i < size(64, 8); i++ {
		var batch []*object
		for j := 0; j < 16; j++ {
			batch = append(batch, newObject(-1-j))
		}
		all = append(all, batch)
	}
	runtime.KeepAlive(all)
}

func collect() {
	runtime.GC()
	churn()
	runtime.GC()
	churn()
}

func main() {
	testGosched()
	testSleep()
	testDeepFrames()
	testChannel()
	testPendingArgs()
	testPanicRecover()
	testPanicRecoverDeep()
	testMixed()
	println("done")
}

// Goroutines that only hold objects in locals across Gosched.
func testGosched() {
	n := size(40, 12)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := newObject(i)
			b := newObject(i + 1000)
			for round := 0; round < 3; round++ {
				runtime.Gosched()
				if i == 0 {
					collect()
				}
				a.check("gosched a")
				b.check("gosched b")
			}
		}(i)
	}
	wg.Wait()
	println("gosched ok")
}

// Goroutines that only hold objects in locals across a sleep.
func testSleep() {
	n := size(20, 8)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := newObject(i + 2000)
			time.Sleep(time.Duration(i%4+1) * time.Millisecond)
			a.check("sleep a")
			time.Sleep(time.Millisecond)
			a.check("sleep a again")
		}(i)
	}
	// The main goroutine collects while the others are asleep.
	time.Sleep(time.Millisecond / 2)
	collect()
	wg.Wait()
	println("sleep ok")
}

// Every frame of a deep call chain holds its own object, and the bottom of the
// chain parks the goroutine.
//
//go:noinline
func deep(n int, park func()) int {
	o := newObject(n + 3000)
	if n == 0 {
		park()
	} else {
		deep(n-1, park)
	}
	o.check("deep frame")
	return n
}

func testDeepFrames() {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			deep(10+i*5, func() {
				runtime.Gosched()
				if i == 0 {
					collect()
				}
				time.Sleep(time.Millisecond)
				if i == 1 {
					collect()
				}
			})
		}(i)
	}
	wg.Wait()
	println("deep ok")
}

// Goroutines blocked on channels, with the main goroutine collecting.
func testChannel() {
	const n = 16
	start := make(chan struct{})
	results := make(chan int)
	for i := 0; i < n; i++ {
		go func(i int) {
			a := newObject(i + 4000)
			<-start
			a.check("channel a")
			results <- a.id
		}(i)
	}
	runtime.Gosched()
	collect()
	close(start)
	sum := 0
	for i := 0; i < n; i++ {
		sum += <-results
	}
	println("channel sum:", sum)
}

//go:noinline
func panicAndRecover(o *object, depth int) (recovered any) {
	defer func() {
		recovered = recover()
		o.check("panic deferred")
	}()
	panicDeep(o, depth)
	return nil
}

//go:noinline
func panicDeep(o *object, depth int) {
	local := newObject(o.id + depth)
	if depth == 0 {
		collect()
		panic("boom")
	}
	panicDeep(o, depth-1)
	local.check("panic frame")
}

// A panic collects while the stack is being unwound to a deferred recover.
func testPanicRecover() {
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			o := newObject(i + 5000)
			for round := 0; round < 3; round++ {
				if r := panicAndRecover(o, 3+round); r != "boom" {
					panic("wrong recovered value")
				}
				runtime.Gosched()
				collect()
				o.check("after recover")
			}
		}(i)
	}
	wg.Wait()
	println("panic ok")
}

//go:noinline
func parkAfterRecover(o *object, depth int) {
	defer func() {
		recover()
		// The recovered frames are gone, but o and the caller's locals
		// must stay alive while the goroutine parks here.
		runtime.Gosched()
		collect()
		time.Sleep(time.Millisecond)
		collect()
		o.check("parked in deferred call")
	}()
	panicDeep(o, depth)
}

// The goroutine parks inside a deferred function that runs after a panic.
func testPanicRecoverDeep() {
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			o := newObject(i + 6000)
			p := newObject(i + 7000)
			parkAfterRecover(o, 8+i*4)
			o.check("after parked recover o")
			p.check("after parked recover p")
		}(i)
	}
	wg.Wait()
	println("panic park ok")
}

// Many goroutines at different states while a collection runs.
func testMixed() {
	n := size(100, 16)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := newObject(i + 8000)
			switch i % 3 {
			case 0:
				runtime.Gosched()
			case 1:
				time.Sleep(time.Millisecond)
			case 2:
				deep(i%7, runtime.Gosched)
			}
			a.check("mixed a")
		}(i)
	}
	runtime.Gosched()
	collect()
	wg.Wait()
	println("mixed ok")
}

//go:noinline
func parkThenInt(collectToo bool) int {
	runtime.Gosched()
	if collectToo {
		collect()
	}
	time.Sleep(time.Millisecond)
	if collectToo {
		collect()
	}
	return 7
}

//go:noinline
func usePair(a, b *object, n int) {
	a.check("pending arg a")
	b.check("pending arg b")
	if n != 7 {
		panic("wrong int")
	}
}

// The objects exist only as call arguments that wait for a later argument,
// which parks the goroutine.
func testPendingArgs() {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			usePair(newObject(i+9000), newObject(i+9500), parkThenInt(i == 0))
		}(i)
	}
	wg.Wait()
	println("pending args ok")
}
