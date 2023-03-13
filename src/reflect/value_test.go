package reflect_test

import (
	"encoding/base64"
	. "reflect"
	"sort"
	"testing"
)

func TestIndirectPointers(t *testing.T) {
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

func TestMap(t *testing.T) {

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
}

func TestSlice(t *testing.T) {
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

func TestBytes(t *testing.T) {
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

func TestNamedTypes(t *testing.T) {
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

}

func TestStruct(t *testing.T) {
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
		t.Errorf("FieldByName(Foo)=%v,%v, want Foo, true")
	}

	if got, want := q.Tag, `foo:"struct tag"`; string(got) != want {
		t.Errorf("StrucTag for Foo=%v, want %v", got, want)
	}

	q, ok = reffb.FieldByName("Snorble")
	if q.Name != "" || ok {
		t.Errorf("FieldByName(Snorble)=%v,%v, want ``, false")
	}
}

func TestZero(t *testing.T) {
	s := "hello, world"
	var sptr *string = &s
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

func TestAddr(t *testing.T) {
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

func TestNilType(t *testing.T) {
	var a any = nil
	typ := TypeOf(a)
	if typ != nil {
		t.Errorf("Type of any{nil} is not nil")
	}
}

func TestSetBytes(t *testing.T) {
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

func (m methodStruct) valueMethod1() int {
	return m.i
}

func (m methodStruct) valueMethod2() int {
	return m.i
}

func (m *methodStruct) pointerMethod1() int {
	return m.i
}

func (m *methodStruct) pointerMethod2() int {
	return m.i
}

func (m *methodStruct) pointerMethod3() int {
	return m.i
}

func TestNumMethods(t *testing.T) {
	refptrt := TypeOf(&methodStruct{})
	if got, want := refptrt.NumMethod(), 2+3; got != want {
		t.Errorf("Pointer Methods=%v, want %v", got, want)
	}

	reft := refptrt.Elem()
	if got, want := reft.NumMethod(), 2; got != want {
		t.Errorf("Value Methods=%v, want %v", got, want)
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
