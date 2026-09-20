package main

import "reflect"

type methodLookup interface {
	MethodByName(string) (reflect.Method, bool)
}

type promotedLookup struct{ methodLookup }

type target int

func (target) Keep() int { return 42 }

func main() {
	value := reflect.ValueOf(target(0))
	lookup := promotedLookup{value.Type()}
	method, ok := lookup.MethodByName("Keep")
	if !ok || method.Func.Call([]reflect.Value{value})[0].Int() != 42 {
		panic("promoted lookup failed")
	}
	println("promoted lookup passed")
}
