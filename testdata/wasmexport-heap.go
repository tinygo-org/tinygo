package main

import "runtime"

//go:wasmexport noop
func noop() {}

//go:wasmexport heapInuse
func heapInuse() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapInuse
}

func main() {}
