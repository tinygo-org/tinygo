package main

import "runtime"

type object struct {
	data [16]byte
}

func main() {
	defer func() {
		println("recovered:", recover())
	}()
	for i := 0; i < 16; i++ {
		runtime.SetFinalizer(&object{}, func(*object) {
			panic("finalizer panic")
		})
	}
	runtime.GC()
	println("finished")
}
