// +build js,arm avr

package runtime

// This file stubs out some external functions declared by the syscall/js
// package. They cannot be used on microcontrollers.

type js_ref uint64

//go:linkname js_valueGet syscall/js.valueGet
func js_valueGet(v js_ref, p string) js_ref {
	return 0
}

//go:linkname js_valueNew syscall/js.valueNew
func js_valueNew(v js_ref, args []js_ref) (js_ref, bool) {
	return 0, true
}
