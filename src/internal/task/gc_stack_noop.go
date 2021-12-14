//go:build (!gc.conservative && !gc.extalloc) || !tinygo.wasm || scheduler.coroutines
// +build !gc.conservative,!gc.extalloc !tinygo.wasm scheduler.coroutines

package task

type gcData struct{}

func (gcd *gcData) swap() {
}
