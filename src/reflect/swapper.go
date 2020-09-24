package reflect

import "unsafe"

// Most code here has been copied from the Go sources:
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

	// Fast path for slices of size 0 and 1. Nothing to swap.
	switch v.Len() {
	case 0:
		return func(i, j int) { panic("reflect: slice index out of range") }
	case 1:
		return func(i, j int) {
			if i != 0 || j != 0 {
				panic("reflect: slice index out of range")
			}
		}
	}

	typ := v.Type().Elem()
	size := typ.Size()

	// Some common & small cases, without using memmove:
	if typ.Kind() == String {
		ss := *(*[]string)((v.value))
		return func(i, j int) {
			ss[i], ss[j] = ss[j], ss[i]
		}
	} else {
		switch size {
		case 8:
			is := *(*[]int64)(v.value)
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 4:
			is := *(*[]int32)(v.value)
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 2:
			is := *(*[]int16)(v.value)
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 1:
			is := *(*[]int8)(v.value)
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		}
	}

	header := (*SliceHeader)(v.value)
	tmp := unsafe.Pointer(&make([]byte, size)[0])

	return func(i, j int) {
		if uint(i) >= uint(header.Len) || uint(j) >= uint(header.Len) {
			panic("reflect: slice index out of range")
		}
		val1 := unsafe.Pointer(header.Data + uintptr(i)*size)
		val2 := unsafe.Pointer(header.Data + uintptr(j)*size)
		memmove(tmp, val1, size)
		memmove(val1, val2, size)
		memmove(val2, tmp, size)
	}
}
