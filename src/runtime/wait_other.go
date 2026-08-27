//go:build !tinygo.riscv && !cortexm && !(linux && !baremetal && !tinygo.wasm && !nintendoswitch) && !darwin && !wasip1 && !uefi

package runtime

func waitForEvents() {
	runtimeFatal("deadlocked: no event source")
}
