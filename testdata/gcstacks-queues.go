package main

import (
	"runtime"
	"sync"
	"time"
)

// Goroutines that can only be reached through a wait queue (timers, mutexes,
// condition variables, wait groups, selects) while the main goroutine collects.

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

func churn() {
	var all [][]*object
	for i := 0; i < 64; i++ {
		var batch []*object
		for j := 0; j < 16; j++ {
			batch = append(batch, newObject(-1-j))
		}
		all = append(all, batch)
	}
	runtime.KeepAlive(all)
}

func collect() {
	for i := 0; i < 3; i++ {
		runtime.Gosched()
		runtime.GC()
		churn()
	}
}

const n = 12

func sleepers() {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := newObject(i + 100)
			time.Sleep(5 * time.Millisecond)
			a.check("sleeper")
		}(i)
	}
	collect()
	wg.Wait()
	println("sleepers ok")
}

func mutexWaiters() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	mu.Lock()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := newObject(i + 200)
			mu.Lock()
			a.check("mutex waiter")
			mu.Unlock()
		}(i)
	}
	collect()
	mu.Unlock()
	wg.Wait()
	println("mutex waiters ok")
}

func condWaiters() {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := false
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := newObject(i + 300)
			mu.Lock()
			for !ready {
				cond.Wait()
			}
			mu.Unlock()
			a.check("cond waiter")
		}(i)
	}
	collect()
	mu.Lock()
	ready = true
	mu.Unlock()
	cond.Broadcast()
	wg.Wait()
	println("cond waiters ok")
}

func waitGroupWaiters() {
	var gate, done sync.WaitGroup
	gate.Add(1)
	for i := 0; i < n; i++ {
		done.Add(1)
		go func(i int) {
			defer done.Done()
			a := newObject(i + 400)
			gate.Wait()
			a.check("wait group waiter")
		}(i)
	}
	collect()
	gate.Done()
	done.Wait()
	println("wait group waiters ok")
}

// Each goroutine selects on one shared channel and several channels that only
// its own stack references.
func selectWaiters() {
	live := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := newObject(i + 500)
			own1 := make(chan int)
			own2 := make(chan *object)
			own3 := make(chan struct{})
			select {
			case v := <-live:
				if v != 1 {
					panic("wrong value")
				}
			case <-own1:
			case <-own2:
			case own3 <- struct{}{}:
			}
			a.check("select waiter")
		}(i)
	}
	collect()
	for i := 0; i < n; i++ {
		live <- 1
	}
	wg.Wait()
	println("select waiters ok")
}

func timerWaiters() {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a := newObject(i + 600)
			ticker := time.NewTicker(2 * time.Millisecond)
			defer ticker.Stop()
			ticks := 0
			timeout := time.After(10 * time.Millisecond)
			for done := false; !done; {
				select {
				case <-ticker.C:
					ticks++
					a.check("ticker")
				case <-timeout:
					done = true
				}
			}
			a.check("timer waiter")
		}(i)
	}
	collect()
	wg.Wait()
	println("timer waiters ok")
}

func main() {
	sleepers()
	mutexWaiters()
	condWaiters()
	waitGroupWaiters()
	selectWaiters()
	timerWaiters()
}
