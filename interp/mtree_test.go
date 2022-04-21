package interp

import (
	"reflect"
	"testing"
)

func TestLeaf(t *testing.T) {
	t.Parallel()

	ptrTy := pointer(defaultAddrSpace, i32)
	x := memObj{
		ptrTy: ptrTy,
		name:  "x",
	}

	// Test stores into a leaf.
	t.Run("Store", func(t *testing.T) {
		cases := []struct {
			name    string
			into    memLeaf
			version uint64
			v       value
			off     uint64
			result  memLeaf
		}{
			// Test a store of a small integer.
			{
				name:    "SmallInt",
				into:    memLeaf{},
				version: 1,
				v:       smallIntValue(i64, 0x8877665544332211),
				off:     4,
				result: memLeaf{
					metaHigh: 0b11111111 << 4,
					metaLow:  0b11111111 << 4,
					pending:  0b11111111 << 4,
					version:  1,
					data: [64]byte{
						0x4: 0x11, 0x5: 0x22, 0x6: 0x33, 0x7: 0x44,
						0x8: 0x55, 0x9: 0x66, 0xA: 0x77, 0xB: 0x88,
					},
				},
			},

			// Test that a booolean can be stored as a byte.
			{
				name:    "Bool",
				into:    memLeaf{},
				version: 1,
				v:       smallIntValue(iType(1), 1),
				off:     1,
				result: memLeaf{
					metaHigh: 1 << 1,
					metaLow:  1 << 1,
					pending:  1 << 1,
					version:  1,
					data: [64]byte{
						0x1: 0x01,
					},
				},
			},

			// Test that a pointer can be stored as a special.
			{
				name:    "Ptr",
				into:    memLeaf{},
				version: 1,
				v:       x.ptr(0),
				off:     8,
				result: memLeaf{
					metaHigh:      0b1111 << 8,
					specialStarts: 0b0001 << 8,
					pending:       0b1111 << 8,
					version:       1,
					specials:      []value{x.ptr(0)},
				},
			},

			// Test a store of a simple undef value.
			{
				name:    "Undef",
				into:    memLeaf{},
				version: 1,
				v:       undefValue(i16),
				off:     2,
				result: memLeaf{
					metaLow: 0b11 << 2,
					pending: 0b11 << 2,
					version: 1,
				},
			},

			// Test a store of an undef boolean.
			// This must be stored as a special.
			{
				name:    "UndefBool",
				into:    memLeaf{},
				version: 1,
				v:       undefValue(iType(1)),
				off:     3,
				result: memLeaf{
					metaHigh:      1 << 3,
					specialStarts: 1 << 3,
					pending:       1 << 3,
					version:       1,
					specials:      []value{undefValue(iType(1))},
				},
			},

			// Test that a zero-extension is split by a store.
			{
				name:    "StoreZeroExtended",
				into:    memLeaf{},
				version: 1,
				v:       cast(i32, runtime(i16, 2)),
				off:     0,
				result: memLeaf{
					metaHigh:      0b1111,
					metaLow:       0b0011 << 2,
					specialStarts: 0b0001,
					pending:       0b1111,
					version:       1,
					specials:      []value{runtime(i16, 2)},
				},
			},

			// Test that a store of concatenated values is broken down.
			{
				name:    "StoreCat",
				into:    memLeaf{},
				version: 1,
				v:       cat([]value{smallIntValue(i8, 1), undefValue(i16), runtime(i8, 2)}),
				off:     2,
				result: memLeaf{
					metaHigh:      0b1001 << 2,
					metaLow:       0b0111 << 2,
					specialStarts: 0b1000 << 2,
					pending:       0b1111 << 2,
					version:       1,
					data: [64]byte{
						0x2: 1,
					},
					specials: []value{runtime(i8, 2)},
				},
			},
			{
				// This used to not get broken down due to the merge of the zero-padding.
				name:    "StoreCatZeroExtendRegression",
				into:    memLeaf{},
				version: 1,
				v:       cat([]value{cast(i8, runtime(iType(1), 2)), smallIntValue(i8, 5)}),
				off:     3,
				result: memLeaf{
					metaHigh:      0b11 << 3,
					metaLow:       0b10 << 3,
					specialStarts: 0b01 << 3,
					pending:       0b11 << 3,
					version:       1,
					data: [64]byte{
						0x4: 5,
					},
					specials: []value{runtime(iType(1), 2)},
				},
			},

			// Test stores of aggregates.
			{
				name:    "StoreArray",
				into:    memLeaf{},
				version: 1,
				v:       arrayValue(i32, smallIntValue(i32, 1), smallIntValue(i32, 2)),
				off:     4,
				result: memLeaf{
					metaHigh: 0b11111111 << 4,
					metaLow:  0b11111111 << 4,
					pending:  0b11111111 << 4,
					version:  1,
					data: [64]byte{
						0x4: 1, 0x8: 2,
					},
				},
			},
			{
				name:    "StoreStruct",
				into:    memLeaf{},
				version: 1,
				v:       structValue(&structType{fields: []structField{{i8, 0}, {i16, 2}}, size: 4}, smallIntValue(i8, 5), smallIntValue(i16, 7)),
				off:     6,
				result: memLeaf{
					metaHigh: 0b1101 << 6,
					metaLow:  0b1111 << 6,
					pending:  0b1111 << 6,
					version:  1,
					data: [64]byte{
						0x6: 5, 0x8: 7,
					},
				},
			},
			{
				name:    "StoreEmpty",
				into:    memLeaf{},
				version: 1,
				v:       structValue(&structType{}),
				off:     63,
				result:  memLeaf{},
			},

			// Test clobbering behavior.
			{
				name: "ClobberBytes",
				into: memLeaf{
					metaHigh: 0b1111,
					metaLow:  0b1111,
					version:  1,
					data: [64]byte{
						0x0: 1, 0x1: 2, 0x2: 3, 0x3: 4,
					},
				},
				version: 1,
				v:       smallIntValue(i16, (6<<8)|5),
				off:     1,
				result: memLeaf{
					metaHigh: 0b1111,
					metaLow:  0b1111,
					pending:  0b0110,
					version:  1,
					data: [64]byte{
						0x0: 1, 0x1: 5, 0x2: 6, 0x3: 4,
					},
				},
			},
			{
				name: "ClobberSplitSpecial",
				into: memLeaf{
					metaHigh:      0b1111,
					specialStarts: 0b0001,
					version:       1,
					specials:      []value{runtime(i32, 1)},
				},
				version: 1,
				v:       smallIntValue(i16, 2|(3<<8)),
				off:     1,
				result: memLeaf{
					metaHigh:      0b1111,
					metaLow:       0b0110,
					specialStarts: 0b1001,
					pending:       0b0110,
					version:       1,
					data: [64]byte{
						0x1: 2, 0x2: 3,
					},
					specials: []value{slice(runtime(i32, 1), 0, 8), slice(runtime(i32, 1), 24, 8)},
				},
			},
		}

		for _, c := range cases {
			c := c
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				into := c.into
				t.Log(c.v)
				res, err := into.store(c.v, c.off, c.version)
				if err != nil {
					t.Errorf("store failed: %s", err)
					return
				}
				if !reflect.DeepEqual(res, &c.result) {
					r := res.(*memLeaf)
					t.Error("bad result:")
					t.Errorf("\tversion:        %d", r.version)
					t.Errorf("\tmeta:           %064b, %064b", r.metaHigh, r.metaLow)
					t.Errorf("\tspecial starts: %064b", r.specialStarts)
					t.Errorf("\tpending:        %064b", r.pending)
					t.Errorf("\tnoinit:         %064b", r.noInit)
					t.Errorf("\tdata:           %x", r.data)
					t.Errorf("\tspecials:       %s", r.specials)
				}
			})
		}
	})

	// Test loads from a leaf.
	t.Run("Load", func(t *testing.T) {
		t.Parallel()

		paddedStruct := &structType{fields: []structField{{i8, 0}, {i16, 2}, {i8, 4}}, size: 6}
		cases := []struct {
			name string
			from memLeaf
			ty   typ
			off  uint64
			v    value
		}{
			// Load a constant integer.
			{
				name: "ConstInt",
				from: memLeaf{
					metaHigh: 0b11111111 << 1,
					metaLow:  0b11111111 << 1,
					version:  1,
					data: [64]byte{
						0x1: 1, 0x2: 2, 0x3: 3, 0x4: 4,
						0x5: 5, 0x6: 6, 0x7: 7, 0x8: 8,
					},
				},
				ty:  i64,
				off: 1,
				v:   smallIntValue(i64, 0x0807060504030201),
			},

			// Load a pointer.
			{
				name: "ConstPtr",
				from: memLeaf{
					metaHigh:      0b11111111 << 1,
					metaLow:       0b00000000 << 1,
					specialStarts: 0b00000001 << 1,
					version:       1,
					specials:      []value{x.ptr(0)},
				},
				ty:  ptrTy,
				off: 1,
				v:   x.ptr(0),
			},

			// Load undef.
			{
				name: "Undef",
				from: memLeaf{
					metaLow: 0b1111 << 1,
					version: 1,
				},
				ty:  i32,
				off: 1,
				v:   undefValue(i32),
			},

			// Load a giant mess.
			{
				name: "Messy",
				from: memLeaf{
					metaHigh:      0b11111100,
					metaLow:       0b00001111,
					specialStarts: 0b00010000,
					version:       1,
					specials:      []value{x.ptr(0)},
					data: [64]byte{
						0x2: 2, 0x3: 3,
					},
				},
				ty:  i64,
				off: 0,
				v:   cat([]value{undefValue(i16), smallIntValue(i16, 0x0302), cast(i32, x.ptr(0))}),
			},

			// Test a load which is unknown.
			{
				name: "Unknown",
				from: memLeaf{},
				ty:   i64,
				off:  7,
			},

			// Test loading composites.
			{
				name: "Array",
				from: memLeaf{
					metaHigh:      0b110,
					metaLow:       0b101,
					specialStarts: 0b010,
					version:       1,
					data: [64]byte{
						0x2: 2,
					},
					specials: []value{runtime(i8, 2)},
				},
				ty:  array(i8, 3),
				off: 0,
				v:   arrayValue(i8, undefValue(i8), runtime(i8, 2), smallIntValue(i8, 2)),
			},
			{
				name: "Struct",
				from: memLeaf{
					// The paddding bytes are flagged as unknown to verify that this does not try to load them.
					metaHigh:      0b011101,
					specialStarts: 0b010101,
					version:       1,
					specials:      []value{runtime(i8, 1), runtime(i16, 2), runtime(i8, 3)},
				},
				ty:  paddedStruct,
				off: 0,
				v:   structValue(paddedStruct, runtime(i8, 1), runtime(i16, 2), runtime(i8, 3)),
			},
		}

		for _, c := range cases {
			c := c
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				switch v, err := c.from.load(c.ty, c.off); err {
				case nil:
					if c.v == (value{}) {
						t.Errorf("expected errRuntime but got a value %s", v.String())
					} else if !reflect.DeepEqual(v, c.v) {
						t.Errorf("expected %s but got %s", c.v.String(), v.String())
					}

				case errRuntime:
					if c.v == (value{}) {
						return
					}
					fallthrough
				default:
					t.Errorf("load failed: %s", err.Error())
				}
			})
		}
	})
}

func TestInitialize(t *testing.T) {
	t.Parallel()

	ptrTy := pointer(defaultAddrSpace, i32)
	x := memObj{
		ptrTy: ptrTy,
		name:  "x",
	}

	cases := []struct {
		name string
		v    value
	}{
		{"Int", smallIntValue(i32, 5)},
		{"Complex", structValue(&structType{
			fields: []structField{
				// This is not a remotely realistic layout.
				// However, it triggers almost every edge case.
				{i8, 0},
				{array(ptrTy, 1), 63},
			},
			size: 120322,
		}, smallIntValue(i8, 1), arrayValue(ptrTy, x.ptr(0)))},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mem, err := makeInitializedMem(c.v)
			if err != nil {
				t.Errorf("failed to initialize: %s", err.Error())
				return
			}
			v, err := mem.load(c.v.typ(), 0)
			if err != nil {
				t.Errorf("failed to reload: %v", err.Error())
				return
			}
			if !reflect.DeepEqual(c.v, v) {
				t.Errorf("initialized with %s but got %s", c.v.String(), v.String())
			}
		})
	}
}
