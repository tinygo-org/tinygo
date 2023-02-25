package reflect_test

import (
	. "reflect"
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
}
