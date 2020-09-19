// +build wasm

package runtime

import (
	"unsafe"

	"github.com/tinygo-org/tinygo/src/syscall/wasi"
)

func postinit() {}

// Using global variables to avoid heap allocation.
var (
	putcharBuf   = byte(0)
	putcharIOVec = wasi.IOVec{
		Buf:    unsafe.Pointer(&putcharBuf),
		BufLen: 1,
	}
)

func putchar(c byte) {
	// write to stdout
	const stdout = 1
	var nwritten uint
	putcharBuf = c
	wasi.Fd_write(stdout, &putcharIOVec, 1, &nwritten)
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
