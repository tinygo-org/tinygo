package main

import (
	"math/rand"
	"runtime"
	"time"
)

func main() {

	ms := runtime.MemStats{}

	for {
		escapesToHeap()
		runtime.ReadMemStats(&ms)
		println("Heap before GC. Used: ", ms.HeapInuse, " Free: ", ms.HeapIdle, " Meta: ", ms.GCSys)
		runtime.GC()
		runtime.ReadMemStats(&ms)
		println("Heap after  GC. Used: ", ms.HeapInuse, " Free: ", ms.HeapIdle, " Meta: ", ms.GCSys)
		time.Sleep(5 * time.Second)
	}

}

func escapesToHeap() {
	n := rand.Intn(100)
	println("Doing ", n, " iterations")
	for i := 0; i < n; i++ {
		s := make([]byte, i)
		_ = append(s, 42)
	}
}
