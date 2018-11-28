package main

import (
	"strconv"
	"syscall/js"
)

func main() {
}

//go:export add
func add(a, b int) int {
	return a + b
}

//go:export update
func update() {
	document := js.Global().Get("document")
	a_str := document.Call("getElementById", "a").Get("value").String()
	b_str := document.Call("getElementById", "b").Get("value").String()
	a, _ := strconv.Atoi(a_str)
	b, _ := strconv.Atoi(b_str)
	result := a + b
	document.Call("getElementById", "result").Set("value", result)
}
