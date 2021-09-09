package main

// Test changes to the language introduced in Go 1.17.
// For details, see: https://tip.golang.org/doc/go1.17#language
// These tests should be merged into the regular slice tests once Go 1.17 is the
// minimun Go version for TinyGo.

import "unsafe"

func Add32(p unsafe.Pointer, len int) unsafe.Pointer {
	return unsafe.Add(p, len)
}

func Add64(p unsafe.Pointer, len int64) unsafe.Pointer {
	return unsafe.Add(p, len)
}

func SliceToArray(s []int) *[4]int {
	return (*[4]int)(s)
}

func SliceToArrayConst() *[4]int {
	s := make([]int, 6)
	return (*[4]int)(s)
}

func SliceInt(ptr *int, len int) []int {
	return unsafe.Slice(ptr, len)
}

func SliceUint16(ptr *byte, len uint16) []byte {
	return unsafe.Slice(ptr, len)
}

func SliceUint64(ptr *int, len uint64) []int {
	return unsafe.Slice(ptr, len)
}

func SliceInt64(ptr *int, len int64) []int {
	return unsafe.Slice(ptr, len)
}
