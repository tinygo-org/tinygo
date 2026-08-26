package main

import (
	"runtime"
	"syscall/js"
)

//go:wasmimport tester finalizerRan
func finalizerRan()

//go:wasmimport tester backgroundRan
func backgroundRan()

//go:wasmimport tester callNestedExport
func callNestedExport()

var nestedCallback js.Func

//go:wasmexport launchBackground
func launchBackground() {
	go func() {
		backgroundRan()
	}()
}

//go:wasmexport installNestedCallback
func installNestedCallback() {
	nestedCallback = js.FuncOf(func(js.Value, []js.Value) any {
		callNestedExport()
		return nil
	})
	js.Global().Set("nestedExportCallback", nestedCallback)
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
