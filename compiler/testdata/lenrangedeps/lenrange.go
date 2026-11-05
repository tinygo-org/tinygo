package lenrangedeps

//go:wasmexport exportedString
func WasmExportString(s string) bool {
	return len(s) < 0
}
