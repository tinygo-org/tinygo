//go:build js && wasm

// This program mimics a component that fetches once when it is mounted and
// then waits to be halted. The promise callbacks create js.Values on the
// stack of a goroutine; the host checks that stopping the program releases
// the bridge slots those values own.
package main

import "syscall/js"

var halt = make(chan struct{})

func main() {
	slot := js.Global().Get("slot")

	mount := js.FuncOf(func(js.Value, []js.Value) interface{} {
		go fetchBody()
		return nil
	})
	slot.Set("mount", mount)

	stop := js.FuncOf(func(js.Value, []js.Value) interface{} {
		close(halt)
		return nil
	})
	slot.Set("halt", stop)

	<-halt
}

// fetchBody runs fetch -> arrayBuffer -> Uint8Array and then drops the values,
// like a component that loaded its data and is waiting for the next event.
func fetchBody() {
	done := make(chan struct{})
	var onFetch js.Func
	onFetch = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		onFetch.Release()
		response := args[0]
		var onBody js.Func
		onBody = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			onBody.Release()
			body := js.Global().Get("Uint8Array").New(args[0])
			buf := make([]byte, body.Get("length").Int())
			js.CopyBytesToGo(buf, body)
			close(done)
			return nil
		})
		response.Call("arrayBuffer").Call("then", onBody)
		return nil
	})
	js.Global().Call("fetch", "/").Call("then", onFetch)
	<-done
}
