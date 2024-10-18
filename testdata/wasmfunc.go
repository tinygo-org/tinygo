package main

import "syscall/js"

func main() {
	js.Global().Call("setCallback", js.FuncOf(func(this js.Value, args []js.Value) any {
		println("inside callback! parameters:")
		sum := 0
		for _, value := range args {
			n := value.Int()
			println("  parameter:", n)
			sum += n
		}
		return sum
	}))
	js.Global().Call("callCallback")
}
