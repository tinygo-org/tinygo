package main

import "reflect"

func constantValue(v reflect.Value) reflect.Value {
	return v.MethodByName("Keep")
}

func constantType(t reflect.Type) (reflect.Method, bool) {
	return t.MethodByName("Keep")
}

func multipleNames(v reflect.Value) (reflect.Value, reflect.Value) {
	v.MethodByName("Other")
	return v.MethodByName("Keep"), v.MethodByName("Other")
}

func invalidName(v reflect.Value) reflect.Value {
	return v.MethodByName("Keep Other")
}

func unexportedName(v reflect.Value) reflect.Value {
	return v.MethodByName("hidden")
}

func dynamicValue(v reflect.Value, name string) reflect.Value {
	return v.MethodByName(name)
}

func dynamicType(t reflect.Type, name string) (reflect.Method, bool) {
	return t.MethodByName(name)
}

func indexedValue(v reflect.Value, i int) reflect.Value {
	return v.Method(i)
}

func indexedType(t reflect.Type, i int) reflect.Method {
	return t.Method(i)
}

func boundValue(v reflect.Value, name string) reflect.Value {
	lookup := v.MethodByName
	return lookup(name)
}

func escapedValue(v reflect.Value) func(string) reflect.Value {
	return v.MethodByName
}

func escapedType(t reflect.Type) func(string) (reflect.Method, bool) {
	return t.MethodByName
}

func methodExpression() func(reflect.Value, string) reflect.Value {
	return reflect.Value.MethodByName
}

func valueMetadata(v reflect.Value) any {
	return struct{ reflect.Value }{v}
}

func invokedMethod(v interface{ Method(int) reflect.Value }) reflect.Value {
	return v.Method(0)
}

type methodLookup interface {
	MethodByName(string) (reflect.Method, bool)
}

type promotedLookup struct{ methodLookup }

func promotedMethodByName(v promotedLookup) (reflect.Method, bool) {
	return v.MethodByName("Keep")
}

func iterateValue(v reflect.Value) {
	for _, method := range v.Methods() {
		method.Call(nil)
	}
}

func iterateType(t reflect.Type) {
	for range t.Methods() {
	}
}
