package main

import (
	"runtime"
	"sync"
	"time"
)

// Deep calls leave pointers to large buffers in the unused part of a goroutine
// stack. The GC must not treat those words as roots when the goroutine is
// asleep in a timed wait. A buffer that a sleeping goroutine holds must stay.

// Baremetal targets report GOARCH arm and have little RAM.
var small = runtime.GOARCH == "arm"

var (
	goroutines = 50
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
	if small {
		goroutines, depth, bufSize = 8, 16, 4<<10
	}
	var started, done sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		started.Add(1)
		done.Add(1)
		go func() {
			defer done.Done()
			deepAlloc(depth)
			held := make([]byte, 64)
			for j := range held {
				held[j] = byte(j + i)
			}
			started.Done()
			time.Sleep(300 * time.Millisecond)
			for j := range held {
				if held[j] != byte(j+i) {
					panic("buffer of a sleeping goroutine was freed")
				}
			}
		}()
	}
	started.Wait()
	before := heapInuse()
	// The garbage of deepAlloc is about goroutines*depth*bufSize bytes.
	garbage := uint64(goroutines * depth * bufSize)
	after := heapInuse()
	done.Wait()
	_ = before
	if after > garbage/4+(1<<20) {
		println("stale stack words kept garbage alive")
		println("heap in use:", after, "bytes")
	} else {
		println("stack garbage collected")
	}
}
