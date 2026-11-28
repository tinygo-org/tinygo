//go:build js

package runtime

// wasi-libc constructors, which find the preopened directories that the os
// package resolves every path against.
//
//export __wasm_call_ctors
func __wasm_call_ctors()

func init() {
	__wasm_call_ctors()
}
