//go:build !darwin && (!linux || baremetal || tinygo.wasm || nintendoswitch)

package runtime

import (
	_ "unsafe"
)

//go:linkname syscall_runtime_BeforeFork syscall.runtime_BeforeFork
func syscall_runtime_BeforeFork() {
}

//go:linkname syscall_runtime_AfterFork syscall.runtime_AfterFork
func syscall_runtime_AfterFork() {
}

//go:linkname syscall_runtime_AfterForkInChild syscall.runtime_AfterForkInChild
func syscall_runtime_AfterForkInChild() {
}
