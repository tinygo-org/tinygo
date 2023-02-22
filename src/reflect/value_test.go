package reflect_test

import (
	"reflect"
	"sort"
	"testing"
)

func TestIndirectPointers(t *testing.T) {
	var m = map[string]int{}
	m["x"] = 1

	var a = &m

	if reflect.ValueOf(a).Elem().Len() != 1 {
		t.Errorf("bad map length via reflect")
	}

	var b struct {
		Decoded *[3]byte
	}

	v1 := reflect.New(reflect.TypeOf(b.Decoded).Elem())

	var bb [3]byte
	bb[0] = 0xaa

	v1.Elem().Set(reflect.ValueOf(bb))

	if v1.Elem().Index(0).Uint() != 0xaa {
		t.Errorf("bad indirect array index via reflect")
	}
}

func TestReflectMap(t *testing.T) {

	m := make(map[string]int)
	m["two"] = 2
	v := reflect.ValueOf(m)
	k := reflect.ValueOf("two")

	if got, want := v.MapIndex(k).Interface().(int), m["two"]; got != want {
		t.Errorf(`MapIndex("two")=%v, want %v`, got, want)

	}

	m["two"] = 3
	v.SetMapIndex(k, reflect.ValueOf(3))
	if got, want := v.MapIndex(k).Interface().(int), m["two"]; got != want {
		t.Errorf(`MapIndex("two")=%v, want %v`, got, want)
	}

	mapT := reflect.TypeOf(m)

	if kstr, estr := mapT.Key().String(), mapT.Elem().String(); kstr != "string" || estr != "int" {
		t.Errorf("mapT.Key()=%v, mapT.Elem()=%v", kstr, estr)
	}

	m["four"] = 4
	m["five"] = 5
	m["seven"] = 7

	// delete m["two"]
	v.SetMapIndex(reflect.ValueOf("two"), reflect.Value{})
	if got, want := v.MapIndex(k), (reflect.Value{}); got != want {
		t.Errorf(`MapIndex("two")=%v, want %v`, got, want)
	}

	keys := v.MapKeys()
	var kstrs []string
	for _, k := range keys {
		kstrs = append(kstrs, k.Interface().(string))
	}

	sort.Strings(kstrs)
	want := []string{"five", "four", "seven"}
	if !equal(kstrs, want) {
		t.Errorf("keys=%v, want=%v", kstrs, want)
	}

	// integer keys / string values
	intm := map[int]string{
		10: "ten",
		20: "twenty",
	}

	intv := reflect.ValueOf(intm)
	if got, want := intv.MapIndex(reflect.ValueOf(10)).Interface().(string), intm[10]; got != want {
		t.Errorf(`MapIndex(10)=%v, want %v`, got, want)
	}

	intv.SetMapIndex(reflect.ValueOf(10), reflect.Value{})
	if got, want := intv.MapIndex(reflect.ValueOf(10)), (reflect.Value{}); got != want {
		t.Errorf(`MapIndex(10)=%v, want %v`, got, want)
	}

	intv.SetMapIndex(reflect.ValueOf(30), reflect.ValueOf("thirty"))

	keys = intv.MapKeys()
	var kints []int
	for _, k := range keys {
		kints = append(kints, k.Interface().(int))
	}

	sort.Ints(kints)
	wantints := []int{20, 30}
	if !equal(kints, wantints) {
		t.Errorf("keys=%v, want=%v", kints, wantints)
	}

	intm2 := make(map[int]string)
	it := intv.MapRange()
	for it.Next() {
		intm2[it.Key().Interface().(int)] = it.Value().Interface().(string)
	}

	if !reflect.DeepEqual(intm2, intm) {
		t.Errorf("intm2 != intm")
	}

	maptype := reflect.TypeOf(m)
	refmap := reflect.MakeMap(maptype)
	refmap.SetMapIndex(reflect.ValueOf("six"), reflect.ValueOf(6))
	six := refmap.MapIndex(reflect.ValueOf("six"))

	if six.Interface().(int) != 6 {
		t.Errorf("MakeMap map lookup 6 = %v", six)
	}

}

func TestNamedTypes(t *testing.T) {
	type namedString string

	named := namedString("foo")
	if got, want := reflect.TypeOf(named).Name(), "namedString"; got != want {
		t.Errorf("TypeOf.Name()=%v, want %v", got, want)
	}

}

func TestStruct(t *testing.T) {
	type barStruct struct {
		QuxString string
		BazInt    int
	}

	type foobar struct {
		Foo string
		Bar barStruct
	}

	var fb foobar
	fb.Bar.QuxString = "qux"

	reffb := reflect.TypeOf(fb)

	q := reffb.FieldByIndex([]int{1, 0})
	if want := "QuxString"; q.Name != want {
		t.Errorf("FieldByIndex=%v, want %v", q.Name, want)
	}

	var ok bool
	q, ok = reffb.FieldByName("Foo")
	if q.Name != "Foo" || !ok {
		t.Errorf("FieldByName(Foo)=%v,%v, want Foo, true")
	}

	q, ok = reffb.FieldByName("Snorble")
	if q.Name != "" || ok {
		t.Errorf("FieldByName(Snorble)=%v,%v, want ``, false")
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
