package main

import "reflect"

type methodLookup interface {
	MethodByName(string) (reflect.Method, bool)
}

type promotedLookup struct{ methodLookup }

type target int

func (target) Keep() int { return 42 }

type boundLookup struct{}

func (boundLookup) Call() int64 {
	return reflect.ValueOf(target(0)).MethodByName("Keep").Call(nil)[0].Int()
}

//go:noinline
func callBoundLookup(v interface{ Call() int64 }) int64 {
	return v.Call()
}

func main() {
	value := reflect.ValueOf(target(0))
	lookup := promotedLookup{value.Type()}
	method, ok := lookup.MethodByName("Keep")
	if !ok || method.Func.Call([]reflect.Value{value})[0].Int() != 42 {
		panic("promoted lookup failed")
	}
	if callBoundLookup(boundLookup{}) != 42 {
		panic("interface-dispatched bound lookup failed")
	}
	println("promoted lookup passed")
}
