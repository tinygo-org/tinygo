package testb

import (
	"github.com/tinygo-org/tinygo/testdata/generics/value"
)

func Test() {
	v := value.New(1)
	vm := value.Map(v, Plus500)
	vm.Get(callback, callback)
}

func callback(v int) {
	println("value:", v)
}

// Plus500 is a `Transform` that adds 500 to `value`.
func Plus500(value int) int {
	return value + 500
}
