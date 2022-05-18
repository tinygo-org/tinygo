package main

import (
	"errors"
	"reflect"
	"unsafe"
)

type (
	myint    int
	myslice  []byte
	myslice2 []myint
	mychan   chan int
	myptr    *int
	point    struct {
		X int16
		Y int16
	}
	mystruct struct {
		n    int `foo:"bar"`
		some point
		zero struct{}
		buf  []byte
		Buf  []byte
	}
	linkedList struct {
		next *linkedList `description:"chain"`
		foo  int
	}
	selfref struct {
		x *selfref
	}
)

var (
	errorValue   = errors.New("test error")
	errorType    = reflect.TypeOf((*error)(nil)).Elem()
	stringerType = reflect.TypeOf((*interface {
		String() string
	})(nil)).Elem()
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
	// by embedding a 0-array func type in your struct, it is not comparable
	type doNotCompare [0]func()
	type notComparable struct {
		doNotCompare
		data *int32
	}
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
		mychan(zeroChan),
		// pointers
		new(int),
		new(error),
		&n,
		myptr(new(int)),
		// slices
		[]byte{1, 2, 3},
		make([]uint8, 2, 5),
		[]rune{3, 5},
		[]string{"xyz", "Z"},
		zeroSlice,
		[]byte{},
		[]float32{1, 1.32},
		[]float64{1, 1.64},
		[]complex64{1, 1.64 + 0.3i},
		[]complex128{1, 1.128 + 0.4i},
		myslice{5, 3, 11},
		// array
		[3]int64{5, 8, 2},
		[2]uint8{3, 5},
		// functions
		zeroFunc,
		emptyFunc,
		// maps
		zeroMap,
		map[string]int{},
		// structs
		struct{}{},
		struct{ error }{},
		struct {
			a uint8
			b int16
			c int8
		}{42, 321, 123},
		mystruct{5, point{-5, 3}, struct{}{}, []byte{'G', 'o'}, []byte{'X'}},
		&linkedList{
			foo: 42,
		},
		// interfaces
		[]interface{}{3, "str", -4 + 2.5i},
	} {
		showValue(reflect.ValueOf(v), "")
	}

	// Test reflect.New().
	newInt8 := reflect.New(reflect.TypeOf(int8(0)))
	newInt8.Elem().SetInt(5)
	newInt16 := reflect.New(reflect.TypeOf(int16(0)))
	newInt16.Elem().SetInt(-800)
	newInt32 := reflect.New(reflect.TypeOf(int32(0)))
	newInt32.Elem().SetInt(1e8)
	newInt64 := reflect.New(reflect.TypeOf(int64(0)))
	newInt64.Elem().SetInt(-1e12)
	newComplex128 := reflect.New(reflect.TypeOf(0 + 0i))
	newComplex128.Elem().SetComplex(-8 - 20e5i)
	for _, val := range []reflect.Value{newInt8, newInt16, newInt32, newInt64, newComplex128} {
		showValue(val, "")
	}

	// test sizes
	println("\nsizes:")
	for _, tc := range []struct {
		name string
		rt   reflect.Type
	}{
		{"int8", reflect.TypeOf(int8(0))},
		{"int16", reflect.TypeOf(int16(0))},
		{"int32", reflect.TypeOf(int32(0))},
		{"int64", reflect.TypeOf(int64(0))},
		{"uint8", reflect.TypeOf(uint8(0))},
		{"uint16", reflect.TypeOf(uint16(0))},
		{"uint32", reflect.TypeOf(uint32(0))},
		{"uint64", reflect.TypeOf(uint64(0))},
		{"float32", reflect.TypeOf(float32(0))},
		{"float64", reflect.TypeOf(float64(0))},
		{"complex64", reflect.TypeOf(complex64(0))},
		{"complex128", reflect.TypeOf(complex128(0))},
	} {
		println(tc.name, int(tc.rt.Size()), tc.rt.Bits())
	}
	assertSize(reflect.TypeOf(uintptr(0)).Size() == unsafe.Sizeof(uintptr(0)), "uintptr")
	assertSize(reflect.TypeOf("").Size() == unsafe.Sizeof(""), "string")
	assertSize(reflect.TypeOf(new(int)).Size() == unsafe.Sizeof(new(int)), "*int")
	assertSize(reflect.TypeOf(zeroFunc).Size() == unsafe.Sizeof(zeroFunc), "func()")

	// make sure embedding a zero-sized "not comparable" struct does not add size to a struct
	assertSize(reflect.TypeOf(doNotCompare{}).Size() == unsafe.Sizeof(doNotCompare{}), "[0]func()")
	assertSize(unsafe.Sizeof(notComparable{}) == unsafe.Sizeof((*int32)(nil)), "struct{[0]func(); *int32}")

	// Test that offset is correctly calculated.
	// This doesn't just test reflect but also (indirectly) that unsafe.Alignof
	// works correctly.
	s := struct {
		small1 byte
		big1   int64
		small2 byte
		big2   int64
	}{}
	st := reflect.TypeOf(s)
	println("offset for int64 matches:", st.Field(1).Offset-st.Field(0).Offset == uintptr(unsafe.Pointer(&s.big1))-uintptr(unsafe.Pointer(&s.small1)))
	println("offset for complex128 matches:", st.Field(3).Offset-st.Field(2).Offset == uintptr(unsafe.Pointer(&s.big2))-uintptr(unsafe.Pointer(&s.small2)))

	// SetBool
	rv := reflect.ValueOf(new(bool)).Elem()
	rv.SetBool(true)
	if rv.Bool() != true {
		panic("could not set bool with SetBool()")
	}

	// SetInt
	for _, v := range []interface{}{
		new(int),
		new(int8),
		new(int16),
		new(int32),
		new(int64),
	} {
		rv := reflect.ValueOf(v).Elem()
		rv.SetInt(99)
		if rv.Int() != 99 {
			panic("could not set integer with SetInt()")
		}
	}

	// SetUint
	for _, v := range []interface{}{
		new(uint),
		new(uint8),
		new(uint16),
		new(uint32),
		new(uint64),
		new(uintptr),
	} {
		rv := reflect.ValueOf(v).Elem()
		rv.SetUint(99)
		if rv.Uint() != 99 {
			panic("could not set integer with SetUint()")
		}
	}

	// SetFloat
	for _, v := range []interface{}{
		new(float32),
		new(float64),
	} {
		rv := reflect.ValueOf(v).Elem()
		rv.SetFloat(2.25)
		if rv.Float() != 2.25 {
			panic("could not set float with SetFloat()")
		}
	}

	// SetComplex
	for _, v := range []interface{}{
		new(complex64),
		new(complex128),
	} {
		rv := reflect.ValueOf(v).Elem()
		rv.SetComplex(3 + 2i)
		if rv.Complex() != 3+2i {
			panic("could not set complex with SetComplex()")
		}
	}

	// SetString
	rv = reflect.ValueOf(new(string)).Elem()
	rv.SetString("foo")
	if rv.String() != "foo" {
		panic("could not set string with SetString()")
	}

	// Set int
	rv = reflect.ValueOf(new(int)).Elem()
	rv.SetInt(33)
	rv.Set(reflect.ValueOf(22))
	if rv.Int() != 22 {
		panic("could not set int with Set()")
	}

	// Set uint8
	rv = reflect.ValueOf(new(uint8)).Elem()
	rv.SetUint(33)
	rv.Set(reflect.ValueOf(uint8(22)))
	if rv.Uint() != 22 {
		panic("could not set uint8 with Set()")
	}

	// Set string
	rv = reflect.ValueOf(new(string)).Elem()
	rv.SetString("foo")
	rv.Set(reflect.ValueOf("bar"))
	if rv.String() != "bar" {
		panic("could not set string with Set()")
	}

	// Set complex128
	rv = reflect.ValueOf(new(complex128)).Elem()
	rv.SetComplex(3 + 2i)
	rv.Set(reflect.ValueOf(4 + 8i))
	if rv.Complex() != 4+8i {
		panic("could not set complex128 with Set()")
	}

	// Set to slice
	rv = reflect.ValueOf([]int{3, 5})
	rv.Index(1).SetInt(7)
	if rv.Index(1).Int() != 7 {
		panic("could not set int in slice")
	}
	rv.Index(1).Set(reflect.ValueOf(8))
	if rv.Index(1).Int() != 8 {
		panic("could not set int in slice")
	}
	if rv.Len() != 2 || rv.Index(0).Int() != 3 {
		panic("slice was changed while setting part of it")
	}

	testAppendSlice()

	// Test types that are created in reflect and never created elsewhere in a
	// value-to-interface conversion.
	v := reflect.ValueOf(new(unreferencedType))
	switch v.Elem().Interface().(type) {
	case unreferencedType:
		println("type assertion succeeded for unreferenced type")
	default:
		println("type assertion failed (but should succeed)")
	}

	// Test type that is not referenced at all: not when creating the
	// reflect.Value (except through the field) and not with a type assert.
	// Previously this would result in a type assert failure because the Int()
	// method wasn't picked up.
	v = reflect.ValueOf(struct {
		X totallyUnreferencedType
	}{})
	if v.Field(0).Interface().(interface {
		Int() int
	}).Int() != 42 {
		println("could not call method on totally unreferenced type")
	}

	if reflect.TypeOf(new(myint)) != reflect.PtrTo(reflect.TypeOf(myint(0))) {
		println("PtrTo failed for type myint")
	}
	if reflect.TypeOf(new(myslice)) != reflect.PtrTo(reflect.TypeOf(make(myslice, 0))) {
		println("PtrTo failed for type myslice")
	}

	if reflect.TypeOf(errorValue).Implements(errorType) != true {
		println("errorValue.Implements(errorType) was false, expected true")
	}
	if reflect.TypeOf(errorValue).Implements(stringerType) != false {
		println("errorValue.Implements(errorType) was true, expected false")
	}

	println("\nstruct tags")
	TestStructTag()

	println("\nv.Interface() method")
	testInterfaceMethod()

	// Test reflect.DeepEqual.
	var selfref1, selfref2 selfref
	selfref1.x = &selfref1
	selfref2.x = &selfref2
	for i, tc := range []struct {
		v1, v2 interface{}
		equal  bool
	}{
		{int(5), int(5), true},
		{int(3), int(5), false},
		{int(5), uint(5), false},
		{struct {
			a int
			b string
		}{3, "x"}, struct {
			a int
			b string
		}{3, "x"}, true},
		{struct {
			a int
			b string
		}{3, "x"}, struct {
			a int
			b string
		}{3, "y"}, false},
		{selfref1, selfref2, true},
	} {
		result := reflect.DeepEqual(tc.v1, tc.v2)
		if result != tc.equal {
			if tc.equal {
				println("reflect.DeepEqual() test", i, "not equal while it should be")
			} else {
				println("reflect.DeepEqual() test", i, "equal while it should not be")
			}
		}
	}
}

func emptyFunc() {
}

func showValue(rv reflect.Value, indent string) {
	rt := rv.Type()
	if rt.Kind() != rv.Kind() {
		panic("type kind is different from value kind")
	}
	print(indent+"reflect type: ", rt.Kind().String())
	if rv.CanSet() {
		print(" settable=true")
	}
	if rv.CanAddr() {
		print(" addrable=true")
	}
	if !rt.Comparable() {
		print(" comparable=false")
	}
	println()
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
			showValue(rv.Index(i), indent+"  ")
		}
	case reflect.UnsafePointer:
		println(indent+"  pointer:", rv.Pointer() != 0)
	case reflect.Array:
		println(indent+"  array:", rt.Len(), rt.Elem().Kind().String(), int(rt.Size()))
		for i := 0; i < rv.Len(); i++ {
			showValue(rv.Index(i), indent+"  ")
		}
	case reflect.Chan:
		println(indent+"  chan:", rt.Elem().Kind().String())
		println(indent+"  nil:", rv.IsNil())
	case reflect.Func:
		println(indent + "  func")
		println(indent+"  nil:", rv.IsNil())
	case reflect.Interface:
		println(indent + "  interface")
		println(indent+"  nil:", rv.IsNil())
		if !rv.IsNil() {
			showValue(rv.Elem(), indent+"  ")
		}
	case reflect.Map:
		println(indent + "  map")
		println(indent+"  nil:", rv.IsNil())
	case reflect.Ptr:
		println(indent+"  pointer:", rv.Pointer() != 0, rt.Elem().Kind().String())
		println(indent+"  nil:", rv.IsNil())
		if !rv.IsNil() {
			showValue(rv.Elem(), indent+"  ")
		}
	case reflect.Slice:
		println(indent+"  slice:", rt.Elem().Kind().String(), rv.Len(), rv.Cap())
		println(indent+"  pointer:", rv.Pointer() != 0)
		println(indent+"  nil:", rv.IsNil())
		for i := 0; i < rv.Len(); i++ {
			println(indent+"  indexing:", i)
			showValue(rv.Index(i), indent+"  ")
		}
	case reflect.Struct:
		println(indent+"  struct:", rt.NumField())
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			println(indent+"  field:", i, field.Name)
			println(indent+"  tag:", field.Tag)
			println(indent+"  embedded:", field.Anonymous)
			println(indent+"  exported:", field.IsExported())
			showValue(rv.Field(i), indent+"  ")
		}
	default:
		println(indent + "  unknown type kind!")
	}
}

func assertSize(ok bool, typ string) {
	if !ok {
		panic("size mismatch for type " + typ)
	}
}

// Test whether appending to a slice is equivalent between reflect and native
// slice append.
func testAppendSlice() {
	for i := 0; i < 100; i++ {
		dst := makeRandomSlice(i)
		src := makeRandomSlice(i)
		result1 := append(dst, src...)
		result2 := reflect.AppendSlice(reflect.ValueOf(dst), reflect.ValueOf(src)).Interface().([]uint32)
		if !sliceEqual(result1, result2) {
			println("slice: mismatch after runtime.SliceAppend with", len(dst), cap(dst), len(src), cap(src))
		}
	}
}

func makeRandomSlice(max int) []uint32 {
	cap := randuint32() % uint32(max+1)
	len := randuint32() % (cap + 1)
	s := make([]uint32, len, cap)
	for i := uint32(0); i < len; i++ {
		s[i] = randuint32()
	}
	return s
}

func sliceEqual(s1, s2 []uint32) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, val := range s1 {
		if s2[i] != val {
			return false
		}
	}
	// Note: can't compare cap because the Go implementation has a different
	// behavior between the built-in append function and
	// reflect.AppendSlice.
	return true
}

type unreferencedType int

type totallyUnreferencedType int

func (totallyUnreferencedType) Int() int {
	return 42
}

func TestStructTag() {
	type S struct {
		F string `species:"gopher" color:"blue"`
	}

	s := S{}
	st := reflect.TypeOf(s)
	field := st.Field(0)
	println(field.Tag.Get("color"), field.Tag.Get("species"))
}

// Test Interface() call: it should never return an interface itself.
func testInterfaceMethod() {
	v := reflect.ValueOf(struct{ X interface{} }{X: 5})
	println("kind:", v.Field(0).Kind().String())
	itf := v.Field(0).Interface()
	switch n := itf.(type) {
	case int:
		println("int", n) // correct
	default:
		println("something else") // incorrect
	}
}

var xorshift32State uint32 = 1

func xorshift32(x uint32) uint32 {
	// Algorithm "xor" from p. 4 of Marsaglia, "Xorshift RNGs"
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	return x
}

func randuint32() uint32 {
	xorshift32State = xorshift32(xorshift32State)
	return xorshift32State
}
