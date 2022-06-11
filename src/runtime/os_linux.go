//go:build linux && !baremetal && !nintendoswitch && !wasi
// +build linux,!baremetal,!nintendoswitch,!wasi

package runtime

// This file is for systems that are _actually_ Linux (not systems that pretend
// to be Linux, like baremetal systems).

import "unsafe"

const GOOS = "linux"

const (
	// See https://github.com/torvalds/linux/blob/master/include/uapi/asm-generic/mman-common.h
	flag_PROT_READ     = 0x1
	flag_PROT_WRITE    = 0x2
	flag_MAP_PRIVATE   = 0x2
	flag_MAP_ANONYMOUS = 0x20
)

// Source: https://github.com/torvalds/linux/blob/master/include/uapi/linux/time.h
const (
	clock_REALTIME      = 0
	clock_MONOTONIC_RAW = 4
)

//go:extern _edata
var globalsStartSymbol [0]byte

//go:extern _end
var globalsEndSymbol [0]byte

// markGlobals marks all globals, which are reachable by definition.
//
// This implementation marks all globals conservatively and assumes it can use
// linker-defined symbols for the start and end of the .data section.
func markGlobals() {
	start := uintptr(unsafe.Pointer(&globalsStartSymbol))
	end := uintptr(unsafe.Pointer(&globalsEndSymbol))
	start = (start + unsafe.Alignof(uintptr(0)) - 1) &^ (unsafe.Alignof(uintptr(0)) - 1) // align on word boundary
	markRoots(start, end)
}

//export getpagesize
func libc_getpagesize() int

//go:linkname syscall_Getpagesize syscall.Getpagesize
func syscall_Getpagesize() int {
	return libc_getpagesize()
}
