// this is intended to be used as wasm32-unknown-unknown module.
// to compile it, run:
// tinygo build -size short -o hello-unknown.wasm -target wasm-unknown -gc=leaking -no-debug ./src/examples/hello-wasm-unknown/
package main

var x int32

//go:wasmimport hosted echo_i32
func echo(x int32)

//go:export update
func update() {
	x++
	echo(x)
}

func main() {
}
