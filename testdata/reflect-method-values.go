package main

import "reflect"

type methodTarget int

func (v methodTarget) Keep() int  { return int(v) + 1 }
func (v methodTarget) Other() int { return int(v) + 2 }

var lookupValue = reflect.Value.MethodByName
var lookupType = reflect.Type.MethodByName
var indexedValue = reflect.Value.Method
var boundLookup = reflect.ValueOf(methodTarget(41)).MethodByName
var enumerateValue = reflect.Value.Methods

func main() {
	checkInitializedMethods()
	checkAggregateMethod()
	checkInterfaceMethod()
	checkArrayBounds()
	value := reflect.ValueOf(methodTarget(41))
	if lookupValue(value, "Keep").Call(nil)[0].Int() != 42 {
		panic("value method expression failed")
	}
	method, ok := lookupType(value.Type(), "Other")
	if !ok || method.Func.Call([]reflect.Value{value})[0].Int() != 43 {
		panic("type method expression failed")
	}
	if indexedValue(value, 1).Call(nil)[0].Int() != 43 {
		panic("indexed method expression failed")
	}
	if boundLookup("Keep").Call(nil)[0].Int() != 42 {
		panic("initialized bound method failed")
	}
	count := 0
	for metadata, method := range enumerateValue(value) {
		if method.Call(nil)[0].Int() != int64(42+metadata.Index) {
			panic("method iterator expression failed")
		}
		count++
	}
	if count != 2 {
		panic("method count changed")
	}
	println("initialized reflected method values passed")
}

type aggregateReceiver [200]byte
type aggregateArgument [800]byte

func (r aggregateReceiver) Sum(a aggregateArgument) byte {
	return r[199] + a[799]
}

func (r aggregateReceiver) Pair(a [2]byte) [2]byte {
	return [2]byte{r[199] + a[0], a[1]}
}

func checkAggregateMethod() {
	receiver := aggregateReceiver{199: 17}
	argument := aggregateArgument{799: 25}
	v := reflect.ValueOf(receiver)
	method, ok := v.Type().MethodByName("Sum")
	if !ok || method.Func.Call([]reflect.Value{v, reflect.ValueOf(argument)})[0].Uint() != 42 {
		panic("aggregate method call failed")
	}
	if method.Func.Interface().(func(aggregateReceiver, aggregateArgument) byte)(receiver, argument) != 42 {
		panic("typed aggregate method call failed")
	}
	pairArg := reflect.ValueOf([1][2]byte{{1, 25}}).Index(0)
	pairMethod, ok := v.Type().MethodByName("Pair")
	if !ok || pairMethod.Func.Call([]reflect.Value{v, pairArg})[0].Interface().([2]byte) != [2]byte{18, 25} {
		panic("small aggregate method expression result changed")
	}
	boundPair := v.MethodByName("Pair")
	converted := reflect.ValueOf([]byte{1, 25}).Convert(pairArg.Type())
	if boundPair.Call([]reflect.Value{converted})[0].Interface().([2]byte) != [2]byte{18, 25} {
		panic("small aggregate method result changed")
	}
	if boundPair.Interface().(func([2]byte) [2]byte)([2]byte{1, 25}) != [2]byte{18, 25} {
		panic("typed small aggregate method result changed")
	}
}

func checkArrayBounds() {
	defer func() {
		if recover() == nil {
			panic("out-of-range array index did not panic")
		}
	}()
	reflect.ValueOf([1]byte{7}).Index(1)
}

type interfaceReceiver interface{ Value() int }
type firstReceiver int
type secondReceiver int

func (firstReceiver) Value() int  { return 1 }
func (secondReceiver) Value() int { return 2 }

func checkInterfaceMethod() {
	var receiver interfaceReceiver = firstReceiver(0)
	method := reflect.ValueOf(&receiver).Elem().Method(0)
	receiver = secondReceiver(0)
	if method.Call(nil)[0].Int() != 2 || method.Interface().(func() int)() != 2 {
		panic("reflected method used the old interface receiver")
	}
}

type number int

func (v number) String() string {
	return "number"
}

type function func() int

func (v function) String() string {
	return "function"
}

var numberMethod = reflect.ValueOf(number(23)).Method(0)
var functionMethod = reflect.ValueOf(function(nil)).Method(0)

func checkInitializedMethods() {
	if got := numberMethod.Call(nil)[0].String(); got != "number" {
		panic("initialized number method returned " + got)
	}
	if got := functionMethod.Call(nil)[0].String(); got != "function" {
		panic("initialized function method returned " + got)
	}
	println("initialized method metadata passed")
}
