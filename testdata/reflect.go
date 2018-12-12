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
	} {
		showValue(v)
	}
}

func showValue(v interface{}) {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	if reflect.TypeOf(v) != rt {
		panic("direct TypeOf() is different from ValueOf().Type()")
	}
	println("reflect type:", rt.Kind().String())
	switch rt.Kind() {
	case reflect.Bool:
		println("  bool:", rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		println("  int:", rv.Int())
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		println("  uint:", rv.Uint())
	case reflect.Float32, reflect.Float64:
		println("  float:", rv.Float())
	case reflect.Complex64, reflect.Complex128:
		println("  complex:", rv.Complex())
	case reflect.String:
		println("  string:", rv.String())
	case reflect.UnsafePointer:
		println("  pointer:", rv.Pointer() != 0)
	default:
		println("  unknown type kind!")
	}
}
