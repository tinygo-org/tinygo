//go:build gc.conservative && tinygo.wasm && !scheduler.coroutines
// +build gc.conservative,tinygo.wasm,!scheduler.coroutines

package task

import "unsafe"

//go:linkname swapStackChain runtime.swapStackChain
func swapStackChain(dst *unsafe.Pointer)

type gcData struct {
	stackChain unsafe.Pointer
}

func (gcd *gcData) swap() {
	swapStackChain(&gcd.stackChain)
}
