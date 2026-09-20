// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect_test

import (
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestLongStructTags(t *testing.T) {
	typ := reflect.TypeOf(struct {
		Empty int
		Short int `x`
		At127 int `1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567`
		At128 int `12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678`
		At255 int `123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345`
		At256 int `1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456`
	}{})
	for i, length := range []int{0, 1, 127, 128, 255, 256} {
		want := (strings.Repeat("1234567890", 26))[:length]
		if length == 1 {
			want = "x"
		}
		field := typ.Field(i)
		t.Run(field.Name, func(t *testing.T) {
			if got := string(field.Tag); got != want {
				t.Errorf("Field(%d).Tag has length %d, want %d", i, len(got), length)
			}
			byName, ok := typ.FieldByName(field.Name)
			if !ok || string(byName.Tag) != want {
				t.Error("FieldByName returned an incorrect tag")
			}
			byFunc, ok := typ.FieldByNameFunc(func(name string) bool { return name == field.Name })
			if !ok || string(byFunc.Tag) != want {
				t.Error("FieldByNameFunc returned an incorrect tag")
			}
			if length != 0 && !strings.Contains(typ.String(), string(field.Tag)) {
				t.Error("Type.String omitted the tag")
			}
		})
	}
}

func TestTypeFieldByIndexEmpty(t *testing.T) {
	type item struct {
		A string `json:"a"`
	}
	for _, typ := range []reflect.Type{
		reflect.TypeFor[item](),
		reflect.TypeFor[struct{}](),
	} {
		t.Run(typ.String(), func(t *testing.T) {
			for _, test := range []struct {
				name  string
				index []int
			}{
				{"nil", nil},
				{"empty", []int{}},
			} {
				t.Run(test.name, func(t *testing.T) {
					field := typ.FieldByIndex(test.index)
					if field.Type != typ {
						t.Fatal("FieldByIndex returned a different type")
					}
					if got := field.Type.Kind(); got != reflect.Struct {
						t.Fatalf("unexpected field kind: got %v want %v", got, reflect.Struct)
					}
					if want := (reflect.StructField{Type: typ}); !reflect.DeepEqual(field, want) {
						t.Fatalf("unexpected field result\ngot %#v\nwant %#v", field, want)
					}
				})
			}
		})
	}
}

func TestTypeFieldByIndexFields(t *testing.T) {
	type fields struct {
		Pad    byte
		Name   string `json:"name"`
		hidden int
	}
	typ := reflect.TypeFor[fields]()
	embedded := reflect.TypeFor[struct {
		Pad byte
		fields
	}]()
	pointer := reflect.TypeFor[struct {
		Pad byte
		*fields
	}]()
	for _, test := range []struct {
		name  string
		typ   reflect.Type
		index []int
		want  reflect.StructField
	}{
		{"direct", typ, []int{1}, typ.Field(1)},
		{"unexported", typ, []int{2}, typ.Field(2)},
		{"embedded", embedded, []int{1}, embedded.Field(1)},
		{"promoted", embedded, []int{1, 1}, typ.Field(1)},
		{"pointer", pointer, []int{1}, pointer.Field(1)},
		{"pointer-promoted", pointer, []int{1, 1}, typ.Field(1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			field := test.typ.FieldByIndex(test.index)
			if field.Type != test.want.Type {
				t.Fatal("FieldByIndex returned a different type")
			}
			if got, want := field.Type.Kind(), test.want.Type.Kind(); got != want {
				t.Fatalf("unexpected field kind: got %v want %v", got, want)
			}
			field.Index = test.want.Index
			if !reflect.DeepEqual(field, test.want) {
				t.Fatalf("unexpected field result\ngot %#v\nwant %#v", field, test.want)
			}
		})
	}
}

func TestTypeFieldByIndexNonStruct(t *testing.T) {
	if runtime.GOARCH == "wasm" {
		t.Log("recover is not supported on wasm")
		return
	}

	type item struct {
		A string
	}
	for _, typ := range []reflect.Type{
		reflect.TypeFor[int](),
		reflect.TypeFor[*item](),
	} {
		t.Run(typ.String(), func(t *testing.T) {
			for _, test := range []struct {
				name  string
				index []int
			}{
				{"nil", nil},
				{"empty", []int{}},
				{"field", []int{0}},
			} {
				t.Run(test.name, func(t *testing.T) {
					defer func() { recover() }()
					typ.FieldByIndex(test.index)
					t.Fatal("FieldByIndex did not panic")
				})
			}
		})
	}
}

func checkConstructedType[T any](t *testing.T, value reflect.Value) {
	t.Helper()
	if _, ok := value.Interface().(T); !ok {
		t.Fatalf("constructed %v failed a concrete type assertion", value.Type())
	}
	switch value.Interface().(type) {
	case T:
	default:
		t.Fatalf("constructed %v failed a type switch", value.Type())
	}
}

func TestArrayOfRuntimeConstruction(t *testing.T) {
	type elem struct {
		Value *int
	}

	elemType := reflect.TypeOf(elem{})
	compiledType := reflect.TypeOf([3]elem{})
	if got := reflect.ArrayOf(3, elemType); got != compiledType {
		t.Fatalf("ArrayOf(3, elem) = %v, want existing type %v", got, compiledType)
	}

	arrayType := reflect.ArrayOf(4, elemType)
	if got, want := arrayType.String(), "[4]reflect_test.elem"; got != want {
		t.Fatalf("ArrayOf(4, elem).String() = %q, want %q", got, want)
	}
	if got := arrayType.Kind(); got != reflect.Array {
		t.Fatalf("ArrayOf(4, elem).Kind() = %v, want %v", got, reflect.Array)
	}
	if got := arrayType.Len(); got != 4 {
		t.Fatalf("ArrayOf(4, elem).Len() = %d, want 4", got)
	}
	if got := arrayType.Elem(); got != elemType {
		t.Fatalf("ArrayOf(4, elem).Elem() = %v, want %v", got, elemType)
	}
	if got := reflect.ArrayOf(4, elemType); got != arrayType {
		t.Fatalf("second ArrayOf(4, elem) = %v, want %v", got, arrayType)
	}
	if got := reflect.PointerTo(arrayType).Elem(); got != arrayType {
		t.Fatalf("PointerTo(ArrayOf(4, elem)).Elem() = %v, want %v", got, arrayType)
	}

	value := 42
	array := reflect.New(arrayType).Elem()
	array.Index(0).Field(0).Set(reflect.ValueOf(&value))
	runtime.GC()
	if got := array.Index(0).Field(0).Elem().Int(); got != int64(value) {
		t.Fatalf("array element after GC = %d, want %d", got, value)
	}
	checkConstructedType[[4]elem](t, array)
	checkConstructedType[*[4]elem](t, array.Addr())

	checkPanic := func(name string, fn func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Errorf("%s did not panic", name)
			}
		}()
		fn()
	}
	checkPanic("negative ArrayOf length", func() {
		reflect.ArrayOf(-1, elemType)
	})
	checkPanic("overflowing ArrayOf size", func() {
		reflect.ArrayOf(int(^uint(0)>>1), reflect.TypeOf(int(0)))
	})
}

func TestMapOfRuntimeConstruction(t *testing.T) {
	type key string
	type elem float64

	keyType := reflect.TypeOf(key(""))
	elemType := reflect.TypeOf(elem(0))
	compiledType := reflect.TypeOf(map[key]elem(nil))
	mapType := reflect.MapOf(keyType, elemType)
	if mapType != compiledType {
		t.Fatalf("MapOf(key, elem) = %v, want existing type %v", mapType, compiledType)
	}
	if got, want := mapType.String(), "map[reflect_test.key]reflect_test.elem"; got != want {
		t.Fatalf("MapOf(key, elem).String() = %q, want %q", got, want)
	}
	if got := mapType.Key(); got != keyType {
		t.Fatalf("MapOf(key, elem).Key() = %v, want %v", got, keyType)
	}
	if got := mapType.Elem(); got != elemType {
		t.Fatalf("MapOf(key, elem).Elem() = %v, want %v", got, elemType)
	}
	if got := reflect.MapOf(keyType, elemType); got != mapType {
		t.Fatalf("second MapOf(key, elem) = %v, want %v", got, mapType)
	}

	m := reflect.MakeMap(mapType)
	m.SetMapIndex(reflect.ValueOf(key("a")), reflect.ValueOf(elem(1)))
	runtime.GC()
	if got := m.MapIndex(reflect.ValueOf(key("a"))).Float(); got != 1 {
		t.Fatalf("constructed map value = %v, want 1", got)
	}

	pointerType := reflect.TypeOf((*int)(nil))
	pointerMap := reflect.MakeMap(reflect.MapOf(pointerType, pointerType))
	keyPointer := new(int)
	valuePointer := new(int)
	*valuePointer = 42
	pointerMap.SetMapIndex(reflect.ValueOf(keyPointer), reflect.ValueOf(valuePointer))
	runtime.GC()
	if got := pointerMap.MapIndex(reflect.ValueOf(keyPointer)).Elem().Int(); got != 42 {
		t.Fatalf("pointer map element after GC = %d, want 42", got)
	}

	type assertionElem uint32
	assertionType := reflect.TypeOf(assertionElem(0))
	checkConstructedType[map[assertionElem]assertionElem](t,
		reflect.MakeMap(reflect.MapOf(assertionType, assertionType)))

	defer func() {
		if recover() == nil {
			t.Fatal("MapOf accepted an invalid key type")
		}
	}()
	reflect.MapOf(reflect.TypeOf((func())(nil)), reflect.TypeOf(false))
}

func TestTypeFor(t *testing.T) {
	type (
		mystring string
		myiface  interface{}
	)

	testcases := []struct {
		wantFrom any
		got      reflect.Type
	}{
		{new(int), reflect.TypeFor[int]()},
		{new(int64), reflect.TypeFor[int64]()},
		{new(string), reflect.TypeFor[string]()},
		{new(mystring), reflect.TypeFor[mystring]()},
		{new(any), reflect.TypeFor[any]()},
		{new(myiface), reflect.TypeFor[myiface]()},
	}
	for _, tc := range testcases {
		want := reflect.ValueOf(tc.wantFrom).Elem().Type()
		if want != tc.got {
			t.Errorf("unexpected reflect.Type: got %v; want %v", tc.got, want)
		}
	}
}
func TestElemOfNamedMultiPointer(t *testing.T) {
	type recursive ***recursive

	tests := []struct {
		typ  reflect.Type
		want reflect.Type
	}{
		{reflect.TypeFor[recursive](), reflect.TypeFor[**recursive]()},
		{reflect.TypeFor[**recursive](), reflect.TypeFor[*recursive]()},
		{reflect.TypeFor[*recursive](), reflect.TypeFor[recursive]()},
	}
	for _, test := range tests {
		if got := test.typ.Elem(); got != test.want {
			t.Errorf("%v.Elem() = %v; want %v", test.typ, got, test.want)
		}
	}
}

func TestReflectedPointerTypeAssertions(t *testing.T) {
	type named uint32

	value := reflect.New(reflect.TypeOf(0)).Interface()
	if _, ok := value.(*int); !ok {
		t.Fatalf("constructed %T failed a type assertion", value)
	}

	pointer := reflect.New(reflect.TypeOf(named(0)))
	value = pointer.Interface()
	if _, ok := value.(*named); !ok {
		t.Fatalf("constructed %T failed a type assertion", value)
	}

	value = reflect.New(pointer.Type()).Interface()
	if _, ok := value.(**named); !ok {
		t.Fatalf("constructed %T failed a type assertion", value)
	}
}

func TestSliceOfRuntimeConstruction(t *testing.T) {
	type compiledElem int
	compiledType := reflect.TypeOf([]compiledElem(nil))
	if got := reflect.SliceOf(reflect.TypeOf(compiledElem(0))); got != compiledType {
		t.Fatalf("SliceOf(compiledElem) = %v, want existing type %v", got, compiledType)
	}

	type elem struct {
		Value *int
	}

	elemType := reflect.TypeOf(elem{})
	sliceType := reflect.SliceOf(elemType)
	if got, want := sliceType.String(), "[]reflect_test.elem"; got != want {
		t.Fatalf("SliceOf(elem).String() = %q, want %q", got, want)
	}
	if got := sliceType.Kind(); got != reflect.Slice {
		t.Fatalf("SliceOf(elem).Kind() = %v, want %v", got, reflect.Slice)
	}
	if got := sliceType.Elem(); got != elemType {
		t.Fatalf("SliceOf(elem).Elem() = %v, want %v", got, elemType)
	}
	if got := reflect.SliceOf(elemType); got != sliceType {
		t.Fatalf("second SliceOf(elem) = %v, want %v", got, sliceType)
	}
	if got := reflect.PointerTo(sliceType).Elem(); got != sliceType {
		t.Fatalf("PointerTo(SliceOf(elem)).Elem() = %v, want %v", got, sliceType)
	}

	value := 42
	slice := reflect.MakeSlice(sliceType, 1, 1)
	slice.Index(0).Field(0).Set(reflect.ValueOf(&value))
	checkConstructedType[[]elem](t, slice)
	runtime.GC()
	if got := slice.Index(0).Field(0).Elem().Int(); got != int64(value) {
		t.Fatalf("slice element after GC = %d, want %d", got, value)
	}
}
