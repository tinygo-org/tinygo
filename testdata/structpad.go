package main

// Structs that end in a zero-sized field are padded as gc does (https://go.dev/issue/9401).

import (
	"reflect"
	"unsafe"
)

type endsEmpty struct {
	data [4096]byte
	end  struct{}
}

type endsZeroArray struct {
	b   byte
	end [0]byte
}

type onlyEmpty struct {
	a struct{}
	b [0]byte
}

type nested struct {
	inner endsEmpty
	end   struct{}
}

type emptyFirst struct {
	end  struct{}
	data [3]byte
}

type threeInts struct {
	a, b, c int32
	_       struct{}
}

//go:noinline
func bumpZeroArray(v endsZeroArray) endsZeroArray {
	v.b++
	return v
}

//go:noinline
func sumInts(v threeInts) threeInts {
	v.a += v.b + v.c
	return v
}

func main() {
	println("endsEmpty:", int(unsafe.Sizeof(endsEmpty{})))
	println("endsZeroArray:", int(unsafe.Sizeof(endsZeroArray{})))
	println("onlyEmpty:", int(unsafe.Sizeof(onlyEmpty{})))
	println("nested:", int(unsafe.Sizeof(nested{})))
	println("emptyFirst:", int(unsafe.Sizeof(emptyFirst{})))
	println("array of endsEmpty:", int(unsafe.Sizeof([2]endsEmpty{})))

	// The address of the final field must stay inside the object.
	p := new(endsEmpty)
	start := uintptr(unsafe.Pointer(p))
	end := uintptr(unsafe.Pointer(&p.end))
	println("end field inside object:", end-start < unsafe.Sizeof(*p))

	// Pass and return by value.
	println("by value:", bumpZeroArray(endsZeroArray{b: 41}).b, sumInts(threeInts{a: 1, b: 2, c: 3}).a)

	// reflect agrees with the compiler.
	println("reflect sizes:",
		reflect.TypeOf(endsEmpty{}).Size() == unsafe.Sizeof(endsEmpty{}),
		reflect.TypeOf(endsZeroArray{}).Size() == unsafe.Sizeof(endsZeroArray{}),
		reflect.TypeOf(onlyEmpty{}).Size() == unsafe.Sizeof(onlyEmpty{}),
		reflect.TypeOf(nested{}).Size() == unsafe.Sizeof(nested{}),
		reflect.TypeOf(threeInts{}).Size() == unsafe.Sizeof(threeInts{}))
	built := reflect.StructOf([]reflect.StructField{
		{Name: "B", Type: reflect.TypeOf(byte(0))},
		{Name: "End", Type: reflect.TypeOf([0]byte{})},
	})
	println("StructOf size:", built.Size() == unsafe.Sizeof(endsZeroArray{}))
}
