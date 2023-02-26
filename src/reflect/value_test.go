package reflect_test

import (
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

	refs = MakeSlice(TypeOf(s), 5, 10)
	s = refs.Interface().([]int)

	if len(s) != refs.Len() || cap(s) != refs.Cap() {
		t.Errorf("len(s)=%v refs.Len()=%v cap(s)=%v refs.Cap()=%v", len(s), refs.Len(), cap(s), refs.Cap())
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
