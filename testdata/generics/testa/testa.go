package testa

import (
	"github.com/tinygo-org/tinygo/testdata/generics/value"
)

func Test() {
	v := value.New(1)
	vm := value.Map(v, Plus100)
	vm.Get(callback, callback)
}

func callback(v int) {
	println("value:", v)
}

// Plus100 is a `Transform` that adds 100 to `value`.
func Plus100(value int) int {
	return value + 100
}
