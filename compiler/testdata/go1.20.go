package main

import "unsafe"

func unsafeSliceData(s []int) *int {
	return unsafe.SliceData(s)
}

func unsafeString(ptr *byte, len int16) string {
	return unsafe.String(ptr, len)
}

func unsafeStringData(s string) *byte {
	return unsafe.StringData(s)
}
