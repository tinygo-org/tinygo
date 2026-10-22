package main

import "runtime"

type object struct {
	data [16]byte
}

var str string

func register() {
	for i := 0; i < 16; i++ {
		runtime.SetFinalizer(&object{}, func(*object) {
			panic("finalizer panic")
		})
	}
}

//go:noinline
func concat() {
	defer func() {
		println("recovered:", recover())
	}()
	for i := 0; i < 5000; i++ {
		str = "x" + str[:min(len(str), 200)]
	}
}

func main() {
	register()
	concat()
	println("finished")
}
