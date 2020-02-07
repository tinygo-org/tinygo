// +build wasm

package runtime

import "unsafe"

type timeUnit float64 // time in milliseconds, just like Date.now() in JavaScript

const tickMicros = 1000000

// Implements __wasi_ciovec_t and __wasi_iovec_t.
type wasiIOVec struct {
	buf    unsafe.Pointer
	bufLen uint
}

//go:wasm-module wasi_unstable
//export fd_write
func fd_write(id uint32, iovs *wasiIOVec, iovs_len uint, nwritten *uint) (errno uint)

func postinit() {}

//export _start
func _start() {
	run()
}

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

var handleEvent func()

//go:linkname setEventHandler syscall/js.setEventHandler
func setEventHandler(fn func()) {
	handleEvent = fn
}

//go:export resume
func resume() {
	go func() {
		handleEvent()
	}()
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
