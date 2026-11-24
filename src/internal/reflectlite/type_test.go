package reflectlite_test

import (
	"internal/reflectlite"
	"testing"
	"unsafe"
)

type Embedded int

type fields struct {
	_     [130]byte `first:"blank"`
	Value byte
	last  byte `last:"field"`
	Embedded
}

func TestFieldMetadata(t *testing.T) {
	var value fields
	typ := reflectlite.TypeOf(value).(*reflectlite.RawType)
	for i, want := range []struct {
		name      string
		pkgPath   string
		typ       reflectlite.Type
		tag       reflectlite.StructTag
		offset    uintptr
		anonymous bool
	}{
		{"_", typ.PkgPath(), reflectlite.TypeOf([130]byte{}), `first:"blank"`, 0, false},
		{"Value", "", reflectlite.TypeOf(byte(0)), "", unsafe.Offsetof(value.Value), false},
		{"last", typ.PkgPath(), reflectlite.TypeOf(byte(0)), `last:"field"`, unsafe.Offsetof(value.last), false},
		{"Embedded", "", reflectlite.TypeOf(Embedded(0)), "", unsafe.Offsetof(value.Embedded), true},
	} {
		got := typ.Field(i)
		if got.Name != want.name || got.PkgPath != want.pkgPath || got.Type != want.typ || got.Tag != want.tag || got.Offset != want.offset || got.Anonymous != want.anonymous {
			t.Errorf("Field(%d) = %+v, want %+v", i, got, want)
		}
		if len(got.Index) != 1 || got.Index[0] != i {
			t.Errorf("Field(%d).Index = %v, want [%d]", i, got.Index, i)
		}
	}
}

func TestFieldPanics(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		index int
	}{
		{"not struct", 0, 0},
		{"empty struct", struct{}{}, 0},
		{"negative index", fields{}, -1},
		{"large index", fields{}, 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("Field did not panic")
				}
			}()
			reflectlite.TypeOf(test.value).(*reflectlite.RawType).Field(test.index)
		})
	}
}

type inner struct{ X int16 }

type outer struct {
	inner
	Y int8
	z int32
}

func TestValueFieldFlags(t *testing.T) {
	v := reflectlite.ValueOf(&outer{inner{1}, 2, 3}).Elem()
	for _, test := range []struct {
		name   string
		field  reflectlite.Value
		canSet bool
		value  int64
	}{
		{"inner.X", v.Field(0).Field(0), true, 1},
		{"Y", v.Field(1), true, 2},
		{"z", v.Field(2), false, 3},
	} {
		if got := test.field.CanSet(); got != test.canSet {
			t.Errorf("%s: CanSet() = %v, want %v", test.name, got, test.canSet)
		}
		if got := test.field.Int(); got != test.value {
			t.Errorf("%s: Int() = %d, want %d", test.name, got, test.value)
		}
	}
	if v.Field(0).CanSet() || v.Field(0).CanInterface() {
		t.Error("unexported embedded field can be set or used as an interface")
	}
	if !reflectlite.ValueOf(outer{}).Field(1).CanInterface() {
		t.Error("exported field of an unaddressable struct cannot be used as an interface")
	}
}

// Interface lowering reads byte 1 of a descriptor as a type ID, so descriptors
// made at runtime must have type ID 0.
func TestDynamicTypeIDByte(t *testing.T) {
	if unsafe.Alignof(uint16(0)) < 2 {
		t.Skip("no type ID field on this target")
	}
	st := reflectlite.StructOf([]reflectlite.StructField{{Name: "X", Type: reflectlite.TypeOf(0)}})
	// Pointers to st are tagged up to 3 levels, so the fourth is allocated.
	ptr := st
	for range 4 {
		ptr = reflectlite.PointerTo(ptr)
	}
	if uintptr(unsafe.Pointer(ptr.(*reflectlite.RawType)))&3 != 0 {
		t.Fatal("PointerTo did not allocate a pointer type")
	}
	for _, tc := range []struct {
		name string
		typ  reflectlite.Type
	}{
		{"StructOf", st},
		{"PointerTo", ptr},
		{"SliceOf", reflectlite.SliceOf(st)},
		{"ArrayOf", reflectlite.ArrayOf(3, st)},
		{"ChanOf", reflectlite.ChanOf(reflectlite.BothDir, st)},
		{"MapOf", reflectlite.MapOf(st, st)},
	} {
		raw := tc.typ.(*reflectlite.RawType)
		if id := *(*uint8)(unsafe.Add(unsafe.Pointer(raw), 1)); id != 0 {
			t.Errorf("%s: byte 1 is %d, want 0", tc.name, id)
		}
	}
}

type taggedIface interface{ Tagged() }

type tagged0 struct{ x int }

func (*tagged0) Tagged() {}

type tagged1 struct{ x int }

func (*tagged1) Tagged() {}

type tagged2 struct{ x int }

func (*tagged2) Tagged() {}

type tagged3 struct{ x int }

func (*tagged3) Tagged() {}

type tagged4 struct{ x int }

func (*tagged4) Tagged() {}

type tagged5 struct{ x int }

func (*tagged5) Tagged() {}

type tagged6 struct{ x int }

func (*tagged6) Tagged() {}

type tagged7 struct{ x int }

func (*tagged7) Tagged() {}

type tagged8 struct{ x int }

func (*tagged8) Tagged() {}

type tagged9 struct{ x int }

func (*tagged9) Tagged() {}

type tagged10 struct{ x int }

func (*tagged10) Tagged() {}

type tagged11 struct{ x int }

func (*tagged11) Tagged() {}

type tagged12 struct{ x int }

func (*tagged12) Tagged() {}

type tagged13 struct{ x int }

func (*tagged13) Tagged() {}

type tagged14 struct{ x int }

func (*tagged14) Tagged() {}

type tagged15 struct{ x int }

func (*tagged15) Tagged() {}

type tagged16 struct{ x int }

func (*tagged16) Tagged() {}

type tagged17 struct{ x int }

func (*tagged17) Tagged() {}

type tagged18 struct{ x int }

func (*tagged18) Tagged() {}

type tagged19 struct{ x int }

func (*tagged19) Tagged() {}

type tagged20 struct{ x int }

func (*tagged20) Tagged() {}

type tagged21 struct{ x int }

func (*tagged21) Tagged() {}

type tagged22 struct{ x int }

func (*tagged22) Tagged() {}

type tagged23 struct{ x int }

func (*tagged23) Tagged() {}

type tagged24 struct{ x int }

func (*tagged24) Tagged() {}

type tagged25 struct{ x int }

func (*tagged25) Tagged() {}

type tagged26 struct{ x int }

func (*tagged26) Tagged() {}

type tagged27 struct{ x int }

func (*tagged27) Tagged() {}

type tagged28 struct{ x int }

func (*tagged28) Tagged() {}

type tagged29 struct{ x int }

func (*tagged29) Tagged() {}

type tagged30 struct{ x int }

func (*tagged30) Tagged() {}

type tagged31 struct{ x int }

func (*tagged31) Tagged() {}

type tagged32 struct{ x int }

func (*tagged32) Tagged() {}

type tagged33 struct{ x int }

func (*tagged33) Tagged() {}

type tagged34 struct{ x int }

func (*tagged34) Tagged() {}

type tagged35 struct{ x int }

func (*tagged35) Tagged() {}

type tagged36 struct{ x int }

func (*tagged36) Tagged() {}

type tagged37 struct{ x int }

func (*tagged37) Tagged() {}

type tagged38 struct{ x int }

func (*tagged38) Tagged() {}

type tagged39 struct{ x int }

func (*tagged39) Tagged() {}

// Enough implementers that their pointer types get type IDs up to at least 40.
var taggedValues = []taggedIface{&tagged0{}, &tagged1{}, &tagged2{}, &tagged3{}, &tagged4{}, &tagged5{}, &tagged6{}, &tagged7{}, &tagged8{}, &tagged9{}, &tagged10{}, &tagged11{}, &tagged12{}, &tagged13{}, &tagged14{}, &tagged15{}, &tagged16{}, &tagged17{}, &tagged18{}, &tagged19{}, &tagged20{}, &tagged21{}, &tagged22{}, &tagged23{}, &tagged24{}, &tagged25{}, &tagged26{}, &tagged27{}, &tagged28{}, &tagged29{}, &tagged30{}, &tagged31{}, &tagged32{}, &tagged33{}, &tagged34{}, &tagged35{}, &tagged36{}, &tagged37{}, &tagged38{}, &tagged39{}}

// A tagged pointer to pointer type points into the descriptor of the pointer
// type, at the type ID byte or numMethod. It must not be read as a meta byte.
func TestTaggedPointerImplements(t *testing.T) {
	itf := reflectlite.TypeOf((*taggedIface)(nil)).Elem()
	for i, v := range taggedValues {
		v.Tagged() // so that the implementers get type IDs
		typ := reflectlite.TypeOf(v)
		if !typ.Implements(itf) {
			t.Errorf("%d: %s does not implement taggedIface", i, typ)
		}
		for depth := 2; depth <= 4; depth++ {
			typ = reflectlite.PointerTo(typ)
			if typ.Implements(itf) || typ.AssignableTo(itf) {
				t.Errorf("%d: %s implements taggedIface", i, typ)
			}
		}
	}
}
