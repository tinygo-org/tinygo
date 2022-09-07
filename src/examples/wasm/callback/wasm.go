package main

import (
	"strconv"
	"syscall/js"
)

var a, b int

func main() {
	wait := make(chan struct{}, 0)
	document := js.Global().Get("document")
	document.Call("getElementById", "a").Set("oninput", updater(&a))
	document.Call("getElementById", "b").Set("oninput", updater(&b))
	update()
	<-wait
}

func updater(n *int) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		*n, _ = strconv.Atoi(this.Get("value").String())
		update()
		return nil
	})
}

func update() {
	js.Global().Get("document").Call("getElementById", "result").Set("value", a+b)
}
