//go:build tinygo.wasm && !wasip3

package runtime

// initInRun tells whether package initializers run from the exported entry
// point instead of _initialize.
const initInRun = false

// wasmExportInit runs before every //go:wasmexport function.
func wasmExportInit() {}
