package main

import "syscall/js"

func main() {

	ch := make(chan bool, 1)

	println("1")

	js.Global().
		Get("document").
		Call("querySelector", "#main").
		Set("innerHTML", `<button id="testbtn">Test</button>`)

	js.Global().
		Get("document").
		Call("querySelector", "#testbtn").
		Call("addEventListener", "click",
			js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				println("2")
				ch <- true
				println("3")
				return nil
			}))

	println("4")
	v := <-ch
	println(v)

}
