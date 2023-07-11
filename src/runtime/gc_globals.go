//go:build (gc.conservative || gc.precise) && (baremetal || tinygo.wasm)

package runtime

// This file implements markGlobals for all the files that don't have a more
// specific implementation.

// markGlobals marks all globals, which are reachable by definition.
//
// This implementation marks all globals conservatively and assumes it can use
// linker-defined symbols for the start and end of the .data section.
func markGlobals() {
	markRoots(globalsStart, globalsEnd)
}
