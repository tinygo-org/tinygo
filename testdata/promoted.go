package main

import "github.com/tinygo-org/tinygo/testdata/promoted/inner"

// Outer has its own unexported methods and also the ones promoted from
// inner.Inner, which Go treats as different methods.
type Outer struct {
	inner.Inner
}

func (Outer) value() int { return 2 }

func (*Outer) ptr() int { return 2 }

func main() {
	println("own:", Outer{}.value())
	println("promoted:", inner.Value(Outer{}))
	println("own ptr:", (&Outer{}).ptr())
	println("promoted ptr:", inner.Ptr(&Outer{}))
}
