package main

import (
	"runtime"
	"sync"
)

// Deep calls leave pointers to large buffers in the unused part of a goroutine
// stack. The GC must not treat those stale words as roots.

const (
	goroutines = 100
	depth      = 20
	bufSize    = 16 << 10
)

var sink byte

//go:noinline
func deepAlloc(n int) {
	buf := make([]byte, bufSize)
	buf[0] = byte(n)
	if n > 0 {
		deepAlloc(n - 1)
	}
	sink += buf[0]
}

func heapInuse() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapInuse
}

func main() {
	release := make(chan struct{})
	var parked, done sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		parked.Add(1)
		done.Add(1)
		go func() {
			defer done.Done()
			deepAlloc(depth)
			parked.Done()
			<-release
		}()
	}
	parked.Wait()

	// Retaining every buffer would take goroutines*depth*bufSize bytes.
	retained := goroutines * depth * bufSize
	inuse := heapInuse()
	if inuse > uint64(retained)/4 {
		println("stale stack words kept garbage alive")
		println("heap in use:", inuse, "bytes")
	} else {
		println("stack garbage collected")
	}

	close(release)
	done.Wait()
}
