package main

import (
	"runtime"
	"time"
)

// The cost of a collection must grow in proportion to the number of goroutines.
// Goroutine i blocks on a channel and holds the only reference to the channel
// of goroutine i+1, so finding all stacks takes a chain of discoveries.

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
	// The goroutine count grows 4 times. Linear growth gives a ratio near 4,
	// quadratic growth a ratio near 16.
	ratio := float64(b) / float64(a)
	if ratio > 10 {
		println("collection time grew too fast")
		println("ratio x100:", int(ratio*100))
	} else {
		println("collection time scales linearly")
	}
}
