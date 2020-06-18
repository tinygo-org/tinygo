package main

import (
	"fmt"
	"log"
	"syscall/js"
)

func main() {

	// try various log and other output directly
	log.Println("log 1")
	log.Print("log 2")
	log.Printf("log %d\n", 3)
	println("println 4")
	fmt.Println("fmt.Println 5")
	log.Printf("log %s", "6")

	// now set up some log output in a button click callback
	js.Global().
		Get("document").
		Call("querySelector", "#main").
		Set("innerHTML", `<button id="testbtn">Test</button>`)

	js.Global().
		Get("document").
		Call("querySelector", "#testbtn").
		Call("addEventListener", "click",
			js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				println("in func 1")
				log.Printf("in func 2")
				return nil
			}))

	// click the button
	js.Global().
		Get("document").
		Call("querySelector", "#testbtn").
		Call("click")

}
