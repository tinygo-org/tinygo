package interp

import (
	"fmt"
	"strings"
	"testing"
)

func TestIType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ty    iType
		bits  uint64
		bytes uint64
		str   string
	}{
		// Test common integer types.
		{iType(1), 1, 1, "i1"},
		{i8, 8, 1, "i8"},
		{i16, 16, 2, "i16"},
		{i32, 32, 4, "i32"},
		{i64, 64, 8, "i64"},

		// Test weird sizes.
		{iType(7), 7, 1, "i7"},
		{iType(9), 9, 2, "i9"},
		{iType(63), 63, 8, "i63"},

		// TODO: test big integers
	}

	for _, c := range cases {
		c := c
		t.Run(c.str, func(t *testing.T) {
			t.Parallel()

			testNonAggTy(t, c.ty, c.bits, c.bytes, c.str)
		})
	}
}

func TestPtrType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		idxTy iType
		in    addrSpace
		bits  uint64
		bytes uint64
		str   string
	}{
		// Test common pointer types.
		{i16, defaultAddrSpace, 16, 2, "ptr(in default, idx i16)"},
		{i32, defaultAddrSpace, 32, 4, "ptr(in default, idx i32)"},
		{i64, defaultAddrSpace, 64, 8, "ptr(in default, idx i64)"},

		// Test pointers in other address spaces.
		{i16, addrSpace(1), 16, 2, "ptr(in addrspace(1), idx i16)"},
		{iType(23), addrSpace(1), 23, 3, "ptr(in addrspace(1), idx i23)"},

		// Test near-overflow conditions.
		{i64, addrSpace((1 << (32 - 6)) - 1), 64, 8, "ptr(in addrspace(67108863), idx i64)"},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d-i%d", uint(c.in), uint(c.idxTy)), func(t *testing.T) {
			t.Parallel()

			ty := pointer(c.in, c.idxTy)
			if idxTy := ty.idxTy(); idxTy != c.idxTy {
				t.Errorf("expected index type %s but got %s", c.idxTy.String(), idxTy.String())
			}
			testNonAggTy(t, ty, c.bits, c.bytes, c.str)
		})
	}
}

func testNonAggTy(t *testing.T, ty nonAggTyp, bits, bytes uint64, str string) {
	t.Helper()

	if tBits := ty.bits(); tBits != bits {
		t.Errorf("expected %d bits but got %d", bits, tBits)
	}
	testTy(t, ty, bytes, str)
}

func TestStruct(t *testing.T) {
	t.Parallel()

	cases := []struct {
		ty        *structType
		elemTypes []typ
		bytes     uint64
		str       string
	}{
		// Test an empty struct.
		{&structType{}, nil, 0, "{}"},

		// Test a struct with a single field.
		{
			ty: &structType{
				fields: []structField{
					{i8, 0},
				},
				size: 1,
			},
			elemTypes: []typ{i8},
			bytes:     1,
			str:       "{0: i8}",
		},

		// Test a struct with multiple fields.
		{
			ty: &structType{
				fields: []structField{
					{i8, 0},
					{iType(1), 1},
				},
				size: 2,
			},
			elemTypes: []typ{i8, iType(1)},
			bytes:     2,
			str:       "{0: i8, 1: i1}",
		},

		// Test a struct with padding.
		{
			ty: &structType{
				fields: []structField{
					{i8, 0},
					{i64, 8},
					{iType(1), 16},
				},
				size: 24,
			},
			elemTypes: []typ{i8, i64, iType(1)},
			bytes:     24,
			str:       "{0: i8, 1: _ x 7, 8: i64, 16: i1, 17: _ x 7}",
		},

		// Test a struct consisting of a single zero-sized field.
		{
			ty: &structType{
				fields: []structField{
					{array(iType(1), 0), 0},
				},
				size: 0,
			},
			elemTypes: []typ{array(iType(1), 0)},
			bytes:     0,
			str:       "{0: [0 x i1]}",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(strings.ReplaceAll(c.str, " ", ""), func(t *testing.T) {
			t.Parallel()

			testAggTy(t, c.ty, c.elemTypes, c.bytes, c.str)
		})
	}
}

func TestArray(t *testing.T) {
	t.Parallel()

	cases := []struct {
		of    typ
		n     uint32
		bytes uint64
		str   string
	}{
		// Test a regular byte array.
		{i8, 5, 5, "[5 x i8]"},

		// Test an array with a multi-byte element type.
		{i16, 64, 128, "[64 x i16]"},

		// Test an array of sub-byte elements.
		{iType(1), 1024, 1024, "[1024 x i1]"},

		// Test a zero-length array.
		{i64, 0, 0, "[0 x i64]"},

		// Test an array with zero-size elements.
		{&structType{}, 5, 0, "[5 x {}]"},

		// Test a zero-length array with zero-size elements.
		{&structType{}, 0, 0, "[0 x {}]"},

		// Test a multidimensional array.
		{array(i8, 5), 10, 50, "[10 x [5 x i8]]"},
	}

	for _, c := range cases {
		c := c
		t.Run(strings.ReplaceAll(c.str, " ", ""), func(t *testing.T) {
			t.Parallel()

			ty := array(c.of, c.n)
			ty2 := array(c.of, c.n)
			if ty2 != ty {
				t.Error("duplicate type is not equal")
			}
			elemTypes := make([]typ, c.n)
			for i := range elemTypes {
				elemTypes[i] = c.of
			}
			testAggTy(t, ty, elemTypes, c.bytes, c.str)
		})
	}
}

func testAggTy(t *testing.T, ty aggTyp, elemTypes []typ, bytes uint64, str string) {
	t.Helper()

	n := ty.elems()
	if n != uint32(len(elemTypes)) {
		t.Errorf("expected %d element types but found %d", len(elemTypes), n)
	} else {
		for i, ee := range elemTypes {
			e := ty.sub(uint32(i))
			if e != ee {
				t.Errorf("expected %s at element %d but got %s", ee.String(), i, e.String())
			}
		}
	}

	testTy(t, ty, bytes, str)
}

func testTy(t *testing.T, ty typ, bytes uint64, str string) {
	t.Helper()

	if tBytes := ty.bytes(); tBytes != bytes {
		t.Errorf("expected %d bytes but got %d", bytes, tBytes)
	}
	if tStr := ty.String(); tStr != str {
		t.Errorf("expected name %q but got %q", str, tStr)
	}
	if ty != ty {
		t.Error("type is not equal to itself")
	}
}
