package compiler

import (
	"strings"
)

var basicTypes = map[string]uint64{
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
		if strings.HasPrefix(t.name, "type:basic:") {
			// Basic types have a typecode with the lowest bit set to 0.
			num, ok := basicTypes[t.name[len("type:basic:"):]]
			if !ok {
				panic("invalid basic type: " + t.name)
			}
			t.num = num<<1 | 0
		} else {
			// Fallback types have a typecode with the lowest bit set to 1.
			t.num = uint64(fallbackIndex<<1 | 1)
			fallbackIndex++
		}
	}
}
