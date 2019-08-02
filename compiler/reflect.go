package compiler

import (
	"math/big"
	"strings"

	"tinygo.org/x/go-llvm"
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
	namedTypes := make(map[string]int)
	for _, t := range typeSlice {
		num := c.getTypeCodeNum(t.typecode, &fallbackIndex, namedTypes)
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
func (c *Compiler) getTypeCodeNum(typecode llvm.Value, fallbackIndex *int, namedTypes map[string]int) *big.Int {
	// Note: see src/reflect/type.go for bit allocations.
	class, value := getClassAndValueFromTypeCode(typecode)
	name := ""
	if class == "named" {
		name = value
		typecode = llvm.ConstExtractValue(typecode.Initializer(), []uint32{0})
		class, value = getClassAndValueFromTypeCode(typecode)
	}
	if class == "basic" {
		// Basic types follow the following bit pattern:
		//    ...xxxxx0
		// where xxxxx is allocated for the 18 possible basic types and all the
		// upper bits are used to indicate the named type.
		num, ok := basicTypes[value]
		if !ok {
			panic("invalid basic type: " + value)
		}
		if name != "" {
			// This type is named, set the upper bits to the name ID.
			num |= int64(getNamedTypeNum(namedTypes, name)) << 5
		}
		return big.NewInt(num << 1)
	} else {
		// Complex types use the following bit pattern:
		//    ...nxxx1
		// where xxx indicates the complex type (any non-basic type). The upper
		// bits contain whatever the type contains. Types that wrap a single
		// other type (channel, interface, pointer, slice) just contain the bits
		// of the wrapped type. Other types (like struct) have a different
		// method of encoding the contents of the type.
		var num *big.Int
		var classNumber int64
		switch class {
		case "chan":
			sub := llvm.ConstExtractValue(typecode.Initializer(), []uint32{0})
			num = c.getTypeCodeNum(sub, fallbackIndex, namedTypes)
			classNumber = 0
		case "interface":
			num = big.NewInt(int64(*fallbackIndex))
			*fallbackIndex++
			classNumber = 1
		case "pointer":
			sub := llvm.ConstExtractValue(typecode.Initializer(), []uint32{0})
			num = c.getTypeCodeNum(sub, fallbackIndex, namedTypes)
			classNumber = 2
		case "slice":
			sub := llvm.ConstExtractValue(typecode.Initializer(), []uint32{0})
			num = c.getTypeCodeNum(sub, fallbackIndex, namedTypes)
			classNumber = 3
		case "array":
			num = big.NewInt(int64(*fallbackIndex))
			*fallbackIndex++
			classNumber = 4
		case "func":
			num = big.NewInt(int64(*fallbackIndex))
			*fallbackIndex++
			classNumber = 5
		case "map":
			num = big.NewInt(int64(*fallbackIndex))
			*fallbackIndex++
			classNumber = 6
		case "struct":
			num = big.NewInt(int64(*fallbackIndex))
			*fallbackIndex++
			classNumber = 7
		default:
			panic("unknown type kind: " + class)
		}
		if name == "" {
			num.Lsh(num, 5).Or(num, big.NewInt((classNumber<<1)+1))
		} else {
			// TODO: store num in a sidetable
			num = big.NewInt(int64(getNamedTypeNum(namedTypes, name))<<1 | 1)
			num.Lsh(num, 4).Or(num, big.NewInt((classNumber<<1)+1))
		}
		return num
	}
}

// getClassAndValueFromTypeCode takes a typecode (a llvm.Value of type
// runtime.typecodeID), looks at the name, and extracts the typecode class and
// value from it. For example, for a typecode with the following name:
//     reflect/types.type:pointer:named:reflect.ValueError
// It extracts:
//     class = "pointer"
//     value = "named:reflect.ValueError"
func getClassAndValueFromTypeCode(typecode llvm.Value) (class, value string) {
	typecodeName := typecode.Name()
	const prefix = "reflect/types.type:"
	if !strings.HasPrefix(typecodeName, prefix) {
		panic("unexpected typecode name: " + typecodeName)
	}
	id := typecodeName[len(prefix):]
	class = id[:strings.IndexByte(id, ':')]
	value = id[len(class)+1:]
	return
}

// getNamedTypeNum returns an appropriate (unique) number for the given named
// type. If the name already has a number that number is returned, else a new
// number is returned. The number is always non-zero.
func getNamedTypeNum(namedTypes map[string]int, name string) int {
	if num, ok := namedTypes[name]; ok {
		return num
	} else {
		num = len(namedTypes) + 1
		namedTypes[name] = num
		return num
	}
}
