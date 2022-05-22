// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux
// +build linux

package runtime

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

// physPageSize is the size in bytes of the OS's physical pages.
var physPageSize uintptr

const (
	_AT_NULL   = 0  // End of vector
	_AT_PAGESZ = 6  // System physical page size
	_AT_HWCAP  = 16 // hardware capability bit vector
	_AT_RANDOM = 25 // introduced in 2.6.29
	_AT_HWCAP2 = 26 // hardware capability bit vector 2
)

// argv_index: borrowed from upstream runtime1.go
func argv_index(argv **byte, i int32) *byte {
	return *(**byte)(add(unsafe.Pointer(argv), uintptr(i)*ptrSize))
}

// add: borrowed from upstream stubs.go
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

// TODO: is this always correct?
const ptrSize = (TargetBits / 8)

func sysargs(argc int32, argv **byte) {
	n := argc + 1

	// skip over argv, envp to get to auxv
	for argv_index(argv, n) != nil {
		n++
	}

	// skip NULL separator
	n++

	// now argv+n is auxv
	auxv := (*[1 << 28]uintptr)(add(unsafe.Pointer(argv), uintptr(n)*ptrSize))
	if sysauxv(auxv[:]) != 0 {
		return
	}
	// TODO: fall back to reading from /proc, as upstream does
}

// startupRandomData holds random bytes initialized at startup. These come from
// the ELF AT_RANDOM auxiliary vector.
var startupRandomData []byte

func sysauxv(auxv []uintptr) int {
	var i int
	for ; auxv[i] != _AT_NULL; i += 2 {
		tag, val := auxv[i], auxv[i+1]
		switch tag {
		case _AT_RANDOM:
			// The kernel provides a pointer to 16-bytes
			// worth of random data.
			startupRandomData = (*[16]byte)(unsafe.Pointer(val))[:]

		case _AT_PAGESZ:
			physPageSize = val
		}

		// TODO: see if we can use any of the goodies here
		//archauxv(tag, val)
		//vdsoauxv(tag, val)
	}
	return i / 2
}

//go:linkname syscall_Getpagesize syscall.Getpagesize
func syscall_Getpagesize() int { return int(physPageSize) }
