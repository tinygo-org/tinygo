package builderr

import _ "unsafe"

//go:linkname x notARealFunction
func x()

func Thing() {
	x()
}
