//go:build tinygo.riscv64
// +build tinygo.riscv64

package runtime

import "unsafe"

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

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := unsafe.Pointer(&_sbss)
	for ptr != unsafe.Pointer(&_ebss) {
		*(*uint64)(ptr) = 0
		ptr = unsafe.Pointer(uintptr(ptr) + 8)
	}

	// Initialize .data: global variables initialized from flash.
	src := unsafe.Pointer(&_sidata)
	dst := unsafe.Pointer(&_sdata)
	for dst != unsafe.Pointer(&_edata) {
		*(*uint64)(dst) = *(*uint64)(src)
		dst = unsafe.Pointer(uintptr(dst) + 8)
		src = unsafe.Pointer(uintptr(src) + 8)
	}
}
