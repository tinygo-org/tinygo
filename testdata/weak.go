package main

import (
	"runtime"
	"weak"
)

type value struct {
	n int
}

func main() {
	v := &value{n: 42}
	p := weak.Make(v)
	if got := p.Value(); got == nil {
		println("weak pointer lost its value")
	} else {
		println("weak value:", got.n)
	}
	runtime.KeepAlive(v)
}
