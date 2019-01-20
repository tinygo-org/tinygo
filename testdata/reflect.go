package main

import (
	"reflect"
	"unsafe"
)

type (
	myint    int
	myslice  []byte
	myslice2 []myint
)

func main() {
	println("matching types")
	println(reflect.TypeOf(int(3)) == reflect.TypeOf(int(5)))
	println(reflect.TypeOf(int(3)) == reflect.TypeOf(uint(5)))
	println(reflect.TypeOf(myint(3)) == reflect.TypeOf(int(5)))
	println(reflect.TypeOf(myslice{}) == reflect.TypeOf([]byte{}))
	println(reflect.TypeOf(myslice2{}) == reflect.TypeOf([]myint{}))
	println(reflect.TypeOf(myslice2{}) == reflect.TypeOf([]int{}))

	println("\nvalues of interfaces")
	var zeroSlice []byte
	var zeroFunc func()
	var zeroMap map[string]int
	var zeroChan chan int
	n := 42
	for _, v := range []interface{}{
		// basic types
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
		myint(32),
		"foo",
		unsafe.Pointer(new(int)),
		// channels
		zeroChan,
		// pointers
		new(int),
		new(error),
		&n,
		// slices
		[]byte{1, 2, 3},
		make([]uint8, 2, 5),
		[]rune{3, 5},
		[]string{"xyz", "Z"},
		zeroSlice,
		[]byte{},
		// array
		[4]int{1, 2, 3, 4},
		// functions
		zeroFunc,
		emptyFunc,
		// maps
		zeroMap,
		map[string]int{},
		// structs
		struct{}{},
		struct{ error }{},
	} {
		showValue(v, "")
	}

	// test sizes
	println("\nsizes:")
	println("int8", int(reflect.TypeOf(int8(0)).Size()))
	println("int16", int(reflect.TypeOf(int16(0)).Size()))
	println("int32", int(reflect.TypeOf(int32(0)).Size()))
	println("int64", int(reflect.TypeOf(int64(0)).Size()))
	println("uint8", int(reflect.TypeOf(uint8(0)).Size()))
	println("uint16", int(reflect.TypeOf(uint16(0)).Size()))
	println("uint32", int(reflect.TypeOf(uint32(0)).Size()))
	println("uint64", int(reflect.TypeOf(uint64(0)).Size()))
	println("float32", int(reflect.TypeOf(float32(0)).Size()))
	println("float64", int(reflect.TypeOf(float64(0)).Size()))
	println("complex64", int(reflect.TypeOf(complex64(0)).Size()))
	println("complex128", int(reflect.TypeOf(complex128(0)).Size()))
	assertSize(reflect.TypeOf(uintptr(0)).Size() == unsafe.Sizeof(uintptr(0)), "uintptr")
	assertSize(reflect.TypeOf("").Size() == unsafe.Sizeof(""), "string")
	assertSize(reflect.TypeOf(new(int)).Size() == unsafe.Sizeof(new(int)), "*int")
}

func emptyFunc() {
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
	case reflect.Array:
		println(indent + "  array")
	case reflect.Chan:
		println(indent+"  chan:", rt.Elem().Kind().String())
		println(indent+"  nil:", rv.IsNil())
	case reflect.Func:
		println(indent + "  func")
		println(indent+"  nil:", rv.IsNil())
	case reflect.Map:
		println(indent + "  map")
		println(indent+"  nil:", rv.IsNil())
	case reflect.Ptr:
		println(indent+"  pointer:", rv.Pointer() != 0, rt.Elem().Kind().String())
		println(indent+"  nil:", rv.IsNil())
	case reflect.Slice:
		println(indent+"  slice:", rt.Elem().Kind().String(), rv.Len(), rv.Cap())
		println(indent+"  pointer:", rv.Pointer() != 0)
		println(indent+"  nil:", rv.IsNil())
		for i := 0; i < rv.Len(); i++ {
			println(indent+"  indexing:", i)
			showValue(rv.Index(i).Interface(), indent+"  ")
		}
	case reflect.Struct:
		println(indent + "  struct")
	default:
		println(indent + "  unknown type kind!")
	}
}

func assertSize(ok bool, typ string) {
	if !ok {
		panic("size mismatch for type " + typ)
	}
}
