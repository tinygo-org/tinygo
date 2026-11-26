//go:build wasip3

package runtime

import (
	"unsafe"
)

const putcharBufferSize = 120

// Using global variables to avoid heap allocation.
var (
	putcharBuffer        = [putcharBufferSize]byte{}
	putcharPosition uint = 0
)

// ssize_t write(int fd, const void *buf, size_t count);
//
//export write
func libc_write(fd int32, buf unsafe.Pointer, count uint) int

func putchar(c byte) {
	putcharBuffer[putcharPosition] = c
	putcharPosition++
	if c == '\n' || putcharPosition >= putcharBufferSize {
		libc_write(1, unsafe.Pointer(&putcharBuffer[0]), putcharPosition) // error return ignored; can't do anything anyways
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

type libcSystemClockInstant struct {
	seconds     int64
	nanoseconds uint32
}

// void system_clock_now(system_clock_instant_t *ret);
//
//export system_clock_now
func system_clock_now(ret *libcSystemClockInstant)

//go:linkname now time.now
func now() (sec int64, nsec int32, mono int64) {
	var t libcSystemClockInstant
	system_clock_now(&t)
	sec = t.seconds
	nsec = int32(t.nanoseconds)
	mono = int64(monotonic_clock_now())
	return
}

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

// void _Exit(int status);
//
//export _Exit
func libc_exit(status int32)

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	libc_exit(int32(code))
}

func mainReturnExit() {
	// WASIp3 does not use _start, instead it uses _initialize and a custom
	// WASIp3-specific main function. So this should never be called in
	// practice.
	runtimeFatal("unreachable: _start was called")
}

// TinyGo does not yet support any form of parallelism on WebAssembly, so these
// can be left empty.

//go:linkname procPin sync/atomic.runtime_procPin
func procPin() {
}

//go:linkname procUnpin sync/atomic.runtime_procUnpin
func procUnpin() {
}

// get-random-u64 does not allocate. get-random-bytes (used by arc4random)
// calls cabi_realloc, which is not usable yet while the runtime seeds itself.
//
//go:wasmimport wasi:random/random@0.3.2 get-random-u64
func wasiRandomU64() uint64

func hardwareRand() (n uint64, ok bool) {
	return wasiRandomU64(), true
}

func libc_errno_location() *int32 {
	// CGo is unavailable, so this function should be unreachable.
	runtimeFatal("runtime: no cgo errno")
	return nil
}

// WASIp3 does not use wasi-libc's stdio.
func flushStdio() {}
