package main

import "github.com/tinygo-org/tinygo/compiler/testdata/pragmadeps"

func main() {
	pragmadeps.Definition()
	new(pragmadeps.Register).Get()
	pragmadeps.Section()
	pragmadeps.Interrupt()
	pragmadeps.Generic[int]()
	pragmadeps.Generic[string]()
	pragmadeps.Noescape(nil)
	pragmadeps.WasmImport()
	pragmadeps.WasmExport()
}
