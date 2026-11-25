package main

import "runtime"

type object struct {
	data [16]byte
}

var finalized bool

//go:noinline
func register() {
	for i := 0; i < 16; i++ {
		runtime.SetFinalizer(&object{}, func(*object) {
			defer func() {
				if recover() != "finalizer panic" {
					panic("finalizer recovery failed")
				}
			}()
			finalized = true
			panic("finalizer panic")
		})
	}
}

//go:noinline
func run() {
	defer func() {
		if recover() != "caller panic" {
			panic("caller recovery failed")
		}
		println("caller recovered")
	}()

	register()
	for i := 0; i < 10 && !finalized; i++ {
		runtime.GC()
	}
	if !finalized {
		panic("finalizer did not run")
	}
	println("finalizer recovered")
	panic("caller panic")
}

func main() {
	run()
}
