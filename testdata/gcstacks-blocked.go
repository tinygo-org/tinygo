package main

import "runtime"

// Goroutines blocked on channels that nothing else references can never run
// again. The GC should be able to free their stacks.

const goroutines = 200

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
