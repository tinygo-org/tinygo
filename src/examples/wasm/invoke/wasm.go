package main

import (
	"syscall/js"
)

func runner(this js.Value, args []js.Value) interface{} {
	return args[0].Invoke(args[1]).String()
}

func main() {
	wait := make(chan struct{}, 0)
	js.Global().Set("runner", js.FuncOf(runner))
	<-wait
}
