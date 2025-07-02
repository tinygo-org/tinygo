//go:build wasm_unknown

package runtime

const (
	stdout = 1
)

func putchar(c byte) {
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	return 0, 0, 0
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	// Because this is the "unknown" target we can't call an exit function.
	// But we also can't just return since the program will likely expect this
	// function to never return. So we panic instead.
	runtimePanic("unsupported: syscall.Exit")
}

// There is not yet any support for any form of parallelism on WebAssembly, so these
// can be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

func hardwareRand() (n uint64, ok bool) {
	return 0, false
}

func libc_errno_location() *int32 {
	// CGo is unavailable, so this function should be unreachable.
	runtimePanic("runtime: no cgo errno")
	return nil
}
