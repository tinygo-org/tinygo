package main

type x struct{}

func (x x) name() string {
	return "x"
}

type y = x

type a struct {
	n int
}

func (a a) fruit() string {
	return "apple"
}

type b = a

type fruit interface {
	fruit() string
}

type f = fruit

func main() {
	// test a basic alias
	println(y{}.name())

	// test using a type alias value as an interface
	var v f = b{}
	println(v.fruit())

	// test comparing an alias interface with the referred-to type
	println(a{} == b{})
	println(a{2} == b{3})
}
