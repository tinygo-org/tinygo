package main

import "github.com/tinygo-org/tinygo/compiler/testdata/pragmadeps"

func main() {
	pragmadeps.NoescapeDefinition(nil)
	pragmadeps.NoescapeGeneric[int]()
	pragmadeps.NoescapeGeneric[string]()
	pragmadeps.InlineDeclaration()
	pragmadeps.NoheapDeclaration()
	pragmadeps.NoruntimeDeclaration()
	pragmadeps.NoboundsDeclaration()
	pragmadeps.SectionDeclaration()
	pragmadeps.InterruptDeclaration()
	pragmadeps.WasmImportDefinition()
	pragmadeps.WasmExportDeclaration()
}
