package main

import "reflect"

type value struct {
	Field int
}

//go:wasmexport typeNameLength
func typeNameLength() uint32 {
	return uint32(len(reflect.TypeOf(value{}).String()))
}

func main() {
}
