package main

import (
	"errors"
	"reflect"
	"strconv"
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
		n    int   `foo:"bar"`
		some point "some\x00tag"
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
		struct{ A, B uintptr }{2, 3},
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
	assertSize(reflect.TypeOf(zeroChan).Size() == unsafe.Sizeof(zeroChan), "chan int")
	assertSize(reflect.TypeOf(zeroMap).Size() == unsafe.Sizeof(zeroMap), "map[string]int")

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

	println("\ninterface implements")
	testImplements()

	println("\nalignment / offset:")
	v2 := struct {
		noCompare [0]func()
		data      byte
	}{}
	println("struct{[0]func(); byte}:", unsafe.Offsetof(v2.data) == uintptr(unsafe.Pointer(&v2.data))-uintptr(unsafe.Pointer(&v2)))

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

	// Test for issue #3794: reflect MapIter.Key() should return a value with
	// interface kind for map[interface{}] keys, not the underlying concrete kind.
	{
		m := make(map[interface{}]int)
		m[1] = 2
		m["hello"] = 3
		rv := reflect.ValueOf(m)
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key()
			if k.Kind() != reflect.Interface {
				println("FAIL #3794: expected interface kind, got", k.Kind().String())
				break
			}
		}
		println("reflect map interface key ok")
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
	if !rv.CanInterface() {
		print(" caninterface=false")
	}
	if !rt.Comparable() {
		print(" comparable=false")
	}
	if name := rt.Name(); name != "" {
		print(" name=", name)
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
		println(indent+"  NumMethod:", rv.NumMethod())
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
			println(indent+"  pkg:", field.PkgPath)
			println(indent+"  tag:", strconv.Quote(string(field.Tag)))
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

// Types for interface Implements/AssignableTo tests.

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

type Closer interface {
	Close() error
}

type ReadCloser interface {
	Read(p []byte) (n int, err error)
	Close() error
}

type myReader struct{}

func (myReader) Read(p []byte) (int, error) { return 0, nil }

type myWriter struct{}

func (*myWriter) Write(p []byte) (int, error) { return 0, nil }

type myReadWriter struct{}

func (myReadWriter) Read(p []byte) (int, error)   { return 0, nil }
func (*myReadWriter) Write(p []byte) (int, error)  { return 0, nil }

type myStringer struct{}

func (myStringer) String() string { return "mystringer" }

type myErrorStringer struct{}

func (myErrorStringer) Error() string  { return "err" }
func (myErrorStringer) String() string { return "str" }

// Interface with unexported method (from upstream set_test.go).
type exprLike interface {
	Pos() int
	End() int
	exprNode()
}

type notAnExpr struct{}

func (notAnExpr) Pos() int  { return 0 }
func (notAnExpr) End() int  { return 0 }
func (notAnExpr) exprNode() {}

// Named types for assignability tests (from upstream set_test.go).
type IntPtr *int
type IntPtr1 *int
type Ch <-chan interface{}

func testImplements() {
	readerType := reflect.TypeOf((*Reader)(nil)).Elem()
	writerType := reflect.TypeOf((*Writer)(nil)).Elem()
	readWriterType := reflect.TypeOf((*ReadWriter)(nil)).Elem()
	closerType := reflect.TypeOf((*Closer)(nil)).Elem()
	readCloserType := reflect.TypeOf((*ReadCloser)(nil)).Elem()
	emptyItf := reflect.TypeOf((*interface{})(nil)).Elem()

	// --- Concrete type implements interface ---
	println("concrete implements:")

	// myReader has value receiver Read → implements Reader
	println("myReader → Reader:", reflect.TypeOf(myReader{}).Implements(readerType))           // true
	println("*myReader → Reader:", reflect.TypeOf(new(myReader)).Elem().Implements(readerType)) // true (value method in pointer set)

	// myWriter has pointer receiver Write → only *myWriter implements Writer
	println("myWriter → Writer:", reflect.TypeOf(myWriter{}).Implements(writerType))    // false (pointer receiver)
	println("*myWriter → Writer:", reflect.TypeOf(&myWriter{}).Implements(writerType))  // true

	// myReadWriter: Read on value, Write on pointer
	println("myReadWriter → Reader:", reflect.TypeOf(myReadWriter{}).Implements(readerType))      // true
	println("myReadWriter → Writer:", reflect.TypeOf(myReadWriter{}).Implements(writerType))      // false (Write is ptr recv)
	println("myReadWriter → ReadWriter:", reflect.TypeOf(myReadWriter{}).Implements(readWriterType)) // false
	println("*myReadWriter → Reader:", reflect.TypeOf(&myReadWriter{}).Implements(readerType))       // true
	println("*myReadWriter → Writer:", reflect.TypeOf(&myReadWriter{}).Implements(writerType))       // true
	println("*myReadWriter → ReadWriter:", reflect.TypeOf(&myReadWriter{}).Implements(readWriterType)) // true

	// Nothing implements Closer (none of our types have Close)
	println("myReader → Closer:", reflect.TypeOf(myReader{}).Implements(closerType))             // false
	println("*myReadWriter → Closer:", reflect.TypeOf(&myReadWriter{}).Implements(closerType))   // false

	// errorValue (*errors.errorString) implements error but not Stringer
	println("errorValue → error:", reflect.TypeOf(errorValue).Implements(errorType))       // true
	println("errorValue → Stringer:", reflect.TypeOf(errorValue).Implements(stringerType)) // false

	// myErrorStringer implements both error and Stringer
	println("myErrorStringer → error:", reflect.TypeOf(myErrorStringer{}).Implements(errorType))       // true
	println("myErrorStringer → Stringer:", reflect.TypeOf(myErrorStringer{}).Implements(stringerType)) // true

	// Everything implements empty interface
	println("myReader → interface{}:", reflect.TypeOf(myReader{}).Implements(emptyItf))   // true
	println("int → interface{}:", reflect.TypeOf(0).Implements(emptyItf))                 // true

	// --- Interface implements interface (superset check, issue #3580) ---
	println("interface implements interface:")

	// ReadWriter is a superset of Reader and Writer
	println("ReadWriter → Reader:", readWriterType.Implements(readerType))         // true
	println("ReadWriter → Writer:", readWriterType.Implements(writerType))         // true
	println("Reader → ReadWriter:", readerType.Implements(readWriterType))         // false
	println("Writer → ReadWriter:", writerType.Implements(readWriterType))         // false

	// ReadCloser has Read+Close, Reader has Read
	println("ReadCloser → Reader:", readCloserType.Implements(readerType))         // true
	println("ReadCloser → Closer:", readCloserType.Implements(closerType))         // true
	println("ReadCloser → Writer:", readCloserType.Implements(writerType))         // false
	println("Reader → ReadCloser:", readerType.Implements(readCloserType))         // false

	// Self-implements
	println("Reader → Reader:", readerType.Implements(readerType))                 // true
	println("ReadWriter → ReadWriter:", readWriterType.Implements(readWriterType)) // true

	// error and Stringer are unrelated
	println("error → Stringer:", errorType.Implements(stringerType)) // false
	println("Stringer → error:", stringerType.Implements(errorType)) // false

	// Everything implements empty interface
	println("Reader → interface{}:", readerType.Implements(emptyItf))       // true
	println("ReadWriter → interface{}:", readWriterType.Implements(emptyItf)) // true

	// --- AssignableTo ---
	println("assignable to:")

	// Identical types
	println("int → int:", reflect.TypeOf(0).AssignableTo(reflect.TypeOf(0)))          // true
	println("string → string:", reflect.TypeOf("").AssignableTo(reflect.TypeOf("")))   // true

	// Different types
	println("int → string:", reflect.TypeOf(0).AssignableTo(reflect.TypeOf("")))      // false
	println("int → int64:", reflect.TypeOf(0).AssignableTo(reflect.TypeOf(int64(0)))) // false

	// Concrete assignable to interface (implements check)
	println("myReader → Reader:", reflect.TypeOf(myReader{}).AssignableTo(readerType))         // true
	println("*myWriter → Writer:", reflect.TypeOf(&myWriter{}).AssignableTo(writerType))       // true
	println("myWriter → Writer:", reflect.TypeOf(myWriter{}).AssignableTo(writerType))         // false
	println("*myReadWriter → ReadWriter:", reflect.TypeOf(&myReadWriter{}).AssignableTo(readWriterType)) // true

	// Interface assignable to interface
	println("ReadWriter → Reader:", readWriterType.AssignableTo(readerType))         // true
	println("Reader → ReadWriter:", readerType.AssignableTo(readWriterType))         // false

	// Everything assignable to empty interface
	println("int → interface{}:", reflect.TypeOf(0).AssignableTo(emptyItf))             // true
	println("Reader → interface{}:", readerType.AssignableTo(emptyItf))                 // true

	// --- Upstream set_test.go: unexported method interfaces ---
	println("unexported method interface:")
	exprType := reflect.TypeOf((*exprLike)(nil)).Elem()
	println("*notAnExpr → exprLike:", reflect.TypeOf(new(notAnExpr)).Implements(exprType))       // true
	println("notAnExpr → exprLike:", reflect.TypeOf(notAnExpr{}).Implements(exprType))            // true
	println("*notAnExpr → exprLike (AssignableTo):", reflect.TypeOf(new(notAnExpr)).AssignableTo(exprType)) // true

	// --- Upstream set_test.go: channel direction assignability ---
	println("channel direction:")
	println("chan int → <-chan int:", reflect.TypeOf(make(chan int)).AssignableTo(reflect.TypeOf(make(<-chan int))))    // true
	println("<-chan int → chan int:", reflect.TypeOf(make(<-chan int)).AssignableTo(reflect.TypeOf(make(chan int))))    // false

	// --- Upstream set_test.go: named type assignability ---
	println("named types:")
	println("*int → IntPtr:", reflect.TypeOf(new(int)).AssignableTo(reflect.TypeOf(IntPtr(nil))))     // true
	println("IntPtr → *int:", reflect.TypeOf(IntPtr(nil)).AssignableTo(reflect.TypeOf(new(int))))     // true
	println("IntPtr → IntPtr1:", reflect.TypeOf(IntPtr(nil)).AssignableTo(reflect.TypeOf(IntPtr1(nil)))) // false
	println("Ch → <-chan interface{}:", reflect.TypeOf(Ch(nil)).AssignableTo(reflect.TypeOf(make(<-chan interface{})))) // true

	// --- reflect.Value.Set with interface (issue #4277) ---
	println("value set interface:")
	type Node interface{ node() }
	type FooNode struct{ V int }
	type BarNode struct{ V int }
	// Make FooNode and BarNode implement Node with pointer receivers
	// (can't add methods to local types in function, use a different approach)
	testValueSetInterface()
	testMakeMapCompositeKey()
	testMakeMapInterfaceKey()
	testMakeMapPaddedKey()
}

type IfaceNode interface {
	ifaceNode()
}
type FooNode struct{ V int }
type BarNode struct{ V int }

func (*FooNode) ifaceNode() {}
func (*BarNode) ifaceNode() {}

type NodeContainer struct {
	Nodes []IfaceNode
}

func testValueSetInterface() {
	c := &NodeContainer{
		Nodes: []IfaceNode{&FooNode{V: 1}, &FooNode{V: 2}},
	}

	// Use reflect to replace elements
	v := reflect.ValueOf(c).Elem().FieldByName("Nodes")
	v.Index(0).Set(reflect.ValueOf(&BarNode{V: 10}))

	switch n := c.Nodes[0].(type) {
	case *BarNode:
		println("Set[0] to BarNode:", n.V) // 10
	default:
		println("FAIL: expected *BarNode")
	}
	switch n := c.Nodes[1].(type) {
	case *FooNode:
		println("Set[1] still FooNode:", n.V) // 2
	default:
		println("FAIL: expected *FooNode")
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

type compositeKey struct {
	S string
	N int32
}

// testMakeMapCompositeKey tests that reflect.MakeMap works correctly with
// composite key types (structs containing strings). This exercises the
// hash/equal dispatch path for maps created through reflection rather
// than by the compiler.
func testMakeMapCompositeKey() {
	println("\nreflect.MakeMap composite key:")
	mapType := reflect.TypeOf(map[compositeKey]int{})
	m := reflect.MakeMap(mapType)

	// Insert two keys that share the same string but differ in the int field.
	key1 := reflect.ValueOf(compositeKey{S: "hello", N: 1})
	key2 := reflect.ValueOf(compositeKey{S: "hello", N: 2})
	m.SetMapIndex(key1, reflect.ValueOf(100))
	m.SetMapIndex(key2, reflect.ValueOf(200))

	println("len:", m.Len())

	v1 := m.MapIndex(key1)
	if v1.IsValid() {
		println("key1:", v1.Int())
	} else {
		println("key1: not found")
	}
	v2 := m.MapIndex(key2)
	if v2.IsValid() {
		println("key2:", v2.Int())
	} else {
		println("key2: not found")
	}

	// Delete key1, verify key2 remains.
	m.SetMapIndex(key1, reflect.Value{})
	println("after delete, len:", m.Len())
	v2 = m.MapIndex(key2)
	if v2.IsValid() {
		println("key2 after delete:", v2.Int())
	} else {
		println("key2 after delete: not found")
	}
}

// testMakeMapInterfaceKey tests that reflect.MakeMap works correctly with
// interface{} key types, including cross-path usage (reflect insert,
// compiled lookup and vice versa).
func testMakeMapInterfaceKey() {
	println("\nreflect.MakeMap interface key:")
	mapType := reflect.TypeOf(map[interface{}]int{})
	rv := reflect.MakeMap(mapType)

	rv.SetMapIndex(reflect.ValueOf(42), reflect.ValueOf(100))
	rv.SetMapIndex(reflect.ValueOf("hello"), reflect.ValueOf(200))
	println("len:", rv.Len())

	v1 := rv.MapIndex(reflect.ValueOf(42))
	if v1.IsValid() {
		println("42:", v1.Int())
	} else {
		println("42: not found")
	}
	v2 := rv.MapIndex(reflect.ValueOf("hello"))
	if v2.IsValid() {
		println("hello:", v2.Int())
	} else {
		println("hello: not found")
	}

	// Cross-path: use from compiled code.
	m := rv.Interface().(map[interface{}]int)
	println("compiled 42:", m[42])
	println("compiled hello:", m["hello"])

	// Addressable small value as key.
	x := 99
	addrVal := reflect.ValueOf(&x).Elem()
	rv.SetMapIndex(addrVal, reflect.ValueOf(300))
	v3 := rv.MapIndex(reflect.ValueOf(99))
	if v3.IsValid() {
		println("addressable 99:", v3.Int())
	} else {
		println("addressable 99: not found")
	}
}

type paddedKey struct {
	A int8
	B int32
}

// testMakeMapPaddedKey tests that struct keys with padding work correctly
// through reflect, using addressable values with poisoned padding bytes.
func testMakeMapPaddedKey() {
	println("\nreflect.MakeMap padded key:")
	var pk1, pk2 paddedKey
	pk1.A = 1
	pk1.B = 42
	pk2.A = 1
	pk2.B = 42

	if unsafe.Offsetof(paddedKey{}.B) > 1 {
		// Poison pk2's padding byte (between A and B).
		*(*byte)(unsafe.Add(unsafe.Pointer(&pk2), 1)) = 0xFF
	}

	// Use addressable values so padding survives into reflect.
	rm := reflect.MakeMap(reflect.TypeOf(map[paddedKey]int{}))
	rm.SetMapIndex(reflect.ValueOf(&pk1).Elem(), reflect.ValueOf(100))
	v := rm.MapIndex(reflect.ValueOf(&pk2).Elem())
	if v.IsValid() {
		println("padded lookup:", v.Int())
	} else {
		println("padded lookup: not found")
	}
}
