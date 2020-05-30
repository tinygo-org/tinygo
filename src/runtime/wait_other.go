// +build !tinygo.riscv
// +build !cortexm

package runtime

func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
