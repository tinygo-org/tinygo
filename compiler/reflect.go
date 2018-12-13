package compiler

import (
	"math/big"
	"strings"
)

var basicTypes = map[string]int64{
	"bool":       1,
	"int":        2,
	"int8":       3,
	"int16":      4,
	"int32":      5,
	"int64":      6,
	"uint":       7,
	"uint8":      8,
	"uint16":     9,
	"uint32":     10,
	"uint64":     11,
	"uintptr":    12,
	"float32":    13,
	"float64":    14,
	"complex64":  15,
	"complex128": 16,
	"string":     17,
	"unsafeptr":  18,
}

func (c *Compiler) assignTypeCodes(typeSlice typeInfoSlice) {
	fn := c.mod.NamedFunction("reflect.ValueOf")
	if fn.IsNil() {
		// reflect.ValueOf is never used, so we can use the most efficient
		// encoding possible.
		for i, t := range typeSlice {
			t.num = uint64(i + 1)
		}
		return
	}

	// Assign typecodes the way the reflect package expects.
	fallbackIndex := 1
	for _, t := range typeSlice {
		if t.name[:5] != "type:" {
			panic("expected type name to start with 'type:'")
		}
		num := c.getTypeCodeNum(t.name[5:])
		if num == nil {
			// Fallback/unsupported types have a typecode with the lowest bits
			// set to 11.
			t.num = uint64(fallbackIndex<<2 | 3)
			fallbackIndex++
			continue
		}
		if num.BitLen() > c.uintptrType.IntTypeWidth() || !num.IsUint64() {
			// TODO: support this in some way, using a side table for example.
			// That's less efficient but better than not working at all.
			// Particularly important on systems with 16-bit pointers (e.g.
			// AVR).
			panic("compiler: could not store type code number inside interface type code")
		}
		t.num = num.Uint64()
	}
}

// getTypeCodeNum returns the typecode for a given type as expected by the
// reflect package. Also see getTypeCodeName, which serializes types to a string
// based on a types.Type value for this function.
func (c *Compiler) getTypeCodeNum(name string) *big.Int {
	if strings.HasPrefix(name, "basic:") {
		// Basic types have a typecode with the lowest bits set to 00.
		num, ok := basicTypes[name[len("basic:"):]]
		if !ok {
			panic("invalid basic type: " + name)
		}
		return big.NewInt(num<<2 | 0)
	} else if strings.HasPrefix(name, "slice:") {
		// Slices have a typecode with the lowest bits set to 01.
		num := c.getTypeCodeNum(name[len("slice:"):])
		num.Lsh(num, 2).Or(num, big.NewInt(1))
		return num
	} else {
		return nil
	}
}
