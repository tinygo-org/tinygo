// +build arm7tdmi

package runtime

import (
	_ "runtime/interrupt" // make sure the interrupt handler is defined
	"unsafe"
)

type timeUnit int64

const tickMicros = 1

func putchar(c byte) {
	// dummy, TODO
}

//go:extern _sbss
var _sbss [0]byte

//go:extern _ebss
var _ebss [0]byte

//go:extern _sdata
var _sdata [0]byte

//go:extern _sidata
var _sidata [0]byte

//go:extern _edata
var _edata [0]byte

// Entry point for Go. Initialize all packages and call main.main().
//go:export main
func main() {
	// Initialize .data and .bss sections.
	preinit()

	// Run initializers of all packages.
	initAll()

	// Compiler-generated call to main.main().
	go callMain()

	// Run the scheduler.
	scheduler()
}

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint32)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 4)
	}

	// Initialize .data: global variables initialized from flash.
	src := unsafe.Pointer(&_sidata)
	dst := unsafe.Pointer(&_sdata)
	for dst != unsafe.Pointer(&_edata) {
		*(*uint32)(dst) = *(*uint32)(src)
		dst = unsafe.Pointer(uintptr(dst) + 4)
		src = unsafe.Pointer(uintptr(src) + 4)
	}
}

func ticks() timeUnit {
	// TODO
	return 0
}

const asyncScheduler = false

func sleepTicks(d timeUnit) {
	// TODO
}

func abort() {
	// TODO
	for {
	}
}

// Implement memset for LLVM and compiler-rt.
//go:export memset
func libc_memset(ptr unsafe.Pointer, c byte, size uintptr) {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = c
	}
}

// Implement memmove for LLVM and compiler-rt.
//go:export memmove
func libc_memmove(dst, src unsafe.Pointer, size uintptr) {
	memmove(dst, src, size)
}
