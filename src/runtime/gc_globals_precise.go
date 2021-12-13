//go:build gc.conservative && !baremetal && !tinygo.wasm
// +build gc.conservative,!baremetal,!tinygo.wasm

package runtime

import (
	"unsafe"
)

//go:extern runtime.trackedGlobalsStart
var trackedGlobalsStart uintptr

//go:extern runtime.trackedGlobalsLength
var trackedGlobalsLength uintptr

//go:extern runtime.trackedGlobalsBitmap
var trackedGlobalsBitmap [0]uint8

// markGlobals marks all globals, which are reachable by definition.
//
// This implementation relies on a compiler pass that stores all globals in a
// single global (adjusting all uses of them accordingly) and creates a bit
// vector with the locations of each pointer. This implementation then walks the
// bit vector and for each pointer it indicates, it marks the root.
//
//go:nobounds
func markGlobals() {
	for i := uintptr(0); i < trackedGlobalsLength; i++ {
		if trackedGlobalsBitmap[i/8]&(1<<(i%8)) != 0 {
			addr := trackedGlobalsStart + i*unsafe.Alignof(uintptr(0))
			root := *(*uintptr)(unsafe.Pointer(addr))
			markRoot(addr, root)
		}
	}
}
