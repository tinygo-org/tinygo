package main

import "github.com/tinygo-org/tinygo/compiler/testdata/lenrangedeps"

//go:wasmexport exportString
func exportString(s string) bool {
	return len(s) < 0
}

func callExportString(s string) bool {
	return lenrangedeps.WasmExportString(s)
}
