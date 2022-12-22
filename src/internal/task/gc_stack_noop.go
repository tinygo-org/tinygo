//go:build (!gc.conservative && !gc.custom) || !tinygo.wasm

package task

type gcData struct{}

func (gcd *gcData) swap() {
}
