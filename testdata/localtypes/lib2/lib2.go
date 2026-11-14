package lib2

import "github.com/tinygo-org/tinygo/testdata/localtypes/lib"

var IntFunction = lib.GenericFunction[int]()
var ArrayFunction = lib.GenericFunction[[4]int]()

func IntPair() (any, any, lib.Checker, lib.Checker) { return lib.GenericWithLocals[int]() }
