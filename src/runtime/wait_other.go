//go:build !tinygo.riscv && !cortexm

package runtime

func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
