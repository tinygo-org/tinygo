package main

import (
	"runtime"
	"runtime/debug"
)

func main() {
	for i := 0; i < 100; i++ {
		runtime.GC()
	}

	s := &debug.GCStats{}
	debug.ReadGCStats(s)

	if s.NumGC < 100 {
		println(s.NumGC)
		panic("NumGC must be greater than or equal to 100")
	}
}
