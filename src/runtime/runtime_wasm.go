// +build wasm

package runtime

import (
	"unsafe"
)

type timeUnit float64 // time in milliseconds, just like Date.now() in JavaScript

const tickMicros = 1000000

//go:export io_get_stdout
func io_get_stdout() int32

//go:export resource_write
func resource_write(id int32, ptr *uint8, len int32) int32

var stdout int32

func init() {
	stdout = io_get_stdout()
}

//go:export _start
func _start() {
	initAll()
}

//go:export cwa_main
func cwa_main() {
	initAll() // _start is not called by olin/cwa so has to be called here
	callMain()
}

func putchar(c byte) {
	resource_write(stdout, &c, 1)
}

var handleEvent func()

//go:linkname setEventHandler syscall/js.setEventHandler
func setEventHandler(fn func()) {
	handleEvent = fn
}

//go:export resume
func resume() {
	handleEvent()
}

//go:export go_scheduler
func go_scheduler() {
	scheduler()
}

const asyncScheduler = true

// This function is called by the scheduler.
// Schedule a call to runtime.scheduler, do not actually sleep.
//go:export runtime.sleepTicks
func sleepTicks(d timeUnit)

//go:export runtime.ticks
func ticks() timeUnit

// Abort executes the wasm 'unreachable' instruction.
func abort() {
	trap()
}

//go:export memset
func memset(ptr unsafe.Pointer, c byte, size uintptr) unsafe.Pointer {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = c
	}
	return ptr
}

// Implement memmove for LLVM and compiler-rt.
//go:export memmove
func libc_memmove(dst, src unsafe.Pointer, size uintptr) unsafe.Pointer {
	memmove(dst, src, size)
	return dst
}

// Implement memcpy for LLVM and compiler-rt.
//go:export memcpy
func libc_memcpy(dst, src unsafe.Pointer, size uintptr) unsafe.Pointer {
	memcpy(dst, src, size)
	return dst
}
