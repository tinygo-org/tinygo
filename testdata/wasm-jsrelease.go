package main

import (
	"runtime"
	"syscall/js"
	"time"
)

const values = 2000

func main() {
	array := js.Global().Get("Uint8Array")
	for i := 0; i < values; i++ {
		buffer := array.New(1024)
		buffer.SetIndex(0, i)
		if i%100 == 99 {
			runtime.GC()
			time.Sleep(time.Millisecond)
		}
	}
	js.Global().Call("report", values)
}
