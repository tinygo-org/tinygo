package main

import (
	"reflect"

	"github.com/tinygo-org/tinygo/testdata/localtypes/lib"
	"github.com/tinygo-org/tinygo/testdata/localtypes/lib2"
)

var intFunction = lib.GenericFunction[int]()
var arrayFunction = lib.GenericFunction[[4]int]()

func callLocal(fn, input any) any {
	value := reflect.ValueOf(fn)
	arg := reflect.New(value.Type().In(0)).Elem()
	arg.Field(0).Set(reflect.ValueOf(input))
	return value.Call([]reflect.Value{arg})[0].Interface()
}

func expect(name string, ok bool) {
	if ok {
		println("ok:", name)
	} else {
		println("BUG:", name)
	}
}

func main() {
	mainU, mainN, mainUC, mainNC := lib.GenericWithLocals[int]()
	libU, libN, libUC, libNC := lib2.IntPair()

	// Same instance compiled in two packages: each package's checker
	// must accept the other package's value.
	expect("main GenericWithLocals[int].UsesT accepts lib2's value", mainUC(libU))
	expect("main GenericWithLocals[int].NoT accepts lib2's value", mainNC(libN))
	expect("lib2 GenericWithLocals[int].UsesT accepts main's value", libUC(mainU))
	expect("lib2 GenericWithLocals[int].NoT accepts main's value", libNC(mainN))

	// Different instantiations: distinct types.
	stringU, stringN, stringUC, stringNC := lib.GenericWithLocals[string]()
	expect("GenericWithLocals[int].UsesT rejects [string].UsesT", !mainUC(stringU))
	expect("GenericWithLocals[int].NoT rejects [string].NoT", !mainNC(stringN))
	expect("GenericWithLocals[string].UsesT rejects [int].UsesT", !stringUC(mainU))
	expect("GenericWithLocals[string].NoT rejects [int].NoT", !stringNC(mainN))

	// Sibling closures inside one function (declared in lib).
	scA, scB, scAC, scBC := lib.SiblingClosures()
	expect("lib.SiblingClosures.Foo#1 accepts own", scAC(scA))
	expect("lib.SiblingClosures.Foo#2 accepts own", scBC(scB))
	expect("lib.SiblingClosures.Foo#1 rejects Foo#2", !scAC(scB))
	expect("lib.SiblingClosures.Foo#2 rejects Foo#1", !scBC(scA))

	expect("Call generic local in main", callLocal(intFunction, 42).(int) == 42)
	expect("Call generic local in lib2", callLocal(lib2.IntFunction, 43).(int) == 43)
	array := [4]int{1, 2, 3, 4}
	expect("Call aggregate generic local in main", callLocal(arrayFunction, array).([4]int) == array)
	expect("Call aggregate generic local in lib2", callLocal(lib2.ArrayFunction, array).([4]int) == array)
}
