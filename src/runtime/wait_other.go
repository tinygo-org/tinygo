//go:build !tinygo.riscv && !cortexm && !(linux && !baremetal && !tinygo.wasm && !nintendoswitch) && !darwin

package runtime

func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
