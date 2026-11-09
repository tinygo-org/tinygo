//go:build scheduler.threads && tinygo.wasm

package runtime

// WebAssembly entry points are shared with the other schedulers, but the
// threads scheduler has no scheduler loop. These are never used.

var mainExited bool

func scheduler(returnAtDeadlock bool) {
	runtimeFatal("unreachable: no scheduler loop")
}
