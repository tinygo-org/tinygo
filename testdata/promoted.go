package main

import (
	"github.com/tinygo-org/tinygo/testdata/promoted/inner"
	"github.com/tinygo-org/tinygo/testdata/promoted/inner2"
)

// Outer has its own unexported methods and also the ones promoted from
// inner.Inner, which Go treats as different methods.
type Outer struct {
	inner.Inner
}

func (Outer) value() int { return 2 }

func (*Outer) ptr() int { return 2 }

type PtrOuter struct {
	*inner.Inner
}

func (PtrOuter) value() int { return 2 }

func (PtrOuter) ptr() int { return 2 }

type Outer2 struct {
	Outer
}

type Both struct {
	inner.Inner
	inner2.Inner2
}

type G[T any] struct {
	inner.Inner
	v T
}

func (G[T]) value() int { return 2 }

type mainValuer interface{ value() int }

func main() {
	println("own:", Outer{}.value())
	println("promoted:", inner.Value(Outer{}))
	println("own ptr:", (&Outer{}).ptr())
	println("promoted ptr:", inner.Ptr(&Outer{}))

	po := PtrOuter{&inner.Inner{}}
	println("embedded ptr own:", po.value())
	println("embedded ptr promoted:", inner.Value(po))
	println("embedded ptr promoted ptr:", inner.Ptr(po))

	println("two levels main:", any(Outer2{}).(mainValuer).value())
	println("two levels inner:", inner.Value(Outer2{}))
	println("two levels inner ptr:", inner.Ptr(&Outer2{}))

	println("both inner:", inner.Value(Both{}))
	println("both inner2:", inner2.Value(Both{}))

	println("generic own:", G[int]{}.value())
	println("generic promoted:", inner.Value(G[int]{}))
	println("generic string promoted:", inner.Value(G[string]{}))
}
