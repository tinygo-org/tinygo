package main

import "github.com/tinygo-org/tinygo/testdata/promoted/inner"

// Outer has its own unexported value method and also the one promoted from
// inner.Inner, which Go treats as a different method.
type Outer struct {
	inner.Inner
}

func (Outer) value() int { return 2 }

func main() {
	println("own:", Outer{}.value())
	println("promoted:", inner.Value(Outer{}))
}
