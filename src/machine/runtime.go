package machine

import (
	_ "unsafe"
)

//
// This file provides access to runtime package that would not otherwise
// be permitted due to linker dependencies.
//

//go:linkname gosched runtime.Gosched
func gosched()
