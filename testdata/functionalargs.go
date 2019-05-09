package main

import "fmt"

type aStruct struct {
	b bool
	i int
}

type aOption func(*aStruct)

func SetB() aOption {
	return func(as *aStruct) {
		as.b = true
	}
}

func WithI(i int) aOption {
	return func(as *aStruct) {
		as.i = i
	}
}

func New(opts ...aOption) *aStruct {
	as := &aStruct{}
	for _, opt := range opts {
		opt(as)
	}
	return as
}

func (as *aStruct) Print() {
	fmt.Printf("%t %d\n", as.b, as.i)
}

func main() {
	fmt.Println("Case 1")
	a := New()
	a.Print()

	fmt.Println("Case 2")
	b := New(SetB())
	b.Print()

	fmt.Println("Case 3")
	c := New(SetB(), WithI(42))
	c.Print()
}
