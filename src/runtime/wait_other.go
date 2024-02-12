//go:build !tinygo.riscv && !cortexm && !uefi

package runtime

func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
