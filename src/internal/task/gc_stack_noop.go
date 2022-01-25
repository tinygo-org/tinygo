//go:build !gc.conservative || !tinygo.wasm
// +build !gc.conservative !tinygo.wasm

package task

type gcData struct{}

func (gcd *gcData) swap() {
}
