//go:build unknown_wasm
// +build unknown_wasm

package runtime

import (
	"unsafe"
)

//export _start
func _start() {
	// These need to be initialized early so that the heap can be initialized.
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd = uintptr(wasm_memory_size(0) * wasmPageSize)
	run()
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string {
	return []string{}
}

//go:linkname syscall_runtime_envs syscall.runtime_envs
func syscall_runtime_envs() []string {
	return []string{}
}

func putchar(c byte) {
}

func getchar() byte {
	return 0
}

func buffered() int {
	return 0
}

type timeUnit int64

func ticksToNanoseconds(ticks timeUnit) int64 {
	panic("unimplemented: ticksToNanoseconds")
}

func nanosecondsToTicks(ns int64) timeUnit {
	panic("unimplemented: nanosecondsToTicks")
}

func sleepTicks(d timeUnit) {
	panic("unimplemented: sleepTicks")
}

func ticks() timeUnit {
	panic("unimplemented: ticks")
}

//go:linkname now time.now
func now() (int64, int32, int64) {
	panic("unimplemented: now")
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	return
}

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}
