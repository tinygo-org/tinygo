package main

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// Goroutines that contend on the sync primitives and on channels. Each part
// has a result that does not depend on the order of the goroutines.

// Baremetal targets report GOARCH arm and have little RAM.
var small = runtime.GOARCH == "arm"

var (
	workers = 16
	iters   = 2000
	pings   = 20000
)

func main() {
	if small {
		workers, iters, pings = 6, 200, 2000
	}
	var wg sync.WaitGroup

	// Mutex, with a yield inside the critical section.
	var mu sync.Mutex
	counter := 0
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				mu.Lock()
				c := counter
				if j%7 == 0 {
					runtime.Gosched()
				}
				counter = c + 1
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	println("mutex:", counter == workers*iters)

	// RWMutex, readers see a pair that a writer updates together.
	var rw sync.RWMutex
	a, b := 0, 0
	var torn atomic.Int64
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				if i%4 == 0 {
					rw.Lock()
					a++
					runtime.Gosched()
					b++
					rw.Unlock()
				} else {
					rw.RLock()
					if a != b {
						torn.Add(1)
					}
					rw.RUnlock()
				}
			}
		}()
	}
	wg.Wait()
	println("rwmutex:", torn.Load() == 0 && a == b)

	// Cond, a bounded queue with producers and consumers.
	var cmu sync.Mutex
	cond := sync.NewCond(&cmu)
	var queue []int
	produced, consumed := 0, 0
	items := iters / 4
	for i := 0; i < workers/2; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < items; j++ {
				cmu.Lock()
				for len(queue) >= 4 {
					cond.Wait()
				}
				queue = append(queue, j)
				produced++
				cond.Broadcast()
				cmu.Unlock()
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < items; j++ {
				cmu.Lock()
				for len(queue) == 0 {
					cond.Wait()
				}
				queue = queue[1:]
				consumed++
				cond.Broadcast()
				cmu.Unlock()
			}
		}()
	}
	wg.Wait()
	println("cond:", produced == consumed && produced == workers/2*items && len(queue) == 0)

	// Once, run by all goroutines at the same time.
	var once sync.Once
	var onceRuns atomic.Int64
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(func() {
				onceRuns.Add(1)
				runtime.Gosched()
			})
		}()
	}
	wg.Wait()
	println("once:", onceRuns.Load() == 1)

	// Unbuffered ping-pong between two goroutines.
	ping, pong := make(chan int), make(chan int)
	go func() {
		for v := range ping {
			pong <- v + 1
		}
		close(pong)
	}()
	v := 0
	for i := 0; i < pings; i++ {
		ping <- v
		v = <-pong
	}
	close(ping)
	println("pingpong:", v == pings)

	// Many senders on one channel, one receiver.
	ch := make(chan int)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 1; j <= iters; j++ {
				ch <- j
			}
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	sum := 0
	for x := range ch {
		sum += x
	}
	println("fan-in:", sum == workers*iters*(iters+1)/2)

	// Select over several channels with a done channel.
	c1, c2, done := make(chan int), make(chan int), make(chan struct{})
	for i, c := range []chan int{c1, c2} {
		go func() {
			for j := 0; j < iters; j++ {
				select {
				case c <- i:
				case <-done:
					return
				}
			}
		}()
	}
	got := 0
	for got < 2*iters {
		select {
		case <-c1:
			got++
		case <-c2:
			got++
		}
	}
	close(done)
	println("select:", got == 2*iters)
}
