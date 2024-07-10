package main

import "unsafe"

const (
	a = "0123456789abcdef"            // 16 bytes
	b = a + a + a + a + a + a + a + a // 128 bytes
	c = b + b + b + b + b + b + b + b // 1024 bytes
	d = c + c + c + c + c + c + c + c // 8192 bytes
	e = d + d + d + d + d + d + d + d // 65536 bytes
	f = e + e + e + e + e + e + e + e // 524288 bytes
)

var s = f

func main() {
	println(unsafe.StringData(s))
}

// ERROR: program too large for this chip (flash overflowed by {{[0-9]+}} bytes)
// ERROR: 	optimization guide: https://tinygo.org/docs/guides/optimizing-binaries/
