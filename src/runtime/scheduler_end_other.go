//go:build wasi || !wasm

package runtime

func postMain() {
	if hasScheduler {
		schedulerDone = true
	}
}
