package main

import (
	"runtime"
	"time"
)

// Collection time must grow linearly with the goroutine count. Each goroutine
// holds the only reference to the channel of the next one.

const (
	small = 400
	large = 1600
)

var head chan int

//go:noinline
func startChain(n int) chan int {
	chans := make([]chan int, n+1)
	for i := range chans {
		chans[i] = make(chan int)
	}
	for i := 0; i < n; i++ {
		cur, next := chans[i], chans[i+1]
		go func() {
			<-cur
			next <- 1
		}()
	}
	head = chans[0]
	return chans[n]
}

func bestGC(n int) time.Duration {
	runtime.Gosched()
	runtime.GC()
	best := time.Duration(1<<63 - 1)
	for i := 0; i < 5; i++ {
		start := time.Now()
		runtime.GC()
		if d := time.Since(start); d < best {
			best = d
		}
	}
	return best
}

func measure(n int) time.Duration {
	last := startChain(n)
	d := bestGC(n)
	head <- 1
	<-last
	head = nil
	runtime.GC()
	return d
}

func main() {
	measure(50) // warm up the heap
	a := measure(small)
	b := measure(large)
	// 4 times the goroutines gives a ratio near 4 when linear, near 16 when
	// quadratic. The limit leaves room for a busy machine.
	ratio := float64(b) / float64(a)
	if ratio > 12 {
		println("collection time grew too fast")
		println("ratio x100:", int(ratio*100))
	} else {
		println("collection time scales linearly")
	}
}
