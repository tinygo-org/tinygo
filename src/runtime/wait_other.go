//go:build !tinygo.riscv && !cortexm
// +build !tinygo.riscv,!cortexm

package runtime

func waitForEvents() {
	runtimePanic("deadlocked: no event source")
}
