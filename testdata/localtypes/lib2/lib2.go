package lib2

import "github.com/tinygo-org/tinygo/testdata/localtypes/lib"

func IntPair() (any, any, lib.Checker, lib.Checker) { return lib.GenericWithLocals[int]() }
