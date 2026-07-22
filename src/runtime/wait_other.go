//go:build !tinygo.riscv && !cortexm && !(linux && !baremetal && !tinygo.wasm && !nintendoswitch) && !darwin && !wasip1 && !wasip2

package runtime

func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
