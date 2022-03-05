package interp

import (
	"strings"
	"testing"
)

func TestSmallIntValue(t *testing.T) {
	t.Parallel()

	cases := []struct {
		typ iType
		raw uint64
		str string
	}{
		// Test common integer values.
		{i8, 0, "i8 0"},
		{i8, 1, "i8 1"},
		{i8, uint64(^uint8(0)), "i8 -1"},
		{i16, 0, "i16 0"},
		{i16, 1, "i16 1"},
		{i16, uint64(^uint16(0)), "i16 -1"},
		{i32, 0, "i32 0"},
		{i32, 1, "i32 1"},
		{i32, uint64(^uint32(0)), "i32 -1"},
		{i64, 0, "i64 0"},
		{i64, 1, "i64 1"},
		{i64, ^uint64(0), "i64 -1"},

		// Test boolean values.
		{iType(1), 0, "i1 false"},
		{iType(1), 1, "i1 true"},

		// Test a weirdly sized value.
		{iType(4), 2, "i4 2"},

		// Test truncation.
		{i8, 256, "i8 0"},
	}
	for _, c := range cases {
		c := c
		t.Run(strings.ReplaceAll(c.str, " ", ":"), func(t *testing.T) {
			t.Parallel()

			v := smallIntValue(c.typ, c.raw)
			testPlainValue(t, v, c.typ, c.str)
		})
	}
}

func TestOffPtr(t *testing.T) {
	t.Parallel()

	defPtrTy := pointer(defaultAddrSpace, i32)
	altPtrTy := pointer(addrSpace(1), i16)
	x := memObj{
		ptrTy: defPtrTy,
		name:  "x",
	}
	y := memObj{
		ptrTy: altPtrTy,
		name:  "y",
	}
	cases := []struct {
		v   value
		ty  typ
		str string
	}{
		{x.ptr(0), defPtrTy, "ptr(in default, idx i32) @x"},
		{x.ptr(5), defPtrTy, "ptr(in default, idx i32) @x + 0x5"},
		{x.ptr((1 << 32) + 1), defPtrTy, "ptr(in default, idx i32) @x + 0x1"},
		{y.ptr(1), altPtrTy, "ptr(in addrspace(1), idx i16) @y + 0x1"},
	}
	replacer := strings.NewReplacer(
		", ", ",",
		" ", ":",
	)
	for _, c := range cases {
		c := c
		t.Run(replacer.Replace(c.str), func(t *testing.T) {
			t.Parallel()

			testPlainValue(t, c.v, c.ty, c.str)
		})
	}
}

func TestUglyGEP(t *testing.T) {
	t.Parallel()

	t.Skip("TODO")
}

func testPlainValue(t *testing.T, v value, ty typ, str string) {
	t.Helper()
	if vTy := v.typ(); vTy != ty {
		t.Errorf("wrong type: %s", vTy.String())
	}
	if vStr := v.String(); vStr != str {
		t.Errorf("expected %q but got %q", str, vStr)
	}
	if v != v {
		t.Error("value is not equivalent to itself")
	}
}

func TestCast(t *testing.T) {
	t.Parallel()

	ptrTy := pointer(defaultAddrSpace, i32)
	x := memObj{
		ptrTy:      ptrTy,
		name:       "x",
		alignScale: 2,
	}
	testExpressions(t,
		// Test a no-op cast.
		exprTest{
			name:   "Passthrough",
			expr:   cast(i32, runtime(i32, 1)),
			ty:     i32,
			expect: "i32 %1",
		},

		// Test a constant integer to pointer conversion.
		exprTest{
			name:   "ConstIntToPtr",
			expr:   cast(ptrTy, smallIntValue(i32, 123)),
			ty:     ptrTy,
			expect: "ptr(in default, idx i32) cast(i32 123)",
		},

		// Test converting an object memory address to an integer.
		exprTest{
			name:   "ConstObjAddr",
			expr:   cast(i32, x.ptr(3)),
			ty:     i32,
			expect: "i32 @x + 0x3",
		},

		// Test cast merging logic.
		exprTest{
			name:   "Cancel",
			expr:   cast(i32, cast(ptrTy, runtime(i32, 1))),
			ty:     i32,
			expect: "i32 %1",
		},
		exprTest{
			name:   "Merge",
			expr:   cast(i64, cast(ptrTy, runtime(i32, 1))),
			ty:     i64,
			expect: "i64 cast(i32 %1)",
		},
		exprTest{
			name:   "TruncateAndExtend",
			expr:   cast(i64, cast(i32, runtime(i64, 1))),
			ty:     i64,
			expect: "i64 cast(i32 cast(i64 %1))",
		},

		// Test integer resizing.
		exprTest{
			name:   "Truncate",
			expr:   cast(i32, smallIntValue(i64, (1<<32)+7)),
			ty:     i32,
			expect: "i32 7",
		},
		exprTest{
			name:   "ZeroExtend",
			expr:   cast(i64, smallIntValue(i32, uint64(^uint32(0)))),
			ty:     i64,
			expect: "i64 4294967295",
		},

		// Test undef handling.
		exprTest{
			name:   "ConvertUndef",
			expr:   cast(ptrTy, undefValue(i32)),
			ty:     ptrTy,
			expect: "ptr(in default, idx i32) undef",
		},
		exprTest{
			name:   "TruncateUndef",
			expr:   cast(i32, undefValue(i64)),
			ty:     i32,
			expect: "i32 undef",
		},
		exprTest{
			name:   "ExtendUndef",
			expr:   cast(i64, undefValue(i32)),
			ty:     i64,
			expect: "i64 cast(i32 undef)",
		},

		// Test resolution.
		exprTest{
			name: "Resolve",
			expr: cast(i32, runtime(i64, 0)).
				resolve([]value{smallIntValue(i64, 1)}),
			ty:     i32,
			expect: "i32 1",
		},

		// Test casting of concatenations.
		exprTest{
			name: "TruncCatToTrunc",
			expr: cast(ptrTy, cat([]value{
				x.ptr(0),
				smallIntValue(i32, 5),
			})),
			ty:     ptrTy,
			expect: "ptr(in default, idx i32) @x",
		},
		exprTest{
			name: "TrimCat",
			expr: cast(i32, cat([]value{
				smallIntValue(i8, 5),
				undefValue(i16),
				x.ptr(0),
				undefValue(i8),
			})),
			ty:     i32,
			expect: "i32 cat(i8 5, i16 undef, i8 cast(i32 @x))",
		},
		exprTest{
			name: "ZeroExtendCat",
			expr: cast(i64, cat([]value{
				smallIntValue(i16, 2),
				undefValue(i16),
			})),
			ty:     i64,
			expect: "i64 cat(i16 2, i16 undef, i32 0)",
		},
		exprTest{
			// This used to incorrectly strip the cast and produce a pointer-width integer.
			name: "ZeroExtendCatConvertRegression",
			expr: cast(ptrTy, cat([]value{
				smallIntValue(i8, 1),
				undefValue(i8),
			})),
			ty:     ptrTy,
			expect: "ptr(in default, idx i32) cast(i32 cat(i8 1, i8 undef, i16 0))",
		},

		// Test casting of slices.
		exprTest{
			name:   "TrimSlice",
			expr:   cast(i8, slice(x.ptr(0), 8, 16)),
			ty:     i8,
			expect: "i8 bitslice(i32 @x)[8:16]",
		},

		// Test truncation of a pointer to below its alignment scale.
		exprTest{
			name:   "AlignCheck",
			expr:   cast(iType(2), x.ptr(2)),
			ty:     iType(2),
			expect: "i2 -2",
		},
	)
}

func TestStructValues(t *testing.T) {
	t.Parallel()

	emptyStruct := &structType{}
	ptrTy := pointer(defaultAddrSpace, i32)
	strHeader := &structType{
		name: "runtime._string",
		fields: []structField{
			{ptrTy, 0},
			{i32, 4},
		},
		size: 8,
	}
	x := memObj{
		ptrTy: ptrTy,
		name:  "x",
	}
	testExpressions(t,
		// Test an empty struct value.
		exprTest{
			name:   "Empty",
			expr:   structValue(emptyStruct),
			ty:     emptyStruct,
			expect: "{} {}",
		},

		// Test a constant.
		exprTest{
			name: "Const",
			expr: structValue(strHeader,
				x.ptr(0),
				smallIntValue(i32, 3),
			),
			ty:     strHeader,
			expect: "%runtime._string {0: ptr(in default, idx i32) @x, 4: i32 3}",
		},

		// Test collapsing of an undefined constant.
		exprTest{
			name: "CollapseUndef",
			expr: structValue(strHeader,
				undefValue(ptrTy),
				undefValue(i32),
			),
			ty:     strHeader,
			expect: "%runtime._string undef",
		},

		// Test field extraction.
		exprTest{
			name:   "Extract",
			expr:   extractValue(structValue(strHeader, x.ptr(0), smallIntValue(i32, 3)), 1),
			ty:     iType(i32),
			expect: "i32 3",
		},
		exprTest{
			name:   "ExtractUndef",
			expr:   extractValue(undefValue(strHeader), 0),
			ty:     ptrTy,
			expect: "ptr(in default, idx i32) undef",
		},
		exprTest{
			name:   "ExtractSymbolic",
			expr:   extractValue(runtime(strHeader, 1), 1),
			ty:     i32,
			expect: "i32 extract 1 from %runtime._string %1",
		},

		// Test field insertion.
		exprTest{
			name:   "Insert",
			expr:   insertValue(insertValue(undefValue(strHeader), x.ptr(0), 0), smallIntValue(i32, 3), 1),
			ty:     strHeader,
			expect: "%runtime._string {0: ptr(in default, idx i32) @x, 4: i32 3}",
		},
		exprTest{
			name:   "InsertSymbolic",
			expr:   insertValue(runtime(strHeader, 1), runtime(i32, 2), 1),
			ty:     strHeader,
			expect: "%runtime._string {0: ptr(in default, idx i32) extract 0 from %runtime._string %1, 4: i32 %2}",
		},

		// Test field resolution.
		exprTest{
			name: "ResolveFields",
			expr: structValue(strHeader, runtime(ptrTy, 0), runtime(i32, 1)).
				resolve([]value{x.ptr(0), runtime(i32, 5)}),
			ty:     strHeader,
			expect: "%runtime._string {0: ptr(in default, idx i32) @x, 4: i32 %5}",
		},
		exprTest{
			name: "ResolveSymbolicExtract",
			expr: extractValue(runtime(strHeader, 0), 1).
				resolve([]value{structValue(strHeader, x.ptr(0), smallIntValue(i32, 3))}),
			ty:     i32,
			expect: "i32 3",
		},
	)
}

func TestArrayValues(t *testing.T) {
	t.Parallel()

	testExpressions(t,
		// Test an empty array value.
		exprTest{
			name:   "Empty",
			expr:   arrayValue(i32),
			ty:     array(i32, 0),
			expect: "[0 x i32] []",
		},

		// Test a constant.
		exprTest{
			name:   "Const",
			expr:   arrayValue(i32, smallIntValue(i32, 1), smallIntValue(i32, 2)),
			ty:     array(i32, 2),
			expect: "[2 x i32] [i32 1, i32 2]",
		},

		// Test collapsing of an undef constant.
		exprTest{
			name:   "CollapseUndef",
			expr:   arrayValue(i32, undefValue(i32), undefValue(i32)),
			ty:     array(i32, 2),
			expect: "[2 x i32] undef",
		},

		// Test element extraction.
		exprTest{
			name:   "Extract",
			expr:   extractValue(arrayValue(i32, smallIntValue(i32, 1), smallIntValue(i32, 2)), 1),
			ty:     i32,
			expect: "i32 2",
		},
		exprTest{
			name:   "ExtractUndef",
			expr:   extractValue(undefValue(array(i32, 5)), 3),
			ty:     i32,
			expect: "i32 undef",
		},
		exprTest{
			name:   "ExtractSymbolic",
			expr:   extractValue(runtime(array(i32, 2), 5), 1),
			ty:     i32,
			expect: "i32 extract 1 from [2 x i32] %5",
		},

		// Test element insertion.
		exprTest{
			name:   "Insert",
			expr:   insertValue(insertValue(undefValue(array(i32, 2)), smallIntValue(i32, 3), 0), smallIntValue(i32, 4), 1),
			ty:     array(i32, 2),
			expect: "[2 x i32] [i32 3, i32 4]",
		},
		exprTest{
			name:   "InsertSymbolic",
			expr:   insertValue(runtime(array(i32, 2), 7), runtime(i32, 8), 1),
			ty:     array(i32, 2),
			expect: "[2 x i32] [i32 extract 0 from [2 x i32] %7, i32 %8]",
		},

		// Test element resolution.
		exprTest{
			name: "ResolveElements",
			expr: arrayValue(i32, runtime(i32, 0), runtime(i32, 1)).
				resolve([]value{smallIntValue(i32, 2), smallIntValue(i32, 3)}),
			ty:     array(i32, 2),
			expect: "[2 x i32] [i32 2, i32 3]",
		},
		exprTest{
			name: "ResolveSymbolicExtract",
			expr: extractValue(runtime(array(i32, 2), 0), 1).
				resolve([]value{arrayValue(i32, smallIntValue(i32, 1), smallIntValue(i32, 2))}),
			ty:     i32,
			expect: "i32 2",
		},
	)
}

func TestCat(t *testing.T) {
	t.Parallel()

	rv1 := runtime(i64, 1)
	rv2 := cast(i16, runtime(iType(14), 2))
	testExpressions(t,
		// Test a single-element concatenation.
		exprTest{
			name:   "Single",
			expr:   cat([]value{runtime(i8, 2)}),
			ty:     i8,
			expect: "i8 %2",
		},

		// Test a concatenation of integers.
		exprTest{
			name: "ConstInts",
			expr: cat([]value{
				smallIntValue(i8, 1),
				smallIntValue(i16, 2),
				smallIntValue(i32, 3),
			}),
			ty:     iType(56),
			expect: "i56 50332161",
		},

		// Test a concatenation of undefined values.
		exprTest{
			name: "Undef",
			expr: cat([]value{
				undefValue(i8),
				undefValue(i16),
			}),
			ty:     iType(24),
			expect: "i24 undef",
		},

		// Test a zero-extension via concatenation.
		exprTest{
			name: "ZeroExtend",
			expr: cat([]value{
				runtime(i16, 3),
				smallIntValue(i16, 0),
			}),
			ty:     i32,
			expect: "i32 cast(i16 %3)",
		},

		// Test splitting of a zero-extended value.
		exprTest{
			name: "SplitZExt",
			expr: cat([]value{
				cast(i16, runtime(i8, 6)),
				smallIntValue(i16, 7),
			}),
			ty:     i32,
			expect: "i32 cat(i8 %6, i24 1792)",
		},

		// Test a re-combined value.
		exprTest{
			name: "Recombine",
			expr: cat([]value{
				slice(rv1, 0, 35),
				slice(rv1, 35, 29),
			}),
			ty:     i64,
			expect: "i64 %1",
		},
		exprTest{
			name: "RecombineMessy",
			expr: cast(iType(14), cat([]value{
				slice(rv2, 0, 8),
				slice(rv2, 8, 6),
			})),
			ty:     iType(14),
			expect: "i14 %2",
		},
		exprTest{
			name: "RecombinePartial",
			expr: cat([]value{
				slice(rv1, 8, 8),
				slice(rv1, 16, 16),
			}),
			ty:     iType(24),
			expect: "i24 bitslice(i64 %1)[8:32]",
		},

		// Test combining nested concatenations.
		exprTest{
			name: "NestedCats",
			expr: cat([]value{
				cat([]value{runtime(i8, 0), runtime(i8, 1)}),
				cat([]value{runtime(i8, 2), runtime(i8, 3)}),
			}),
			ty:     i32,
			expect: "i32 cat(i8 %0, i8 %1, i8 %2, i8 %3)",
		},

		// Test a complex value which cannot be simplified.
		exprTest{
			name: "Complex",
			expr: cat([]value{
				smallIntValue(i8, 2),
				undefValue(i8),
			}),
			ty:     i16,
			expect: "i16 cat(i8 2, i8 undef)",
		},

		// Test resolution of a concatenation.
		exprTest{
			name: "Resolve",
			expr: cat([]value{
				runtime(i8, 1),
				smallIntValue(i8, 2),
				undefValue(i16),
				runtime(i16, 0),
			}).resolve([]value{
				undefValue(i16),
				smallIntValue(i8, 1),
			}),
			ty:     iType(48),
			expect: "i48 cat(i16 513, i32 undef)",
		},
	)
}

func TestSlice(t *testing.T) {
	t.Parallel()

	x := memObj{
		ptrTy:      pointer(defaultAddrSpace, i32),
		name:       "x",
		alignScale: 2,
	}
	testExpressions(t,
		// Test a no-op slice.
		exprTest{
			name:   "Pasthrough",
			expr:   slice(runtime(i16, 2), 0, 16),
			ty:     i16,
			expect: "i16 %2",
		},

		// Test slicing an integer constant.
		exprTest{
			name:   "ConstInt",
			expr:   slice(smallIntValue(i16, (2<<8)|3), 8, 8),
			ty:     i8,
			expect: "i8 2",
		},

		// Test slicing undef.
		exprTest{
			name:   "Undef",
			expr:   slice(undefValue(i64), 16, 32),
			ty:     i32,
			expect: "i32 undef",
		},

		// Test a slice of an unknown value (cannot be simplified).
		exprTest{
			name:   "Unknown",
			expr:   slice(runtime(i16, 2), 8, 8),
			ty:     i8,
			expect: "i8 bitslice(i16 %2)[8:16]",
		},

		// Test slicing a concatenation.
		exprTest{
			name:   "Cat",
			expr:   slice(cat([]value{smallIntValue(i8, 1), undefValue(iType(24))}), 16, 8),
			ty:     i8,
			expect: "i8 undef",
		},
		exprTest{
			name:   "CatComplex",
			expr:   slice(cat([]value{smallIntValue(i16, 3<<7), undefValue(i16)}), 8, 16),
			ty:     i16,
			expect: "i16 cat(i8 1, i8 undef)",
		},
		exprTest{
			name: "SliceOfCatToCatOfSlice",
			expr: slice(cat([]value{
				runtime(i16, 1),
				runtime(i16, 2),
				runtime(i16, 3),
			}), 8, 16),
			ty:     i16,
			expect: "i16 cat(i8 bitslice(i16 %1)[8:16], i8 cast(i16 %2))",
		},

		// Test that a slice with an offset of zero can be reduced to a truncation.
		exprTest{
			name:   "Trunc",
			expr:   slice(runtime(i16, 2), 0, 8),
			ty:     i8,
			expect: "i8 cast(i16 %2)",
		},

		// Test slicing of a zero-extension.
		exprTest{
			name:   "SliceZExtToZExtSlice",
			expr:   slice(cast(i32, runtime(i16, 2)), 8, 16),
			ty:     i16,
			expect: "i16 cast(i8 bitslice(i16 %2)[8:16])",
		},
		exprTest{
			name:   "SliceZExtPadding",
			expr:   slice(cast(i32, runtime(i16, 2)), 16, 8),
			ty:     i8,
			expect: "i8 0",
		},

		// Test combining of nested slices.
		exprTest{
			name:   "Nest",
			expr:   slice(slice(runtime(i64, 2), 8, 32), 16, 8),
			ty:     i8,
			expect: "i8 bitslice(i64 %2)[24:32]",
		},

		// Test slice resolution.
		exprTest{
			name: "Resolve",
			expr: slice(runtime(i16, 0), 8, 8).
				resolve([]value{smallIntValue(i16, (3<<8)|7)}),
			ty:     i8,
			expect: "i8 3",
		},

		// Test slicing of pointer alignment bits.
		exprTest{
			name:   "AlignCheck",
			expr:   slice(x.ptr(2), 1, 1),
			ty:     iType(1),
			expect: "i1 true",
		},
	)
}

func testExpressions(t *testing.T, cases ...exprTest) {
	t.Helper()

	for _, c := range cases {
		c := c
		t.Run(c.name, c.run)
	}
}

type exprTest struct {
	name   string
	expr   value
	ty     typ
	expect string
}

func (c *exprTest) run(t *testing.T) {
	t.Parallel()

	if str := c.expr.String(); str != c.expect {
		t.Errorf("expected %q but got %q", c.expect, str)
	}

	if ty := c.expr.typ(); ty != c.ty {
		t.Errorf("expected result type %q but got %q", c.ty.String(), ty.String())
	}

	if c.expr != c.expr {
		t.Error("expression is not equivalent to itself")
	}
}
