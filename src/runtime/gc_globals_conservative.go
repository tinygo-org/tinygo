//go:build gc.conservative && (baremetal || tinygo.wasm)
// +build gc.conservative
// +build baremetal tinygo.wasm

package runtime

// markGlobals marks all globals, which are reachable by definition.
//
// This implementation marks all globals conservatively and assumes it can use
// linker-defined symbols for the start and end of the .data section.
func markGlobals() {
	markRoots(globalsStart, globalsEnd)
}
