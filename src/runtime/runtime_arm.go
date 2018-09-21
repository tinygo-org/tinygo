/// +build arm

package runtime

import (
	"device/arm"
	"unsafe"
)

//go:extern _sbss
var _sbss unsafe.Pointer

//go:extern _ebss
var _ebss unsafe.Pointer

//go:extern _sdata
var _sdata unsafe.Pointer

//go:extern _sidata
var _sidata unsafe.Pointer

//go:extern _edata
var _edata unsafe.Pointer

func preinit() {
	// Initialize .bss: zero-initialized global variables.
	ptr := uintptr(unsafe.Pointer(&_sbss))
	for ptr != uintptr(unsafe.Pointer(&_ebss)) {
		*(*uint32)(unsafe.Pointer(ptr)) = 0
		ptr += 4
	}

	// Initialize .data: global variables initialized from flash.
	src := uintptr(unsafe.Pointer(&_sidata))
	dst := uintptr(unsafe.Pointer(&_sdata))
	for dst != uintptr(unsafe.Pointer(&_edata)) {
		*(*uint32)(unsafe.Pointer(dst)) = *(*uint32)(unsafe.Pointer(src))
		dst += 4
		src += 4
	}
}

func postinit() {
}

func abort() {
	for {
		arm.Asm("wfi")
	}
}

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}
