package main

import "runtime"

// Goroutines blocked on channels that nothing else references can never run
// again. The GC should be able to free their stacks.

// Baremetal targets report GOARCH arm and have little RAM.
var small = runtime.GOARCH == "arm"

var goroutines = 200

func heapInuse() uint64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapInuse
}

//go:noinline
func spawn() {
	for i := 0; i < goroutines; i++ {
		ch := make(chan int)
		go func() { <-ch }()
	}
}

func main() {
	if small {
		goroutines = 40
	}
	before := heapInuse()
	spawn()
	runtime.Gosched()
	after := heapInuse()

	// Each stack is tens of KiB, so 1 MiB allows for ordinary growth only.
	if after > before+1<<20 {
		println("blocked goroutine stacks were retained")
		println("heap growth:", after-before, "bytes")
	} else {
		println("blocked goroutine stacks released")
	}
}
