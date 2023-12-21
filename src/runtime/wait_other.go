//go:build !tinygo.riscv && !cortexm && !scheduler.threads

package runtime

func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
