package main

import "runtime"

//go:wasmimport tester finalizerRan
func finalizerRan()

//go:wasmimport tester backgroundRan
func backgroundRan()

//go:wasmexport launchBackground
func launchBackground() {
	go func() {
		backgroundRan()
	}()
}

//go:noinline
func registerFinalizersImpl() {
	for i := 0; i < 32; i++ {
		p := new([2]int)
		runtime.SetFinalizer(p, func(*[2]int) { finalizerRan() })
	}
}

//go:wasmexport registerFinalizers
func registerFinalizers() {
	registerFinalizersImpl()
	go func() {
		backgroundRan()
	}()
}

func main() {}
