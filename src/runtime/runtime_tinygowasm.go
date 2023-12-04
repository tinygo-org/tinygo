//go:build tinygo.wasm && !wasm_unknown && !wasip2

package runtime

import (
	"unsafe"
)

// Implements __wasi_iovec_t.
type __wasi_iovec_t struct {
	buf    unsafe.Pointer
	bufLen uint
}

//go:wasmimport wasi_snapshot_preview1 fd_write
func fd_write(id uint32, iovs *__wasi_iovec_t, iovs_len uint, nwritten *uint) (errno uint)

// See:
// https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#-proc_exitrval-exitcode
//
//go:wasmimport wasi_snapshot_preview1 proc_exit
func proc_exit(exitcode uint32)

const (
	putcharBufferSize = 120
	stdout            = 1
)

// Using global variables to avoid heap allocation.
var (
	putcharBuffer        = [putcharBufferSize]byte{}
	putcharPosition uint = 0
	putcharIOVec         = __wasi_iovec_t{
		buf: unsafe.Pointer(&putcharBuffer[0]),
	}
	putcharNWritten uint
)

func putchar(c byte) {
	putcharBuffer[putcharPosition] = c
	putcharPosition++

	if c == '\n' || putcharPosition >= putcharBufferSize {
		putcharIOVec.bufLen = putcharPosition
		fd_write(stdout, &putcharIOVec, 1, &putcharNWritten)
		putcharPosition = 0
	}
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
	mono = nanotime()
	sec = mono / (1000 * 1000 * 1000)
	nsec = int32(mono - sec*(1000*1000*1000))
	return
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	proc_exit(uint32(code))
}

// TinyGo does not yet support any form of parallelism on WebAssembly, so these
// can be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

func hardwareRand() (n uint64, ok bool) {
	n |= uint64(libc_arc4random())
	n |= uint64(libc_arc4random()) << 32
	return n, true
}

// uint32_t arc4random(void);
//
//export arc4random
func libc_arc4random() uint32
