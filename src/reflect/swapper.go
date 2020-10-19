package reflect

import "unsafe"

// Some of code here has been copied from the Go sources:
//   https://github.com/golang/go/blob/go1.15.2/src/reflect/swapper.go
// It has the following copyright note:
//
// 		Copyright 2016 The Go Authors. All rights reserved.
// 		Use of this source code is governed by a BSD-style
// 		license that can be found in the LICENSE file.

func Swapper(slice interface{}) func(i, j int) {
	v := ValueOf(slice)
	if v.Kind() != Slice {
		panic(&ValueError{Method: "Swapper"})
	}

	// Just return Nop func if nothing to swap.
	if v.Len() < 2 {
		return func(i, j int) {}
	}

	typ := v.Type().Elem()
	size := typ.Size()

	header := (*SliceHeader)(v.value)
	tmp := unsafe.Pointer(&make([]byte, size)[0])

	return func(i, j int) {
		if uint(i) >= uint(header.Len) || uint(j) >= uint(header.Len) {
			panic("reflect: slice index out of range")
		}
		val1 := unsafe.Pointer(header.Data + uintptr(i)*size)
		val2 := unsafe.Pointer(header.Data + uintptr(j)*size)
		memcpy(tmp, val1, size)
		memcpy(val1, val2, size)
		memcpy(val2, tmp, size)
	}
}
