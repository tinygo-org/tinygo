//go:build !(gc.conservative || gc.custom || gc.precise) || !tinygo.wasm

package task

type gcData struct{}

func (gcd *gcData) swap() {
}
