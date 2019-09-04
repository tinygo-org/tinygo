package reflectlite

import "reflect"

func Swapper(slice interface{}) func(i, j int) {
	return reflect.Swapper(slice)
}

type Value = reflect.Value
type Kind = reflect.Kind
type ValueError = reflect.ValueError
