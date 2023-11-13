//go:build (gc.conservative || gc.precise) && (baremetal || tinygo.wasm) && !uefi

package runtime

// This file implements findGlobals for all systems where the start and end of
// the globals section can be found through linker-defined symbols.

// findGlobals finds all globals (which are reachable by definition) and calls
// the callback for them.
//
// This implementation marks all globals conservatively and assumes it can use
// linker-defined symbols for the start and end of the .data section.
func findGlobals(found func(start, end uintptr)) {
	found(globalsStart, globalsEnd)
}
