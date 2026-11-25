package reflect_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"internal/reflectlite"
	"math/rand"
	randv2 "math/rand/v2"
	. "reflect"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestTinyIndirectPointers(t *testing.T) {
	var m = map[string]int{}
	m["x"] = 1

	var a = &m

	if ValueOf(a).Elem().Len() != 1 {
		t.Errorf("bad map length via reflect")
	}

	var b struct {
		Decoded *[3]byte
	}

	v1 := New(TypeOf(b.Decoded).Elem())

	var bb [3]byte
	bb[0] = 0xaa

	v1.Elem().Set(ValueOf(bb))

	if v1.Elem().Index(0).Uint() != 0xaa {
		t.Errorf("bad indirect array index via reflect")
	}
}

func TestSetPointer(t *testing.T) {
	type pointer unsafe.Pointer
	var p pointer
	value := ValueOf(&p).Elem()
	object := new(int)
	*object = 17
	value.SetPointer(unsafe.Pointer(object))
	runtime.GC()
	if got := *(*int)(unsafe.Pointer(p)); got != 17 {
		t.Fatalf("SetPointer stored %d, want 17", got)
	}
	value.SetPointer(nil)
	if p != nil {
		t.Fatal("SetPointer(nil) did not clear the pointer")
	}
	fields := struct{ pointer unsafe.Pointer }{unsafe.Pointer(object)}
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"Unaddressable", func() { ValueOf(p).SetPointer(nil) }},
		{"WrongKind", func() { ValueOf(new(*int)).Elem().SetPointer(nil) }},
		{"ZeroValue", func() { Value{}.SetPointer(nil) }},
		{"UnexportedField", func() { ValueOf(&fields).Elem().Field(0).SetPointer(nil) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("SetPointer did not panic")
				}
			}()
			tc.fn()
		})
	}
	if fields.pointer != unsafe.Pointer(object) {
		t.Fatal("SetPointer changed an unexported field")
	}
}

func TestTinyInterfaceData(t *testing.T) {
	var empty any
	var nonempty = fmt.Errorf("interface data")
	fields := struct{ hidden any }{new(int)}
	for _, test := range []struct {
		value  Value
		header unsafe.Pointer
	}{
		{ValueOf(&empty).Elem(), unsafe.Pointer(&empty)},
		{ValueOf(&nonempty).Elem(), unsafe.Pointer(&nonempty)},
		{ValueOf(&fields).Elem().Field(0), unsafe.Pointer(&fields.hidden)},
	} {
		if got, want := test.value.InterfaceData(), *(*[2]uintptr)(test.header); got != want {
			t.Errorf("InterfaceData = %v, want %v", got, want)
		}
	}
	shouldPanic("", func() { ValueOf(1).InterfaceData() })
	shouldPanic("", func() { Value{}.InterfaceData() })
}

func TestTinyInvalidValueString(t *testing.T) {
	if got := (Value{}).String(); got != "<invalid Value>" {
		t.Errorf("Value{}.String() = %q, want %q", got, "<invalid Value>")
	}
}

func TestNewAt(t *testing.T) {
	value := 42
	v := NewAt(TypeOf(value), unsafe.Pointer(&value))
	if got, want := v.Type(), TypeOf((*int)(nil)); got != want {
		t.Fatalf("NewAt type = %v, want %v", got, want)
	}
	if got := v.Interface().(*int); got != &value {
		t.Fatalf("NewAt pointer = %p, want %p", got, &value)
	}
	v.Elem().SetInt(23)
	if value != 23 {
		t.Fatalf("NewAt value = %d, want 23", value)
	}
	if !v.Elem().CanAddr() || !v.Elem().CanSet() {
		t.Fatal("NewAt element is not addressable and settable")
	}
	if got := v.Elem().Addr().Interface(); got != v.Interface() {
		t.Fatalf("address of NewAt element = %v, want %v", got, v.Interface())
	}

	type composite struct {
		Value   int
		Pointer *int
	}
	pointed := 1
	compositeValue := composite{Value: 42, Pointer: &pointed}
	compositePointer := NewAt(TypeOf(compositeValue), unsafe.Pointer(&compositeValue))
	compositePointer.Elem().Field(0).SetInt(23)
	if compositeValue.Value != 23 {
		t.Fatalf("NewAt struct field = %d, want 23", compositeValue.Value)
	}
	other := 2
	compositePointer.Elem().Field(1).Set(ValueOf(&other))
	if compositeValue.Pointer != &other {
		t.Fatalf("NewAt pointer field = %p, want %p", compositeValue.Pointer, &other)
	}

	nilValue := NewAt(TypeOf(value), nil)
	if !nilValue.IsNil() {
		t.Fatal("NewAt with nil pointer is not nil")
	}
	if nilValue.Elem().IsValid() {
		t.Fatal("Elem of NewAt with nil pointer is valid")
	}
}

func TestTinyMap(t *testing.T) {

	m := make(map[string]int)

	mtyp := TypeOf(m)

	if got, want := mtyp.Key().Kind().String(), "string"; got != want {
		t.Errorf("m.Type().Key().String()=%q, want %q", got, want)
	}

	if got, want := mtyp.Elem().Kind().String(), "int"; got != want {
		t.Errorf("m.Elem().String()=%q, want %q", got, want)
	}

	m["foo"] = 2

	mref := ValueOf(m)
	two := mref.MapIndex(ValueOf("foo"))

	if got, want := two.Interface().(int), 2; got != want {
		t.Errorf("MapIndex(`foo`)=%v, want %v", got, want)
	}

	m["bar"] = 3
	m["baz"] = 4
	m["qux"] = 5

	it := mref.MapRange()

	var gotKeys []string
	for it.Next() {
		k := it.Key()
		v := it.Value()

		kstr := k.Interface().(string)
		vint := v.Interface().(int)

		gotKeys = append(gotKeys, kstr)

		if m[kstr] != vint {
			t.Errorf("m[%v]=%v, want %v", kstr, vint, m[kstr])
		}
	}
	var wantKeys []string
	for k := range m {
		wantKeys = append(wantKeys, k)
	}
	sort.Strings(gotKeys)
	sort.Strings(wantKeys)

	if !equal(gotKeys, wantKeys) {
		t.Errorf("MapRange return unexpected keys: got %v, want %v", gotKeys, wantKeys)
	}

	refMapKeys := mref.MapKeys()
	gotKeys = gotKeys[:0]
	for _, v := range refMapKeys {
		gotKeys = append(gotKeys, v.Interface().(string))
	}

	sort.Strings(gotKeys)
	if !equal(gotKeys, wantKeys) {
		t.Errorf("MapKeys return unexpected keys: got %v, want %v", gotKeys, wantKeys)
	}

	mref.SetMapIndex(ValueOf("bar"), Value{})
	if _, ok := m["bar"]; ok {
		t.Errorf("SetMapIndex failed to delete `bar`")
	}

	mref.SetMapIndex(ValueOf("baz"), ValueOf(6))
	if got, want := m["baz"], 6; got != want {
		t.Errorf("SetMapIndex(bar, 6) got %v, want %v", got, want)
	}

	m2ref := MakeMap(mref.Type())
	m2ref.SetMapIndex(ValueOf("foo"), ValueOf(2))

	m2 := m2ref.Interface().(map[string]int)

	if m2["foo"] != 2 {
		t.Errorf("MakeMap failed to create map")
	}

	type stringint struct {
		s string
		i int
	}

	simap := make(map[stringint]int)

	refsimap := MakeMap(TypeOf(simap))

	refsimap.SetMapIndex(ValueOf(stringint{"hello", 4}), ValueOf(6))

	six := refsimap.MapIndex(ValueOf(stringint{"hello", 4}))

	if six.Interface().(int) != 6 {
		t.Errorf("m[hello, 4]=%v, want 6", six)
	}

	keys := refsimap.MapKeys()
	if len(keys) != 1 {
		t.Errorf("refsimap: MapKeys()=%v, want 1", len(keys))
	}
	if keys[0].Type() != TypeOf(stringint{}) {
		t.Errorf("keys[0] has wrong type: %v", keys[0].Type().String())
	}

	sikey := keys[0].Interface().(stringint)

	if sikey != (stringint{"hello", 4}) {
		t.Errorf("sikey has unexpected value: %#v", sikey)
	}

	// make sure we can look up interface keys with reflection
	type unmarshalerText struct {
		A, B string
	}

	ut := make(map[unmarshalerText]bool)

	refut := ValueOf(ut)

	// put in a key with the compiler
	ut[unmarshalerText{"x", "y"}] = true

	// make sure we can get it out
	v2 := refut.MapIndex(ValueOf(unmarshalerText{"x", "y"}))
	if !v2.IsValid() || !v2.Bool() {
		t.Errorf("Failed to look up map struct key with reflection")
	}

	// put in a key with reflection
	refut.SetMapIndex(ValueOf(unmarshalerText{"y", "z"}), ValueOf(true))

	// make sure we can get it out with the compiler
	if !ut[unmarshalerText{"y", "z"}] {
		t.Errorf("Failed to look up up reflect-set map key with compiler")
	}

	utKeys := refut.MapKeys()

	// make sure keys extracted via reflection have the correct type
	if _, ok := utKeys[0].Interface().(unmarshalerText); !ok {
		t.Errorf("Map keys via MapKeys() have wrong type: %v", utKeys[0].Type().String())
	}

	// and via iteration

	utIter := refut.MapRange()
	utIter.Next()
	utIterKey := utIter.Key()

	if _, ok := utIterKey.Interface().(unmarshalerText); !ok {
		t.Errorf("Map keys via MapIter() have wrong type: %v", utIterKey.Type().String())
	}

	{
		m := map[any]any{1: 2}
		rv := ValueOf(m)
		iter := rv.MapRange()

		iter.Next()
		k := iter.Key()
		if k.Kind().String() != "interface" {
			t.Errorf("map[any]any MapRange has wrong key type: %v", k.Kind().String())
		}

		keys := rv.MapKeys()
		if k := keys[0]; k.Kind().String() != "interface" {
			t.Errorf("map[any]any MapRange has wrong key type: %v", k.Kind().String())
		}

	}
}

// For an interface type, it returns the number of exported and unexported methods.

type counter interface {
	count() int
}

type count struct {
	i int
}

func (c *count) count() int {
	return c.i
}

func TestMapInterfaceKeys(t *testing.T) {
	m := make(map[interface{}]int)
	for i := 0; i < 20; i++ {
		c := &count{i}
		m[c] = i
	}

	for k, v := range m {
		if got := m[k]; got != v {
			t.Error("lookup failure got", got, "want", v)
		}
	}

	refm := ValueOf(m)

	keys := refm.MapKeys()
	if len(keys) != 20 {
		t.Error("failed to look up 20 keys")
	}
	for _, k := range keys {
		c := k.Interface().(*count)
		e := refm.MapIndex(k)
		if e.Interface().(int) != c.i {
			t.Error("reflect lookup failure:", c.i)
		}
	}

	it := refm.MapRange()
	var totalKeys int
	for it.Next() {
		totalKeys++
		k := it.Key()
		v := it.Value().Interface().(int)

		c := k.Interface().(*count)
		if v != c.i || refm.MapIndex(k).Interface().(int) != c.i {
			t.Error("reflect iter lookup failure:", c.i)
		}
	}

	if totalKeys != 20 {
		t.Errorf("failed to get 20 keys")
	}
}

func TestMapInterfaceElem(t *testing.T) {
	m := make(map[string]interface{})
	refm := ValueOf(m)

	four := ValueOf(4)
	hello := ValueOf("hello")

	refm.SetMapIndex(hello, four)

	if v := refm.MapIndex(hello).Interface().(int); v != 4 {
		t.Errorf("failed to get value assigned to interface via MapIndex")
	}

	if v := m["hello"].(int); v != 4 {
		t.Errorf("failed to get value assigned to interface via direct lookup")
	}
}

func TestTinySlice(t *testing.T) {
	s := []int{0, 10, 20}
	refs := ValueOf(s)

	for i := 3; i < 10; i++ {
		refs = Append(refs, ValueOf(i*10))
	}

	s = refs.Interface().([]int)

	for i := 0; i < 10; i++ {
		if s[i] != i*10 {
			t.Errorf("s[%d]=%d, want %d", i, s[i], i*10)
		}
	}

	s28 := s[2:8]
	s28ref := refs.Slice(2, 8)

	if len(s28) != s28ref.Len() || cap(s28) != s28ref.Cap() {
		t.Errorf("Slice: len(s28)=%d s28ref.Len()=%d cap(s28)=%d s28ref.Cap()=%d\n", len(s28), s28ref.Len(), cap(s28), s28ref.Cap())
	}

	for i, got := range s28 {
		want := int(s28ref.Index(i).Int())
		if got != want {
			t.Errorf("s28[%d]=%d, want %d", i, got, want)
		}
	}

	s268 := s[2:6:8]
	s268ref := refs.Slice3(2, 6, 8)

	if len(s268) != s268ref.Len() || cap(s268) != s268ref.Cap() {
		t.Errorf("Slice3: len(s268)=%d s268ref.Len()=%d cap(s268)=%d s268ref.Cap()=%d\n", len(s268), s268ref.Len(), cap(s268), s268ref.Cap())
	}

	for i, got := range s268 {
		want := int(s268ref.Index(i).Int())
		if got != want {
			t.Errorf("s268[%d]=%d, want %d", i, got, want)
		}
	}

	// should be equivalent to s28 now, except for the capacity which doesn't change

	s268ref = ValueOf(&s268).Elem()
	s268ref.SetLen(6)
	if len(s28) != s268ref.Len() || cap(s268) != s268ref.Cap() {
		t.Errorf("SetLen: len(s268)=%d s268ref.Len()=%d cap(s268)=%d s268ref.Cap()=%d\n", len(s28), s268ref.Len(), cap(s268), s268ref.Cap())
	}

	for i, got := range s28 {
		want := int(s268ref.Index(i).Int())
		if got != want {
			t.Errorf("s28[%d]=%d, want %d", i, got, want)
		}
	}

	refs = MakeSlice(TypeOf(s), 5, 10)
	s = refs.Interface().([]int)

	if len(s) != refs.Len() || cap(s) != refs.Cap() {
		t.Errorf("len(s)=%v refs.Len()=%v cap(s)=%v refs.Cap()=%v", len(s), refs.Len(), cap(s), refs.Cap())
	}
}

func TestTinySlice3(t *testing.T) {
	slice := make([]int, 3, 6)
	for i := range cap(slice) {
		slice[:cap(slice)][i] = i
	}

	value := ValueOf(slice)
	sliced := value.Slice3(4, 5, 6).Interface().([]int)
	if len(sliced) != 1 || cap(sliced) != 2 || sliced[0] != 4 {
		t.Fatalf("Slice3(4, 5, 6) = %v with cap %d, want [4] with cap 2", sliced, cap(sliced))
	}

	first := value.Slice3(2, 4, 6)
	empty := first.Slice3(4, 4, 4)
	if got, want := empty.UnsafePointer(), first.UnsafePointer(); got != want {
		t.Fatalf("empty Slice3 pointer = %p, want %p", got, want)
	}

	array := [4]int{1, 2, 3, 4}
	arraySlice := ValueOf(&array).Elem().Slice3(1, 2, 3).Interface().([]int)
	if len(arraySlice) != 1 || cap(arraySlice) != 2 || arraySlice[0] != 2 {
		t.Fatalf("array Slice3(1, 2, 3) = %v with cap %d, want [2] with cap 2", arraySlice, cap(arraySlice))
	}

	checkPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()
		fn()
	}
	checkPanic("Slice3 with reversed indexes", func() {
		value.Slice3(2, 1, 3)
	})
	checkPanic("Slice3 with negative index", func() {
		value.Slice3(-1, 1, 2)
	})
	checkPanic("Slice3 beyond capacity", func() {
		value.Slice3(0, 3, 7)
	})
	checkPanic("Slice3 on unaddressable array", func() {
		ValueOf(array).Slice3(0, 1, 2)
	})
	checkPanic("Slice3 on unsupported kind", func() {
		ValueOf("abc").Slice3(0, 1, 2)
	})
}

func TestTinySetCap(t *testing.T) {
	slice := make([]int, 3, 6)
	settable := ValueOf(&slice).Elem()
	settable.SetCap(5)
	if len(slice) != 3 || cap(slice) != 5 {
		t.Fatalf("after SetCap(5), len, cap = %d, %d, want 3, 5", len(slice), cap(slice))
	}
	settable.SetLen(5)
	settable.SetCap(5)

	checkPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()
		fn()
	}
	checkPanic("SetCap below length", func() {
		settable.SetCap(4)
	})
	checkPanic("SetCap above capacity", func() {
		settable.SetCap(6)
	})
	checkPanic("SetCap with negative capacity", func() {
		settable.SetCap(-1)
	})
	checkPanic("SetCap on non-slice", func() {
		ValueOf(new(string)).Elem().SetCap(0)
	})
	checkPanic("SetCap on unaddressable slice", func() {
		ValueOf(slice).SetCap(5)
	})
	checkPanic("SetCap on read-only slice", func() {
		value := ValueOf(&struct {
			slice []int
		}{slice}).Elem().Field(0)
		value.SetCap(5)
	})
}

func TestTinySetLen(t *testing.T) {
	slice := []int{1, 2, 3}
	value := ValueOf(&slice).Elem()
	value.SetLen(1)
	if len(slice) != 1 || cap(slice) != 3 || slice[0] != 1 {
		t.Fatalf("SetLen changed the wrong slice data: %v, cap %d", slice, cap(slice))
	}
	value.SetLen(3)
	shouldPanic("", func() { value.SetLen(-1) })
	shouldPanic("", func() { value.SetLen(4) })
	shouldPanic("", func() { ValueOf(slice).SetLen(1) })
	shouldPanic("", func() { ValueOf(new(int)).Elem().SetLen(1) })

	fields := struct{ slice []int }{slice}
	shouldPanic("", func() { ValueOf(&fields).Elem().Field(0).SetLen(1) })
	if len(fields.slice) != 3 {
		t.Fatal("SetLen changed an unexported field")
	}
}

func TestSliceAtSet(t *testing.T) {
	array := [3]int{1, 2, 3}
	value := SliceAt(TypeOf(0), unsafe.Pointer(&array[0]), len(array))
	if got, want := value.Type(), TypeOf([]int{}); got != want {
		t.Fatalf("SliceAt type = %v, want %v", got, want)
	}
	if got, want := value.Len(), len(array); got != want {
		t.Fatalf("SliceAt len = %d, want %d", got, want)
	}
	if got, want := value.Cap(), len(array); got != want {
		t.Fatalf("SliceAt cap = %d, want %d", got, want)
	}
	value.Index(1).SetInt(4)
	if array[1] != 4 {
		t.Fatalf("SliceAt update = %d, want 4", array[1])
	}
}

func TestTinyBytes(t *testing.T) {
	s := []byte("abcde")
	refs := ValueOf(s)

	s2 := refs.Bytes()

	if !equal(s, s2) {
		t.Errorf("Failed to get Bytes(): %v != %v", s, s2)
	}

	Copy(refs, ValueOf("12345"))

	if string(s) != "12345" {
		t.Errorf("Copy()=%q, want `12345`", string(s))
	}

	// test small arrays that fit in a pointer
	a := [3]byte{10, 20, 30}
	v := ValueOf(&a)
	vslice := v.Elem().Bytes()
	if len(vslice) != 3 || cap(vslice) != 3 {
		t.Errorf("len(vslice)=%v, cap(vslice)=%v", len(vslice), cap(vslice))
	}

	for i, got := range vslice {
		if want := (byte(i) + 1) * 10; got != want {
			t.Errorf("vslice[%d]=%d, want %d", i, got, want)
		}
	}
}

func TestTinyNamedTypes(t *testing.T) {
	type namedString string

	named := namedString("foo")
	if got, want := TypeOf(named).Name(), "namedString"; got != want {
		t.Errorf("TypeOf.Name()=%v, want %v", got, want)
	}

	if got, want := TypeOf(named).String(), "reflect_test.namedString"; got != want {
		t.Errorf("TypeOf.String()=%v, want %v", got, want)
	}

	errorType := TypeOf((*error)(nil)).Elem()
	if s := errorType.String(); s != "error" {
		t.Errorf("error type = %v, want error", s)
	}

	m := make(map[[4]uint16]string)

	if got, want := TypeOf(m).String(), "map[[4]uint16]string"; got != want {
		t.Errorf("Type.String()=%v, want %v", got, want)
	}

	s := struct {
		a int8
		b int8
		c int8
		d int8
		e int8
		f int32
	}{}

	if got, want := TypeOf(s).String(), "struct { a int8; b int8; c int8; d int8; e int8; f int32 }"; got != want {
		t.Errorf("Type.String()=%v, want %v", got, want)
	}

	if got, want := ValueOf(m).String(), "<map[[4]uint16]string Value>"; got != want {
		t.Errorf("Value.String()=%v, want %v", got, want)
	}

	if got, want := TypeOf(base64.Encoding{}).String(), "base64.Encoding"; got != want {
		t.Errorf("Type.String(base64.Encoding{})=%v, want %v", got, want)
	}

	type Repository struct {
		RoleName *string `json:"role_name,omitempty"`
	}

	var repo *Repository
	v := ValueOf(&repo).Elem()
	n := New(v.Type().Elem())
	v.Set(n)
}

func TestTinyStruct(t *testing.T) {
	type barStruct struct {
		QuxString string
		BazInt    int
	}

	type foobar struct {
		Foo string `foo:"struct tag"`
		Bar barStruct
	}

	var fb foobar
	fb.Bar.QuxString = "qux"

	reffb := TypeOf(fb)

	q := reffb.FieldByIndex([]int{1, 0})
	if want := "QuxString"; q.Name != want {
		t.Errorf("FieldByIndex=%v, want %v", q.Name, want)
	}

	var ok bool
	q, ok = reffb.FieldByName("Foo")
	if q.Name != "Foo" || !ok {
		t.Errorf("FieldByName(Foo)=%v,%v, want Foo, true", q.Name, ok)
	}

	if got, want := q.Tag, `foo:"struct tag"`; string(got) != want {
		t.Errorf("StrucTag for Foo=%v, want %v", got, want)
	}

	q, ok = reffb.FieldByNameFunc(func(s string) bool { return strings.ToLower(s) == "bar" })
	if q.Name != "Bar" || !ok {
		t.Errorf("FieldByNameFunc(bar)=%v,%v, want Bar, true", q.Name, ok)
	}

	q, ok = reffb.FieldByName("Snorble")
	if q.Name != "" || ok {
		t.Errorf("FieldByName(Snorble)=%v,%v, want ``, false", q.Name, ok)
	}

	q, ok = reffb.FieldByNameFunc(func(s string) bool { return strings.ToLower(s) == "snorble" })
	if q.Name != "" || ok {
		t.Errorf("FieldByName(snorble)=%v,%v, want ``, false", q.Name, ok)
	}
}

func TestTinyZero(t *testing.T) {
	for _, value := range []any{(*int)(nil), [16]byte{}, [64]byte{}} {
		zero := Zero(TypeOf(value))
		if IsRO(zero) {
			t.Errorf("Zero(%v) is read-only", zero.Type())
		}
	}

	s := "hello, world"
	sptr := &s
	v := ValueOf(&sptr).Elem()
	v.Set(Zero(v.Type()))

	sptr = v.Interface().(*string)

	if sptr != nil {
		t.Errorf("failed to set a nil string pointer")
	}

	sl := []int{1, 2, 3}
	v = ValueOf(&sl).Elem()
	v.Set(Zero(v.Type()))
	sl = v.Interface().([]int)

	if sl != nil {
		t.Errorf("failed to set a nil slice")
	}
}

func addrDecode(body interface{}) {
	vbody := ValueOf(body)
	ptr := vbody.Elem()
	pptr := ptr.Addr()
	addrSetInt(pptr.Interface())
}

func addrSetInt(intf interface{}) {
	ptr := intf.(*uint64)
	*ptr = 112358
}

func TestTinyAddr(t *testing.T) {
	var n uint64
	addrDecode(&n)
	if n != 112358 {
		t.Errorf("Failed to set t=112358, got %v", n)
	}

	v := ValueOf(&n)
	if got, want := v.Elem().Addr().CanAddr(), false; got != want {
		t.Errorf("Elem.Addr.CanAddr=%v, want %v", got, want)
	}
}

func TestTinyNilType(t *testing.T) {
	var a any
	typ := TypeOf(a)
	if typ != nil {
		t.Errorf("Type of any{nil} is not nil")
	}
}

func TestTinySetBytes(t *testing.T) {
	var b []byte
	refb := ValueOf(&b).Elem()
	s := []byte("hello")
	refb.SetBytes(s)
	s[0] = 'b'

	refbSlice := refb.Interface().([]byte)

	if len(refbSlice) != len(s) || b[0] != s[0] || refbSlice[0] != s[0] {
		t.Errorf("SetBytes(): reflection slice mismatch")
	}
}

type methodStruct struct {
	i int
}

func (m methodStruct) ValueMethod1() int {
	return m.i
}

func (m methodStruct) valueMethod2() int {
	return m.i
}

func (m *methodStruct) PointerMethod1() int {
	return m.i
}

func (m *methodStruct) PointerMethod2() int {
	return m.i
}

func (m *methodStruct) pointerMethod3() int {
	return m.i
}

func TestTinyNumMethods(t *testing.T) {
	refptrt := TypeOf(&methodStruct{})
	if got, want := refptrt.NumMethod(), 1+2; got != want {
		t.Errorf("Pointer Methods=%v, want %v", got, want)
	}

	reft := refptrt.Elem()
	if got, want := reft.NumMethod(), 1; got != want {
		t.Errorf("Value Methods=%v, want %v", got, want)
	}

	for _, typ := range []Type{
		TypeOf((**methodStruct)(nil)),
		TypeOf((***methodStruct)(nil)),
		TypeOf((****methodStruct)(nil)),
	} {
		if got := typ.NumMethod(); got != 0 {
			t.Errorf("%v Methods=%v, want 0", typ, got)
		}
		if method, ok := typ.MethodByName("ValueMethod1"); ok {
			t.Errorf("%v MethodByName=%v, want no method", typ, method)
		}
	}
}

func TestAssignableTo(t *testing.T) {
	var a any
	refa := ValueOf(&a).Elem()
	refa.Set(ValueOf(4))
	if got, want := refa.Interface().(int), 4; got != want {
		t.Errorf("AssignableTo / Set failed, got %v, want %v", got, want)
	}

	b := []byte{0x01, 0x02}
	refb := ValueOf(&b).Elem()
	refb.Set(ValueOf([]byte{0x02, 0x03}))
	if got, want := refb.Interface().([]byte), []byte{0x02, 0x03}; !bytes.Equal(got, want) {
		t.Errorf("AssignableTo / Set failed, got %v, want %v", got, want)
	}

	type bstr []byte

	c := bstr{0x01, 0x02}
	refc := ValueOf(&c).Elem()
	refc.Set(ValueOf([]byte{0x02, 0x03}))
	if got, want := refb.Interface().([]byte), []byte{0x02, 0x03}; !bytes.Equal(got, want) {
		t.Errorf("AssignableTo / Set failed, got %v, want %v", got, want)
	}

	d := []byte{0x01, 0x02}
	refd := ValueOf(&d).Elem()
	refd.Set(ValueOf(bstr{0x02, 0x03}))
	if got, want := refb.Interface().([]byte), []byte{0x02, 0x03}; !bytes.Equal(got, want) {
		t.Errorf("AssignableTo / Set failed, got %v, want %v", got, want)
	}
}

func TestConvertBasic(t *testing.T) {
	v := ValueOf(int64(3))
	c := v.Convert(TypeOf(byte(0)))

	if c.Type().Kind() != Uint8 || c.Uint() != 3 {
		t.Errorf("Convert(uint64 -> byte failed: kind=%v, value=%d", c.Type().Kind().String(), c.Uint())
	}

	v = ValueOf("hello")
	c = v.Convert(TypeOf([]byte("")))

	if c.Type().Kind() != Slice || c.Type().Elem().Kind() != Uint8 && c.Len() != 5 && string(c.Bytes()) != "hello" {
		t.Errorf("Convert(string -> []byte")
	}

	type namedString string

	c = v.Convert(TypeOf(namedString("")))
	if c.Type().Kind() != String || c.Type().Name() != "namedString" {
		t.Errorf("Convert(string -> namedString")
	}
}

func TestConvertSliceToArrayOrArrayPointer(t *testing.T) {
	s := make([]byte, 2, 4)
	// a0 := [0]byte(s)
	// a1 := [1]byte(s[1:])     // a1[0] == s[1]
	// a2 := [2]byte(s)         // a2[0] == s[0]
	// a4 := [4]byte(s)         // panics: len([4]byte) > len(s)

	v := ValueOf(s).Convert(TypeFor[[0]byte]())
	if v.Kind() != Array || v.Type().Len() != 0 {
		t.Error("Convert([]byte -> [0]byte)")
	}
	nilArray := ValueOf([]int(nil)).Convert(TypeFor[[0]int]())
	if nilArray.CanAddr() || nilArray.Interface().([0]int) != [0]int{} {
		t.Error("Convert(nil []int -> [0]int)")
	}
	v = ValueOf(s[1:]).Convert(TypeFor[[1]byte]())
	if v.Kind() != Array || v.Type().Len() != 1 {
		t.Error("Convert([]byte -> [1]byte)")
	}
	v = ValueOf(s).Convert(TypeFor[[2]byte]())
	if v.Kind() != Array || v.Type().Len() != 2 {
		t.Error("Convert([]byte -> [2]byte)")
	}
	if ValueOf(s).CanConvert(TypeFor[[4]byte]()) {
		t.Error("Converting a slice with len smaller than array to array should fail")
	}

	// s0 := (*[0]byte)(s)      // s0 != nil
	// s1 := (*[1]byte)(s[1:])  // &s1[0] == &s[1]
	// s2 := (*[2]byte)(s)      // &s2[0] == &s[0]
	// s4 := (*[4]byte)(s)      // panics: len([4]byte) > len(s)
	v = ValueOf(s).Convert(TypeFor[*[0]byte]())
	if v.Kind() != Pointer || v.Elem().Kind() != Array || v.Elem().Type().Len() != 0 {
		t.Error("Convert([]byte -> *[0]byte)")
	}
	v = ValueOf(s[1:]).Convert(TypeFor[*[1]byte]())
	if v.Kind() != Pointer || v.Elem().Kind() != Array || v.Elem().Type().Len() != 1 {
		t.Error("Convert([]byte -> *[1]byte)")
	}
	v = ValueOf(s).Convert(TypeFor[*[2]byte]())
	if v.Kind() != Pointer || v.Elem().Kind() != Array || v.Elem().Type().Len() != 2 {
		t.Error("Convert([]byte -> *[2]byte)")
	}
	if ValueOf(s).CanConvert(TypeFor[*[4]byte]()) {
		t.Error("Converting a slice with len smaller than array to array pointer should fail")
	}

	// Test converting slices with backing arrays <= and >64bits
	slice64 := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	array64 := ValueOf(slice64).Convert(TypeFor[[8]byte]()).Interface().([8]byte)
	if !bytes.Equal(slice64, array64[:]) {
		t.Errorf("converted array %x does not match backing array of slice %x", array64, slice64)
	}

	slice72 := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
	array72 := ValueOf(slice72).Convert(TypeFor[[9]byte]()).Interface().([9]byte)
	if !bytes.Equal(slice72, array72[:]) {
		t.Errorf("converted array %x does not match backing array of slice %x", array72, slice72)
	}
}

func TestConvertToEmptyInterface(t *testing.T) {
	anyType := TypeFor[interface{}]()

	v := ValueOf(false).Convert(anyType)
	if v.Kind() != Interface || v.NumMethod() > 0 {
		t.Error("Convert(bool -> interface{})")
	}
	_ = v.Interface().(interface{}).(bool)

	v = ValueOf(int64(3)).Convert(anyType)
	if v.Kind() != Interface || v.NumMethod() > 0 {
		t.Error("Convert(int64 -> interface{})")
	}
	_ = v.Interface().(interface{}).(int64)

	v = ValueOf(struct{}{}).Convert(anyType)
	if v.Kind() != Interface || v.NumMethod() > 0 {
		t.Error("Convert(struct -> interface{})")
	}
	_ = v.Interface().(interface{}).(struct{})

	v = ValueOf([]struct{}{}).Convert(anyType)
	if v.Kind() != Interface || v.NumMethod() > 0 {
		t.Error("Convert(slice -> interface{})")
	}
	_ = v.Interface().(interface{}).([]struct{})

	v = ValueOf(map[string]string{"A": "B"}).Convert(anyType)
	if v.Kind() != Interface || v.NumMethod() > 0 {
		t.Error("Convert(map -> interface{})")
	}
	_ = v.Interface().(interface{}).(map[string]string)
}

type conversionInt int

type conversionBox[T any] struct {
	Value int
}

type conversionStructA struct {
	Value int `json:"a"`
}

type conversionStructB struct {
	Value int `json:"b"`
}

type conversionChan chan int

type conversionInterface interface {
	Value() int
}

type conversionExtendedInterface interface {
	conversionInterface
	Extra()
}

type conversionMethods int

func (v conversionMethods) Value() int {
	return int(v)
}

func (conversionMethods) Extra() {
}

type conversionLarge struct {
	N       int
	Padding int
}

func (v conversionLarge) Value() int {
	return v.N
}

func TestConvertInterfaces(t *testing.T) {
	for _, test := range []struct {
		from, to Type
		want     bool
	}{
		{TypeFor[any](), TypeFor[error](), false},
		{TypeFor[error](), TypeFor[any](), true},
		{TypeFor[conversionInterface](), TypeFor[error](), false},
		{TypeFor[conversionInterface](), TypeFor[conversionExtendedInterface](), false},
		{TypeFor[conversionExtendedInterface](), TypeFor[conversionInterface](), true},
	} {
		t.Run(test.from.String()+"->"+test.to.String(), func(t *testing.T) {
			value := New(test.from).Elem()
			if got := test.from.ConvertibleTo(test.to); got != test.want {
				t.Fatalf("ConvertibleTo = %v, want %v", got, test.want)
			}
			if got := value.CanConvert(test.to); got != test.want {
				t.Fatalf("CanConvert = %v, want %v", got, test.want)
			}
			if !test.want {
				shouldPanic("", func() { value.Convert(test.to) })
			} else if converted := value.Convert(test.to); !converted.IsNil() || converted.CanAddr() {
				t.Fatal("converting a nil interface did not produce a non-addressable nil interface")
			}
		})
	}
}

func TestConvertSnapshots(t *testing.T) {
	channel := make(chan int)
	for _, test := range []struct {
		name        string
		value, want any
	}{
		{"struct", conversionStructA{23}, conversionStructB{23}},
		{"large-struct", struct {
			Value [2]int `json:"a"`
		}{[2]int{23, 24}}, struct {
			Value [2]int `json:"b"`
		}{[2]int{23, 24}}},
		{"channel", conversionChan(channel), (<-chan int)(channel)},
		{"same-struct", conversionStructA{23}, conversionStructA{23}},
		{"same-large-struct", conversionLarge{23, 24}, conversionLarge{23, 24}},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := New(TypeOf(test.value)).Elem()
			value.Set(ValueOf(test.value))
			converted := value.Convert(TypeOf(test.want))
			if converted.CanAddr() || converted.CanSet() {
				t.Error("converted value is addressable or settable")
			}
			value.Set(Zero(value.Type()))
			if got := converted.Interface(); !DeepEqual(got, test.want) {
				t.Errorf("converted value = %v, want %v", got, test.want)
			}
		})
	}
}

func TestConvertInterfaceSnapshots(t *testing.T) {
	for _, original := range []conversionInterface{conversionMethods(9), conversionLarge{11, 12}} {
		t.Run(TypeOf(original).String(), func(t *testing.T) {
			value := New(TypeOf(original)).Elem()
			value.Set(ValueOf(original))
			converted := value.Convert(TypeFor[conversionInterface]())
			value.Set(Zero(value.Type()))
			if got, want := converted.Interface().(conversionInterface).Value(), original.Value(); got != want {
				t.Errorf("converted interface value = %d, want %d", got, want)
			}
		})
	}
}

func TestConvertUnexportedFields(t *testing.T) {
	fields := struct {
		small conversionMethods
		large conversionLarge
		iface conversionInterface
	}{conversionMethods(9), conversionLarge{11, 12}, conversionMethods(13)}
	value := ValueOf(&fields).Elem()
	for i := 0; i < value.NumField(); i++ {
		t.Run(value.Type().Field(i).Name, func(t *testing.T) {
			converted := value.Field(i).Convert(TypeFor[conversionInterface]())
			if converted.CanInterface() || converted.CanAddr() || converted.CanSet() {
				t.Fatal("conversion exposed an unexported field")
			}
			shouldPanic("", func() { converted.Interface() })
		})
	}
}

func TestConvertChannelElementTypes(t *testing.T) {
	type elementA = struct {
		Value int `json:"a"`
	}
	type elementB = struct {
		Value int `json:"b"`
	}
	for _, test := range []struct {
		name     string
		from, to Type
		want     bool
	}{
		{"identical", TypeFor[chan elementA](), TypeFor[<-chan elementA](), true},
		{"receive-different-tags", TypeFor[chan elementA](), TypeFor[<-chan elementB](), false},
		{"send-different-tags", TypeFor[chan elementA](), TypeFor[chan<- elementB](), false},
		{"bidirectional-different-tags", TypeFor[chan elementA](), TypeFor[chan elementB](), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := Zero(test.from)
			if test.from.ConvertibleTo(test.to) != test.want || value.CanConvert(test.to) != test.want {
				t.Fatal("channel conversion did not respect element identity")
			}
			if test.want {
				if converted := value.Convert(test.to); converted.Type() != test.to || !converted.IsNil() {
					t.Fatal("channel conversion did not preserve the nil channel")
				}
			} else {
				shouldPanic("", func() { value.Convert(test.to) })
			}
		})
	}
}

func TestConvertGenericTypes(t *testing.T) {
	for _, test := range []struct {
		typ  Type
		name string
	}{
		{TypeFor[conversionBox[rand.Source]](), "conversionBox[math/rand.Source]"},
		{TypeFor[conversionBox[randv2.Source]](), "conversionBox[math/rand/v2.Source]"},
		{TypeFor[conversionBox[[]rand.Source]](), "conversionBox[[]math/rand.Source]"},
		{TypeFor[conversionBox[byte]](), "conversionBox[uint8]"},
		{TypeFor[conversionBox[rune]](), "conversionBox[int32]"},
		{TypeFor[conversionBox[func(value int)]](), "conversionBox[func(int)]"},
		{TypeFor[conversionBox[struct{ Value int }]](), "conversionBox[struct { Value int }]"},
	} {
		if got := test.typ.Name(); got != test.name {
			t.Errorf("generic type name = %q, want %q", got, test.name)
		}
	}
	for _, test := range []struct {
		name     string
		from, to Type
	}{
		{"slice", TypeFor[[]conversionBox[rand.Source]](), TypeFor[[]conversionBox[randv2.Source]]()},
		{"array", TypeFor[[1]conversionBox[rand.Source]](), TypeFor[[1]conversionBox[randv2.Source]]()},
		{"map", TypeFor[map[int]conversionBox[rand.Source]](), TypeFor[map[int]conversionBox[randv2.Source]]()},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := Zero(test.from)
			if test.from.ConvertibleTo(test.to) || value.CanConvert(test.to) {
				t.Error("conversion accepted distinct generic element types")
			}
			shouldPanic("", func() { value.Convert(test.to) })
		})
	}
	value := ValueOf(conversionBox[rand.Source]{Value: 42})
	target := TypeFor[conversionBox[randv2.Source]]()
	if !value.Type().ConvertibleTo(target) || !value.CanConvert(target) {
		t.Fatal("generic structs with identical underlying types should be convertible")
	}
	if got := value.Convert(target).Interface().(conversionBox[randv2.Source]).Value; got != 42 {
		t.Fatalf("converted generic struct value = %d, want 42", got)
	}
}

func TestConvertGenericLocalTypes(t *testing.T) {
	var value Value
	var target Type
	{
		type token int
		value = ValueOf([]conversionBox[token]{{Value: 42}})
	}
	{
		type token int
		target = TypeFor[[]conversionBox[token]]()
	}
	if value.Type() == target {
		t.Fatal("distinct local type arguments share a descriptor")
	}
	if value.Type().ConvertibleTo(target) || value.CanConvert(target) {
		t.Error("conversion accepted distinct local type arguments")
	}
	shouldPanic("", func() { value.Convert(target) })
}

func TestConvertCompositeTypes(t *testing.T) {
	n := 42
	pointer := ValueOf(&n).Convert(TypeOf((*conversionInt)(nil))).Interface().(*conversionInt)
	if got := int(*pointer); got != n {
		t.Fatalf("converted pointer value = %d, want %d", got, n)
	}

	structValue := conversionStructA{Value: 23}
	convertedStruct := ValueOf(structValue).Convert(TypeOf(conversionStructB{})).Interface().(conversionStructB)
	if convertedStruct.Value != structValue.Value {
		t.Fatalf("converted struct value = %d, want %d", convertedStruct.Value, structValue.Value)
	}
	structPointer := ValueOf(&structValue)
	pointerType := TypeOf((*conversionStructB)(nil))
	if !structPointer.Type().ConvertibleTo(pointerType) || !structPointer.CanConvert(pointerType) {
		t.Fatal("pointers to structs with different tags should be convertible")
	}
	convertedPointer := structPointer.Convert(pointerType).Interface().(*conversionStructB)
	if convertedPointer.Value != structValue.Value {
		t.Fatalf("converted struct pointer value = %d, want %d", convertedPointer.Value, structValue.Value)
	}

	channel := make(conversionChan)
	recv := ValueOf(channel).Convert(TypeOf((<-chan int)(nil))).Interface().(<-chan int)
	send := ValueOf(channel).Convert(TypeOf((chan<- int)(nil))).Interface().(chan<- int)
	go func() {
		send <- 7
	}()
	if got := <-recv; got != 7 {
		t.Fatalf("converted channel received %d, want 7", got)
	}

	concrete := ValueOf(conversionMethods(9)).Convert(TypeFor[conversionInterface]())
	if got := concrete.Interface().(conversionInterface).Value(); got != 9 {
		t.Fatalf("converted concrete interface value = %d, want 9", got)
	}

	var extended conversionExtendedInterface = conversionMethods(11)
	convertedInterface := ValueOf(&extended).Elem().Convert(TypeFor[conversionInterface]())
	if got := convertedInterface.Interface().(conversionInterface).Value(); got != 11 {
		t.Fatalf("converted interface value = %d, want 11", got)
	}
}

func TestCanConvertNoAlloc(t *testing.T) {
	slice := make([]byte, 32)
	for _, test := range []struct {
		value  Value
		target Type
	}{
		{ValueOf("conversion"), TypeFor[[]byte]()},
		{ValueOf("conversion"), TypeFor[[]rune]()},
		{ValueOf(slice), TypeFor[[32]byte]()},
		{ValueOf(slice), TypeFor[*[32]byte]()},
	} {
		if allocations := testing.AllocsPerRun(100, func() {
			if !test.value.CanConvert(test.target) {
				t.Fatal("CanConvert rejected a valid conversion")
			}
		}); allocations != 0 {
			t.Errorf("CanConvert(%v to %v) allocated %v times", test.value.Type(), test.target, allocations)
		}
	}
	shouldPanic("", func() { Value{}.CanConvert(TypeFor[int]()) })
	shouldPanic("", func() { ValueOf(1).CanConvert(nil) })
}

func TestClearSlice(t *testing.T) {
	type stringSlice []string
	for _, test := range []struct {
		slice  any
		expect any
	}{
		{
			slice:  []bool{true, false, true},
			expect: []bool{false, false, false},
		},
		{
			slice:  []byte{0x00, 0x01, 0x02, 0x03},
			expect: []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			slice:  [][]int{[]int{2, 1}, []int{3}, []int{}},
			expect: [][]int{nil, nil, nil},
		},
		{
			slice:  []stringSlice{stringSlice{"hello", "world"}, stringSlice{}, stringSlice{"goodbye"}},
			expect: []stringSlice{nil, nil, nil},
		},
	} {
		v := ValueOf(test.slice)
		expectLen, expectCap := v.Len(), v.Cap()

		v.Clear()
		if len := v.Len(); len != expectLen {
			t.Errorf("Clear(slice) altered len, got %d, expected %d", len, expectLen)
		}
		if cap := v.Cap(); cap != expectCap {
			t.Errorf("Clear(slice) altered cap, got %d, expected %d", cap, expectCap)
		}
		if !DeepEqual(test.slice, test.expect) {
			t.Errorf("Clear(slice) got %v, expected %v", test.slice, test.expect)
		}
	}
}

func TestClearMap(t *testing.T) {
	for _, test := range []struct {
		m      any
		expect any
	}{
		{
			m:      map[int]bool{1: true, 2: false, 3: true},
			expect: map[int]bool{},
		},
		{
			m:      map[string][]byte{"hello": []byte("world")},
			expect: map[string][]byte{},
		},
	} {
		v := ValueOf(test.m)

		v.Clear()
		if len := v.Len(); len != 0 {
			t.Errorf("Clear(map) should set len to 0, got %d", len)
		}
		if !DeepEqual(test.m, test.expect) {
			t.Errorf("Clear(map) got %v, expected %v", test.m, test.expect)
		}
	}
}

func TestCopyArrayToSlice(t *testing.T) {
	// Test copying array <=64 bits and >64bits
	// See issue #4554
	a1 := [1]int64{1}
	s1 := make([]int64, 1)
	a2 := [2]int64{1, 2}
	s2 := make([]int64, 2)
	a8 := [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	s8 := make([]byte, 8)
	a9 := [9]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
	s9 := make([]byte, 9)

	Copy(ValueOf(s1), ValueOf(a1))
	if !slices.Equal(s1, a1[:]) {
		t.Errorf("copied slice %x does not match input array %x", s1, a1[:])
	}
	Copy(ValueOf(s2), ValueOf(a2))
	if !slices.Equal(s2, a2[:]) {
		t.Errorf("copied slice %x does not match input array %x", s2, a2[:])
	}
	Copy(ValueOf(s8), ValueOf(a8))
	if !bytes.Equal(s8, a8[:]) {
		t.Errorf("copied slice %x does not match input array %x", s8, a8[:])
	}
	Copy(ValueOf(s9), ValueOf(a9))
	if !bytes.Equal(s9, a9[:]) {
		t.Errorf("copied slice %x does not match input array %x", s9, a9[:])
	}
}

func TestIssue4040(t *testing.T) {
	var value interface{} = uint16(0)

	// get the pointer to the interface value
	inPtr := ValueOf(&value)

	// dereference to get the actual value (an interface)
	inElem := inPtr.Elem()

	// create a new value of the same concrete type
	uint16Type := TypeOf(uint16(0))
	newUint16Value := New(uint16Type).Elem()
	newUint16Value.Set(ValueOf(uint16(13)))

	// set the new value to the interface
	inElem.Set(newUint16Value)

	if value.(uint16) != 13 {
		t.Errorf("Failed to set interface value from uint16")
	}
}

func equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i, aa := range a {
		if b[i] != aa {
			return false
		}
	}
	return true
}

func TestTypeAssert(t *testing.T) {
	testTypeAssert(t, int(123456789), int(123456789), true)
	testTypeAssert(t, int(-123456789), int(-123456789), true)
	testTypeAssert(t, int32(123456789), int32(123456789), true)
	testTypeAssert(t, int8(-123), int8(-123), true)
	testTypeAssert(t, [2]int{1234, -5678}, [2]int{1234, -5678}, true)
	testTypeAssert(t, "test value", "test value", true)
	testTypeAssert(t, any("test value"), any("test value"), true)

	v := 123456789
	testTypeAssert(t, &v, &v, true)

	testTypeAssert(t, int(123), uint(0), false)

	testTypeAssert[any](t, 1, 1, true)
	testTypeAssert[any](t, any(nil), nil, false)
	testTypeAssert[any](t, error(nil), nil, false)
	testTypeAssert[any](t, (*int)(nil), (*int)(nil), true)
	testTypeAssert[fmt.Stringer](t, 1, nil, false)

	vv := testTypeWithMethod{"test"}
	testTypeAssert[any](t, vv, vv, true)
	testTypeAssert[any](t, &vv, &vv, true)
	testTypeAssert[fmt.Stringer](t, vv, vv, true)
	testTypeAssert[fmt.Stringer](t, &vv, &vv, true)
	testTypeAssert[interface{ A() }](t, vv, nil, false)
	testTypeAssert[interface{ A() }](t, &vv, nil, false)
	testTypeAssert(t, any(vv), any(vv), true)
	testTypeAssert(t, fmt.Stringer(vv), fmt.Stringer(vv), true)

	testTypeAssert(t, fmt.Stringer(vv), any(vv), true)
	testTypeAssert(t, any(vv), fmt.Stringer(vv), true)
	testTypeAssert(t, fmt.Stringer(vv), interface{ M() }(vv), true)
	testTypeAssert(t, interface{ M() }(vv), fmt.Stringer(vv), true)

	testTypeAssert(t, any(int(1)), int(1), true)
	testTypeAssert(t, any(int(1)), byte(0), false)
	testTypeAssert(t, fmt.Stringer(vv), vv, true)
}

func testTypeAssert[T comparable, V any](t *testing.T, val V, wantVal T, wantOk bool) {
	t.Helper()

	v, ok := TypeAssert[T](ValueOf(&val).Elem())
	if v != wantVal || ok != wantOk {
		t.Errorf("TypeAssert[%v](%#v) = (%#v, %v); want = (%#v, %v)", TypeFor[T](), val, v, ok, wantVal, wantOk)
	}

	// Additionally make sure that TypeAssert[T](v) behaves in the same way as v.Interface().(T).
	v2, ok2 := ValueOf(&val).Elem().Interface().(T)
	if v != v2 || ok != ok2 {
		t.Errorf("reflect.ValueOf(%#v).Interface().(%v) = (%#v, %v); want = (%#v, %v)", val, TypeFor[T](), v2, ok2, v, ok)
	}
}

func TestTypeAssertStruct(t *testing.T) {
	type taStruct struct {
		i int
		b bool
	}

	var a any

	// struct
	a = taStruct{3, true}
	if s, ok := a.(taStruct); ok {
		if s.i != 3 || s.b != true {
			t.Errorf("a.(S) failed: got s.i=%v, s.b=%v\n", s.i, s.b)
		}
	} else {
		t.Errorf("a.(S) failed: got ok=false")
	}

	if s, ok := TypeAssert[taStruct](ValueOf(a)); ok {
		if s.i != 3 || s.b != true {
			t.Errorf("TypeAssert[S] failed: got s.i=%v, s.b=%v\n", s.i, s.b)
		}
	} else {
		t.Errorf("TypeAssert[S] failed: got ok=false")
	}

	// struct ptr
	a = &taStruct{3, true}
	if s, ok := a.(*taStruct); ok {
		if s.i != 3 || s.b != true {
			t.Errorf("a.(*S) failed: got s.i=%v, s.b=%v\n", s.i, s.b)
		}
	} else {
		t.Errorf("a.(*S) failed: got ok=false")
	}

	if s, ok := TypeAssert[*taStruct](ValueOf(a)); ok {
		if s.i != 3 || s.b != true {
			t.Errorf("TypeAssert[*S] failed: got s.i=%v, s.b=%v\n", s.i, s.b)
		}
	} else {
		t.Errorf("TypeAssert[*S] failed: got ok=false")
	}

}

type testTypeWithMethod struct{ val string }

func (v testTypeWithMethod) String() string { return v.val }
func (v testTypeWithMethod) M()             {}

func TestTypeAssertPanic(t *testing.T) {
	t.Run("zero val", func(t *testing.T) {
		defer func() { recover() }()
		TypeAssert[int](Value{})
		t.Fatalf("TypeAssert did not panic")
	})
	t.Run("read only", func(t *testing.T) {
		defer func() { recover() }()
		TypeAssert[int](ValueOf(&testTypeWithMethod{}).FieldByName("val"))
		t.Fatalf("TypeAssert did not panic")
	})
}

type tinyMakeChanElement struct {
	ptr  *int
	text string
}

type namedReflectChan chan int

var tinyMakeChanChurn []*int

//go:noinline
func fillTinyMakeChan(ch chan tinyMakeChanElement) {
	value := new(int)
	*value = 42
	ch <- tinyMakeChanElement{ptr: value, text: "hello"}
}

func TestChannelSmallValues(t *testing.T) {
	testChannelSmallValue(t, true)
	testChannelSmallValue(t, int8(-7))
	testChannelSmallValue(t, int16(-0x1234))
	testChannelSmallValue(t, uint16(0xabcd))
	testChannelSmallValue(t, float32(1.25))
	testChannelSmallValue(t, [3]byte{0x12, 0x34, 0x56})
}

func testChannelSmallValue[T comparable](t *testing.T, want T) {
	t.Helper()
	t.Run(TypeFor[T]().String(), func(t *testing.T) {
		ch := make(chan T, 1)
		channel := ValueOf(ch)
		for _, value := range []Value{ValueOf(want), ValueOf(&want).Elem()} {
			channel.Send(value)
			if got := <-ch; got != want {
				t.Fatalf("Send = %v, want %v", got, want)
			}
			if !channel.TrySend(value) {
				t.Fatal("TrySend failed")
			}
			if got := <-ch; got != want {
				t.Fatalf("TrySend = %v, want %v", got, want)
			}
			Select([]SelectCase{{Dir: SelectSend, Chan: channel, Send: value}})
			if got := <-ch; got != want {
				t.Fatalf("Select send = %v, want %v", got, want)
			}
			interfaceChannel := make(chan any, 1)
			for _, method := range []string{"Send", "TrySend", "Select"} {
				channel := ValueOf(interfaceChannel)
				switch method {
				case "Send":
					channel.Send(value)
				case "TrySend":
					if !channel.TrySend(value) {
						t.Fatal("TrySend to interface channel failed")
					}
				case "Select":
					Select([]SelectCase{{Dir: SelectSend, Chan: channel, Send: value}})
				}
				if got := (<-interfaceChannel).(T); got != want {
					t.Fatalf("%s to interface channel = %v, want %v", method, got, want)
				}
			}
		}
		for _, method := range []string{"Recv", "TryRecv", "Select"} {
			ch <- want
			var value Value
			var ok bool
			switch method {
			case "Recv":
				value, ok = channel.Recv()
			case "TryRecv":
				value, ok = channel.TryRecv()
			case "Select":
				_, value, ok = Select([]SelectCase{{Dir: SelectRecv, Chan: channel}})
			}
			if !ok || value.Interface().(T) != want {
				t.Fatalf("%s = %v, %t, want %v, true", method, value, ok, want)
			}
		}
	})
}

func TestChannelSendAssignabilityPanic(t *testing.T) {
	channel := ValueOf(make(chan int, 1))
	value := ValueOf("invalid")
	const message = "reflect.Value.Send: value of type string is not assignable to type int"
	shouldPanic(message, func() { channel.Send(value) })
	shouldPanic(message, func() { channel.TrySend(value) })
	shouldPanic("reflect.Select: value of type string is not assignable to type int", func() {
		Select([]SelectCase{{Dir: SelectSend, Chan: channel, Send: value}})
	})
}

func TestChannelSelectNilBlocks(t *testing.T) {
	for _, cases := range [][]SelectCase{
		nil,
		{{Dir: SelectRecv}},
		{{Dir: SelectRecv, Chan: ValueOf((chan int)(nil))}},
		{
			{Dir: SelectRecv, Chan: ValueOf((chan int)(nil))},
			{Dir: SelectSend, Chan: ValueOf((chan int)(nil)), Send: ValueOf(1)},
		},
	} {
		started := make(chan struct{})
		done := make(chan struct{})
		go func() {
			close(started)
			Select(cases)
			close(done)
		}()
		<-started
		select {
		case <-done:
			t.Fatalf("Select(%v) returned", cases)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestChannelOperations(t *testing.T) {
	channel := make(chan int, 1)
	value := ValueOf(channel)

	value.Send(ValueOf(1))
	if got := <-channel; got != 1 {
		t.Fatalf("Send value = %d, want 1", got)
	}

	channel <- 2
	got, ok := value.Recv()
	if !ok || got.Int() != 2 {
		t.Fatalf("Recv = %v, %t, want 2, true", got, ok)
	}

	if got, ok := value.TryRecv(); got.IsValid() || ok {
		t.Fatalf("TryRecv on empty channel = %v, %t, want invalid, false", got, ok)
	}
	channel <- 3
	if got, ok := value.TryRecv(); !ok || got.Int() != 3 {
		t.Fatalf("TryRecv = %v, %t, want 3, true", got, ok)
	}

	channel <- 4
	if value.TrySend(ValueOf(5)) {
		t.Fatal("TrySend on full channel succeeded")
	}
	<-channel
	if !value.TrySend(ValueOf(6)) {
		t.Fatal("TrySend on empty channel failed")
	}
	if got := <-channel; got != 6 {
		t.Fatalf("TrySend value = %d, want 6", got)
	}

	channel <- 7
	value.Close()
	if got, ok := value.Recv(); !ok || got.Int() != 7 {
		t.Fatalf("Recv after Close = %v, %t, want 7, true", got, ok)
	}
	if got, ok := value.Recv(); ok || got.Int() != 0 {
		t.Fatalf("Recv from closed channel = %v, %t, want 0, false", got, ok)
	}
	if got, ok := value.TryRecv(); ok || got.Int() != 0 {
		t.Fatalf("TryRecv from closed channel = %v, %t, want 0, false", got, ok)
	}

	interfaceChannel := make(chan any, 1)
	ValueOf(interfaceChannel).Send(ValueOf("value"))
	if got := <-interfaceChannel; got != "value" {
		t.Fatalf("interface channel value = %v, want value", got)
	}

	named := make(namedReflectChan, 1)
	namedValue := ValueOf(named)
	namedValue.Send(ValueOf(8))
	if got, ok := namedValue.Recv(); !ok || got.Int() != 8 {
		t.Fatalf("named channel Recv = %v, %t, want 8, true", got, ok)
	}
}

func TestChannelBlockedOperations(t *testing.T) {
	a, b := 7, 8
	want := [2]*int{&a, &b}
	for _, channel := range []Value{
		ValueOf(make(chan [2]*int)),
		ValueOf(make(chan any)),
	} {
		t.Run(channel.Type().String(), func(t *testing.T) {
			started := make(chan struct{})
			done := make(chan struct{})
			go func() {
				close(started)
				channel.Send(ValueOf(want))
				close(done)
			}()
			<-started
			runtime.Gosched()
			runtime.GC()
			got, ok := channel.Recv()
			if !ok || got.Interface().([2]*int) != want {
				t.Fatalf("blocked Send delivered %v, %t, want %v, true", got, ok, want)
			}
			<-done

			started = make(chan struct{})
			received := make(chan Value)
			go func() {
				close(started)
				value, ok := channel.Recv()
				if !ok {
					panic("blocked Recv failed")
				}
				received <- value
			}()
			<-started
			runtime.Gosched()
			runtime.GC()
			channel.Send(ValueOf(want))
			got = <-received
			if got.Interface().([2]*int) != want || got.CanAddr() || got.CanSet() {
				t.Fatalf("blocked Recv returned %v with incorrect value or flags", got)
			}
		})
	}
}

func TestChannelReceiveFlags(t *testing.T) {
	channels := []Value{
		ValueOf(make(chan int, 1)),
		ValueOf(make(chan [2]*int, 1)),
		ValueOf(make(chan any, 1)),
		ValueOf(make(chan struct{}, 1)),
	}
	for _, channel := range channels {
		elem := channel.Type().Elem()
		t.Run(elem.String(), func(t *testing.T) {
			for _, closed := range []bool{false, true} {
				if closed {
					channel.Close()
				}
				for _, method := range []string{"Recv", "TryRecv", "Select"} {
					if !closed {
						channel.Send(Zero(elem))
					}
					var got Value
					var ok bool
					switch method {
					case "Recv":
						got, ok = channel.Recv()
					case "TryRecv":
						got, ok = channel.TryRecv()
					case "Select":
						_, got, ok = Select([]SelectCase{{Dir: SelectRecv, Chan: channel}})
					}
					if !got.IsValid() || got.Type() != elem || ok == closed {
						t.Fatalf("%s on closed=%t returned %v, %t", method, closed, got, ok)
					}
					if got.CanAddr() || got.CanSet() {
						t.Errorf("%s on closed=%t returned an addressable or settable value", method, closed)
					}
				}
			}
		})
	}
}

func TestChannelSelect(t *testing.T) {
	t.Run("blocked interface sends", func(t *testing.T) {
		first := make(chan any)
		second := make(chan any)
		started := make(chan struct{})
		chosen := make(chan int, 1)
		a, b := 1, 2
		go func() {
			close(started)
			index, _, _ := Select([]SelectCase{
				{Dir: SelectSend, Chan: ValueOf(first), Send: ValueOf(&a)},
				{Dir: SelectSend, Chan: ValueOf(second), Send: ValueOf(&b)},
			})
			chosen <- index
		}()
		<-started
		runtime.Gosched()
		runtime.GC()
		if got := <-second; got != &b {
			t.Fatalf("Select sent %v, want %p", got, &b)
		}
		if index := <-chosen; index != 1 {
			t.Fatalf("Select chose %d, want 1", index)
		}
	})

	t.Run("ready receive before send", func(t *testing.T) {
		for _, blocking := range []bool{false, true} {
			received := false
			for range 64 {
				in := make(chan int, 1)
				out := make(chan int, 1)
				in <- 42
				cases := []SelectCase{
					{Dir: SelectRecv, Chan: ValueOf(in)},
					{Dir: SelectSend, Chan: ValueOf(out), Send: ValueOf(42)},
				}
				if !blocking {
					cases = append(cases, SelectCase{Dir: SelectDefault})
				}
				chosen, _, _ := Select(cases)
				received = received || chosen == 0
			}
			if !received {
				t.Fatalf("Select blocking=%t always preferred the ready send", blocking)
			}
		}
	})

	t.Run("mixed case ordering", func(t *testing.T) {
		for _, blocking := range []bool{false, true} {
			for ready := 0; ready < 4; ready++ {
				var channels [4]chan int
				channels[ready] = make(chan int, 1)
				if ready%2 == 0 {
					channels[ready] <- 42
				}
				cases := []SelectCase{
					{Dir: SelectRecv, Chan: ValueOf(channels[0])},
					{Dir: SelectSend, Chan: ValueOf(channels[1]), Send: ValueOf(42)},
					{Dir: SelectRecv, Chan: ValueOf(channels[2])},
					{Dir: SelectSend, Chan: ValueOf(channels[3]), Send: ValueOf(42)},
				}
				if !blocking {
					cases = append(cases, SelectCase{Dir: SelectDefault})
				}
				chosen, recv, ok := Select(cases)
				if chosen != ready {
					t.Fatalf("Select blocking=%t chose %d, want %d", blocking, chosen, ready)
				}
				if ready%2 == 0 {
					if !ok || recv.Int() != 42 {
						t.Fatalf("Select receive = %v, %t, want 42, true", recv, ok)
					}
				} else {
					if recv.IsValid() || ok {
						t.Fatalf("Select send = %v, %t, want invalid, false", recv, ok)
					}
					if got := <-channels[ready]; got != 42 {
						t.Fatalf("Select sent %d, want 42", got)
					}
				}
			}
		}
	})

	first := make(chan *int)
	second := make(chan [2]*int, 1)
	a, b := 1, 2
	second <- [2]*int{&a, &b}

	chosen, recv, ok := Select([]SelectCase{
		{Dir: SelectRecv, Chan: ValueOf(first)},
		{Dir: SelectRecv, Chan: ValueOf(second)},
	})
	if chosen != 1 || !ok {
		t.Fatalf("Select receive chose %d, %t, want 1, true", chosen, ok)
	}
	got := recv.Interface().([2]*int)
	if got[0] != &a || got[1] != &b {
		t.Fatalf("Select receive = %v, want [%p %p]", got, &a, &b)
	}

	sendChannel := make(chan string, 1)
	chosen, recv, ok = Select([]SelectCase{
		{Dir: SelectSend, Chan: ValueOf(sendChannel), Send: ValueOf("sent")},
	})
	if chosen != 0 || recv.IsValid() || ok {
		t.Fatalf("Select send = %d, %v, %t, want 0, invalid, false", chosen, recv, ok)
	}
	if got := <-sendChannel; got != "sent" {
		t.Fatalf("Select sent %q, want sent", got)
	}

	chosen, recv, ok = Select([]SelectCase{
		{Dir: SelectRecv, Chan: ValueOf(first)},
		{Dir: SelectDefault},
	})
	if chosen != 1 || recv.IsValid() || ok {
		t.Fatalf("Select default = %d, %v, %t, want 1, invalid, false", chosen, recv, ok)
	}

	blocking := make(chan int)
	go func() {
		blocking <- 9
	}()
	chosen, recv, ok = Select([]SelectCase{
		{Dir: SelectRecv, Chan: ValueOf(blocking)},
	})
	if chosen != 0 || !ok || recv.Int() != 9 {
		t.Fatalf("blocking Select = %d, %v, %t, want 0, 9, true", chosen, recv, ok)
	}

	blockingSend := make(chan int)
	done := make(chan int)
	go func() {
		done <- <-blockingSend
	}()
	chosen, recv, ok = Select([]SelectCase{
		{Dir: SelectSend, Chan: ValueOf(blockingSend), Send: ValueOf(10)},
	})
	if chosen != 0 || recv.IsValid() || ok {
		t.Fatalf("blocking send Select = %d, %v, %t, want 0, invalid, false", chosen, recv, ok)
	}
	if got := <-done; got != 10 {
		t.Fatalf("blocking Select sent %d, want 10", got)
	}

	interfaceChannel := make(chan any, 1)
	chosen, _, _ = Select([]SelectCase{
		{Dir: SelectSend, Chan: ValueOf(interfaceChannel), Send: ValueOf("interface")},
	})
	if chosen != 0 {
		t.Fatalf("interface send Select chose %d, want 0", chosen)
	}
	if got := <-interfaceChannel; got != "interface" {
		t.Fatalf("interface send Select value = %v, want interface", got)
	}
}

func TestTinyMakeChan(t *testing.T) {
	t.Run("buffered", func(t *testing.T) {
		v := MakeChan(TypeOf(make(chan int)), 2)
		if got, want := v.Kind(), Chan; got != want {
			t.Errorf("Kind()=%v, want %v", got, want)
		}
		if got, want := v.Cap(), 2; got != want {
			t.Errorf("Cap()=%v, want %v", got, want)
		}
		if got, want := v.Len(), 0; got != want {
			t.Errorf("Len()=%v, want %v", got, want)
		}

		ch, ok := v.Interface().(chan int)
		if !ok {
			t.Fatalf("Interface() is %T, want chan int", v.Interface())
		}
		ch <- 1
		ch <- 2
		if got, want := v.Len(), 2; got != want {
			t.Errorf("Len()=%v after two sends, want %v", got, want)
		}
		if got, want := <-ch, 1; got != want {
			t.Errorf("<-ch=%v, want %v", got, want)
		}
		if got, want := <-ch, 2; got != want {
			t.Errorf("<-ch=%v, want %v", got, want)
		}
	})

	t.Run("buffered pointers survive GC", func(t *testing.T) {
		v := MakeChan(TypeOf(make(chan tinyMakeChanElement)), 1)
		ch := v.Interface().(chan tinyMakeChanElement)
		fillTinyMakeChan(ch)
		runtime.GC()
		tinyMakeChanChurn = make([]*int, 128)
		for i := range tinyMakeChanChurn {
			tinyMakeChanChurn[i] = new(int)
		}

		got := <-ch
		if *got.ptr != 42 || got.text != "hello" {
			t.Errorf("<-ch=%v, want {42 hello}", got)
		}
	})

	t.Run("unbuffered", func(t *testing.T) {
		v := MakeChan(TypeOf(make(chan string)), 0)
		if got, want := v.Cap(), 0; got != want {
			t.Errorf("Cap()=%v, want %v", got, want)
		}
		ch, ok := v.Interface().(chan string)
		if !ok {
			t.Fatalf("Interface() is %T, want chan string", v.Interface())
		}
		go func() { ch <- "hello" }()
		if got, want := <-ch, "hello"; got != want {
			t.Errorf("<-ch=%q, want %q", got, want)
		}
	})

	t.Run("not a channel", func(t *testing.T) {
		defer func() { recover() }()
		MakeChan(TypeOf(0), 0)
		t.Fatalf("MakeChan did not panic on a non-channel type")
	})

	t.Run("negative buffer", func(t *testing.T) {
		defer func() { recover() }()
		MakeChan(TypeOf(make(chan int)), -1)
		t.Fatalf("MakeChan did not panic on a negative buffer size")
	})

	t.Run("directional", func(t *testing.T) {
		defer func() { recover() }()
		var recvOnly <-chan int
		MakeChan(TypeOf(recvOnly), 0)
		t.Fatalf("MakeChan did not panic on a unidirectional channel type")
	})
}

func TestTinyDeepEqualInterfaceVisits(t *testing.T) {
	type pair struct {
		First, Second any
	}
	for _, tc := range []struct {
		name string
		x, y any
		want bool
	}{
		{"equal slice", []any{1, 2}, []any{1, 2}, true},
		{"unequal slice", []any{1, 2}, []any{1, 3}, false},
		{"unequal array", &[2]any{1, 2}, &[2]any{1, 3}, false},
		{"equal fields", &pair{1, 2}, &pair{1, 2}, true},
		{"unequal fields", &pair{1, 2}, &pair{1, 3}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DeepEqual(tc.x, tc.y); got != tc.want {
				t.Errorf("DeepEqual(%v, %v) = %v, want %v", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

func TestTinyDeepEqualSliceVisits(t *testing.T) {
	type pair struct {
		First, Second []int
	}
	left, right := []int{1, 2, 3}, []int{1, 2, 3}
	x := &pair{left[:1], left[:2]}
	y := &pair{right[:1], right[:3]}
	if DeepEqual(x, y) {
		t.Error("DeepEqual conflated slices with different lengths")
	}
	y.Second = right[:2]
	if !DeepEqual(x, y) {
		t.Error("DeepEqual rejected equal slices")
	}
}

func TestTinyDeepEqualLargeCycle(t *testing.T) {
	type node struct {
		Value int
		Next  *node
	}
	makeCycle := func(backEdge int) []*node {
		nodes := make([]*node, 32)
		for i := range nodes {
			nodes[i] = &node{Value: i}
		}
		for i := range nodes {
			nodes[i].Next = nodes[(i+1)%len(nodes)]
		}
		nodes[len(nodes)-1].Next = nodes[backEdge]
		return nodes
	}
	for _, backEdge := range []int{0, 16} {
		left, right := makeCycle(backEdge), makeCycle(backEdge)
		if !DeepEqual(left[0], right[0]) {
			t.Fatalf("equal large cycles compare unequal with back edge %d", backEdge)
		}
		right[24].Value++
		if DeepEqual(left[0], right[0]) {
			t.Errorf("unequal large cycles compare equal with back edge %d", backEdge)
		}
		if DeepEqual(right[0], left[0]) {
			t.Errorf("reverse comparison of unequal large cycles compares equal with back edge %d", backEdge)
		}
	}
}

// Functions needed by all_test.go

func IsRO(v Value) bool {
	return reflectlite.IsRO(v.Value)
}

func MakeRO(v Value) Value {
	return Value{reflectlite.MakeRO(v.Value, true)}
}
