package main

import (
	"reflect"
	"unsafe"
)

type myint int

func main() {
	println("matching types")
	println(reflect.TypeOf(int(3)) == reflect.TypeOf(int(5)))
	println(reflect.TypeOf(int(3)) == reflect.TypeOf(uint(5)))
	println(reflect.TypeOf(myint(3)) == reflect.TypeOf(int(5)))

	println("\nvalues of interfaces")
	for _, v := range []interface{}{
		true,
		false,
		int(2000),
		int(-2000),
		uint(2000),
		int8(-3),
		int8(3),
		uint8(200),
		int16(-300),
		int16(300),
		uint16(50000),
		int32(7 << 20),
		int32(-7 << 20),
		uint32(7 << 20),
		int64(9 << 40),
		int64(-9 << 40),
		uint64(9 << 40),
		uintptr(12345),
		float32(3.14),
		float64(3.14),
		complex64(1.2 + 0.3i),
		complex128(1.3 + 0.4i),
		"foo",
		unsafe.Pointer(new(int)),
		[]byte{1, 2, 3},
		make([]uint8, 2, 5),
		[]rune{3, 5},
		[]string{"xyz", "Z"},
	} {
		showValue(v, "")
	}
}

func showValue(v interface{}, indent string) {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	if reflect.TypeOf(v) != rt {
		panic("direct TypeOf() is different from ValueOf().Type()")
	}
	if rt.Kind() != rv.Kind() {
		panic("type kind is different from value kind")
	}
	if reflect.ValueOf(rv.Interface()) != rv {
		panic("reflect.ValueOf(Value.Interface()) did not return the same value")
	}
	println(indent+"reflect type:", rt.Kind().String())
	switch rt.Kind() {
	case reflect.Bool:
		println(indent+"  bool:", rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		println(indent+"  int:", rv.Int())
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		println(indent+"  uint:", rv.Uint())
	case reflect.Float32, reflect.Float64:
		println(indent+"  float:", rv.Float())
	case reflect.Complex64, reflect.Complex128:
		println(indent+"  complex:", rv.Complex())
	case reflect.String:
		println(indent+"  string:", rv.String(), rv.Len())
		for i := 0; i < rv.Len(); i++ {
			showValue(rv.Index(i).Interface(), indent+"  ")
		}
	case reflect.UnsafePointer:
		println(indent+"  pointer:", rv.Pointer() != 0)
	case reflect.Slice:
		println(indent+"  slice:", rt.Elem().Kind().String(), rv.Len(), rv.Cap())
		for i := 0; i < rv.Len(); i++ {
			println(indent+"  indexing:", i)
			showValue(rv.Index(i).Interface(), indent+"  ")
		}
	default:
		println(indent + "  unknown type kind!")
	}
}
