// +build wasm

package runtime

import (
	"unsafe"
)

// Implements __wasi_iovec_t.
type __wasi_iovec_t struct {
	buf    unsafe.Pointer
	bufLen uint
}

//go:wasm-module wasi_unstable
//export fd_write
func fd_write(id uint32, iovs *__wasi_iovec_t, iovs_len uint, nwritten *uint) (errno uint)

func postinit() {}

// Using global variables to avoid heap allocation.
var (
	putcharBuf   = byte(0)
	putcharIOVec = __wasi_iovec_t{
		buf:    unsafe.Pointer(&putcharBuf),
		bufLen: 1,
	}
)

func nativePutchar(c byte) {
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
