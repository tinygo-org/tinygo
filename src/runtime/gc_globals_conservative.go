//go:build (gc.conservative || gc.extalloc) && (baremetal || tinygo.wasm)
// +build gc.conservative gc.extalloc
// +build baremetal tinygo.wasm

package runtime

import "unsafe"

// markGlobals marks all globals, which are reachable by definition.
//
// This implementation marks all globals conservatively and assumes it can use
// linker-defined symbols for the start and end of the .data section.
func markGlobals() {
	end := globalsEnd
	if GOARCH == "wasm" {
		// This is a workaround for a bug in wasm-ld: wasm-ld doesn't always
		// align __heap_base and when this memory is shared through an API, it
		// might result in unaligned memory. For details, see:
		// https://reviews.llvm.org/D106499
		// It should be removed once we switch to LLVM 13, where this is fixed.
		end = end &^ (unsafe.Alignof(end) - 1)
	}
	markRoots(globalsStart, end)
}
