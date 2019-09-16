package main

import (
	"strings"
	"syscall/js"
)

func splitter(this js.Value, args []js.Value) interface{} {
	values := strings.Split(args[0].String(), ",")

	result := make([]interface{}, 0)
	for _, each := range values {
		result = append(result, each)
	}

	return js.ValueOf(result)
}

func main() {
	wait := make(chan struct{}, 0)
	js.Global().Set("splitter", js.FuncOf(splitter))
	<-wait
}
