// +build wasm

package runtime

import "unsafe"

// Implements __wasi_ciovec_t and __wasi_iovec_t.
type wasiIOVec struct {
	buf    unsafe.Pointer
	bufLen uint
}

//go:wasm-module wasi_unstable
//export fd_write
func fd_write(id uint32, iovs *wasiIOVec, iovs_len uint, nwritten *uint) (errno uint)

func postinit() {}

// Using global variables to avoid heap allocation.
var (
	putcharBuf   = byte(0)
	putcharIOVec = wasiIOVec{
		buf:    unsafe.Pointer(&putcharBuf),
		bufLen: 1,
	}
)

func putchar(c byte) {
	// write to stdout
	const stdout = 1
	var nwritten uint
	putcharBuf = c
	fd_write(stdout, &putcharIOVec, 1, &nwritten)
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

// TinyGo does not yet support any form of parallelism on WebAssembly, so these
// can be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}
