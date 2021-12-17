//go:build !gc.conservative || !tinygo.wasm || scheduler.coroutines
// +build !gc.conservative !tinygo.wasm scheduler.coroutines

package task

type gcData struct{}

func (gcd *gcData) swap() {
}
