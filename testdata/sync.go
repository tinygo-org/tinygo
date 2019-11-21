package main

import (
	"runtime"
	"sync"
)

func schedWait(duration uint) {
	for ; duration > 0; duration-- {
		runtime.Gosched()
	}
}

func mutexUncontended() {
	var m sync.Mutex
	m.Lock()
	m.Unlock()
}

func mutexWait() {
	var done bool
	var m sync.Mutex
	m.Lock()
	go func() {
		defer m.Unlock()

		schedWait(8)

		done = true
	}()
	m.Lock()
	if !done {
		panic("mutex acquired up too early")
	}
}

func waitGroup(n int) {
	val := n
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		ival := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			for val != ival {
				runtime.Gosched()
			}
			val--
		}()
	}

	// this should cause each goroutine to run
	val--

	// wait for completion
	wg.Wait()

	if val != -1 {
		panic("early waitgroup termination")
	}
}

func once() {
	var wg sync.WaitGroup
	var once sync.Once

	var count uint
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			once.Do(func() { count++ })
		}()
	}

	wg.Wait()
	println("number of times run:", count)
}

func cond() {
	var m sync.Mutex
	cond := &sync.Cond{
		L: &m,
	}

	// should do nothing
	cond.Signal()
	cond.Broadcast()

	var finished uint
	for i := 0; i < 4; i++ {
		go func() {
			m.Lock()
			cond.Wait()
			finished++
			m.Unlock()
		}()
	}

	schedWait(8)
	if finished > 0 {
		panic("premature condition wakeup")
	}

	cond.Signal()
	schedWait(8)
	if finished != 1 {
		panic("wrong number of wakeups")
	}

	cond.Signal()
	schedWait(8)
	if finished != 2 {
		panic("wrong number of wakeups")
	}

	cond.Broadcast()
	schedWait(16)
	if finished < 4 {
		panic("not all goroutines woke up")
	}
}

func pool() {
	println("zero pool value:", (&sync.Pool{
		New: func() interface{} {
			return 0
		},
	}).Get())
	println("nil pool value is nil:", new(sync.Pool).Get() == nil)
	new(sync.Pool).Put(1)
}

func main() {
	println("testing uncontended mutex")
	mutexUncontended()

	println("testing mutex wait")
	mutexWait()

	println("testing waitgroup")
	waitGroup(2)

	println("testing empty waitgroup")
	waitGroup(0)

	println("testing once")
	once()

	println("testing cond")
	cond()

	println("testing pool")
	pool()

	println("done")
}
