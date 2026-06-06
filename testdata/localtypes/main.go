package main

import (
	"github.com/tinygo-org/tinygo/testdata/localtypes/lib"
	"github.com/tinygo-org/tinygo/testdata/localtypes/lib2"
)

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
}
