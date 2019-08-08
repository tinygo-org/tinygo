package compiler

// This file has some compiler support for run-time reflection using the reflect
// package. In particular, it encodes type information in type codes in such a
// way that the reflect package can decode the type from this information.
// Where needed, it also adds some side tables for looking up more information
// about a type, when that information cannot be stored directly in the type
// code.
//
// Go has 26 different type kinds.
//
// Type kinds are subdivided in basic types (see the list of basicTypes below)
// that are mostly numeric literals and non-basic (or "complex") types that are
// more difficult to encode. These non-basic types come in two forms:
//   * Prefix types (pointer, slice, interface, channel): these just add
//     something to an existing type. For example, a pointer like *int just adds
//     the fact that it's a pointer to an existing type (int).
//     These are encoded efficiently by adding a prefix to a type code.
//   * Types with multiple fields (struct, array, func, map). All of these have
//     multiple fields contained within. Most obviously structs can contain many
//     types as fields. Also arrays contain not just the element type but also
//     the length parameter which can be any arbitrary number and thus may not
//     fit in a type code.
//     These types are encoded using side tables.
//
// This distinction is also important for how named types are encoded. At the
// moment, named basic type just get a unique number assigned while named
// non-basic types have their underlying type stored in a sidetable.

import (
	"encoding/binary"
	"go/ast"
	"math/big"
	"strings"

	"tinygo.org/x/go-llvm"
)

// A list of basic types and their numbers. This list should be kept in sync
// with the list of Kind constants of type.go in the runtime package.
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

// typeCodeAssignmentState keeps some global state around for type code
// assignments, used to assign one unique type code to each Go type.
type typeCodeAssignmentState struct {
	// An integer that's incremented each time it's used to give unique IDs to
	// type codes that are not yet fully supported otherwise by the reflect
	// package (or are simply unused in the compiled program).
	fallbackIndex int

	// This is the length of an uintptr. Only used occasionally to know whether
	// a given number can be encoded as a varint.
	uintptrLen int

	// Map of named types to their type code. It is important that named types
	// get unique IDs for each type.
	namedBasicTypes    map[string]int
	namedNonBasicTypes map[string]int

	// Map of struct types to their type code.
	structTypes               map[string]int
	structTypesSidetable      []byte
	needsStructNamesSidetable bool

	// Map of struct names and tags to their name string.
	structNames               map[string]int
	structNamesSidetable      []byte
	needsStructTypesSidetable bool

	// This byte array is stored in reflect.namedNonBasicTypesSidetable and is
	// used at runtime to get details about a named non-basic type.
	// Entries are varints (see makeVarint below and readVarint in
	// reflect/sidetables.go for the encoding): one varint per entry. The
	// integers in namedNonBasicTypes are indices into this array. Because these
	// are varints, most type codes are really small (just one byte).
	//
	// Note that this byte buffer is not created when it is not needed
	// (reflect.namedNonBasicTypesSidetable has no uses), see
	// needsNamedTypesSidetable.
	namedNonBasicTypesSidetable []byte

	// This indicates whether namedNonBasicTypesSidetable needs to be created at
	// all. If it is false, namedNonBasicTypesSidetable will contain simple
	// monotonically increasing numbers.
	needsNamedNonBasicTypesSidetable bool
}

// assignTypeCodes is used to assign a type code to each type in the program
// that is ever stored in an interface. It tries to use the smallest possible
// numbers to make the code that works with interfaces as small as possible.
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
	state := typeCodeAssignmentState{
		fallbackIndex:                    1,
		uintptrLen:                       c.uintptrType.IntTypeWidth(),
		namedBasicTypes:                  make(map[string]int),
		namedNonBasicTypes:               make(map[string]int),
		structTypes:                      make(map[string]int),
		structNames:                      make(map[string]int),
		needsNamedNonBasicTypesSidetable: len(getUses(c.mod.NamedGlobal("reflect.namedNonBasicTypesSidetable"))) != 0,
		needsStructTypesSidetable:        len(getUses(c.mod.NamedGlobal("reflect.structTypesSidetable"))) != 0,
		needsStructNamesSidetable:        len(getUses(c.mod.NamedGlobal("reflect.structNamesSidetable"))) != 0,
	}
	for _, t := range typeSlice {
		num := state.getTypeCodeNum(t.typecode)
		if num.BitLen() > c.uintptrType.IntTypeWidth() || !num.IsUint64() {
			// TODO: support this in some way, using a side table for example.
			// That's less efficient but better than not working at all.
			// Particularly important on systems with 16-bit pointers (e.g.
			// AVR).
			panic("compiler: could not store type code number inside interface type code")
		}
		t.num = num.Uint64()
	}

	// Only create this sidetable when it is necessary.
	if state.needsNamedNonBasicTypesSidetable {
		global := c.replaceGlobalByteWithArray("reflect.namedNonBasicTypesSidetable", state.namedNonBasicTypesSidetable)
		global.SetLinkage(llvm.InternalLinkage)
		global.SetUnnamedAddr(true)
	}
	if state.needsStructTypesSidetable {
		global := c.replaceGlobalByteWithArray("reflect.structTypesSidetable", state.structTypesSidetable)
		global.SetLinkage(llvm.InternalLinkage)
		global.SetUnnamedAddr(true)
	}
	if state.needsStructNamesSidetable {
		global := c.replaceGlobalByteWithArray("reflect.structNamesSidetable", state.structNamesSidetable)
		global.SetLinkage(llvm.InternalLinkage)
		global.SetUnnamedAddr(true)
	}
}

// getTypeCodeNum returns the typecode for a given type as expected by the
// reflect package. Also see getTypeCodeName, which serializes types to a string
// based on a types.Type value for this function.
func (state *typeCodeAssignmentState) getTypeCodeNum(typecode llvm.Value) *big.Int {
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
			num |= int64(state.getBasicNamedTypeNum(name)) << 5
		}
		return big.NewInt(num << 1)
	} else {
		// Non-baisc types use the following bit pattern:
		//    ...nxxx1
		// where xxx indicates the non-basic type. The upper bits contain
		// whatever the type contains. Types that wrap a single other type
		// (channel, interface, pointer, slice) just contain the bits of the
		// wrapped type. Other types (like struct) need more fields and thus
		// cannot be encoded as a simple prefix.
		var num *big.Int
		var classNumber int64
		switch class {
		case "chan":
			sub := llvm.ConstExtractValue(typecode.Initializer(), []uint32{0})
			num = state.getTypeCodeNum(sub)
			classNumber = 0
		case "interface":
			num = big.NewInt(int64(state.fallbackIndex))
			state.fallbackIndex++
			classNumber = 1
		case "pointer":
			sub := llvm.ConstExtractValue(typecode.Initializer(), []uint32{0})
			num = state.getTypeCodeNum(sub)
			classNumber = 2
		case "slice":
			sub := llvm.ConstExtractValue(typecode.Initializer(), []uint32{0})
			num = state.getTypeCodeNum(sub)
			classNumber = 3
		case "array":
			num = big.NewInt(int64(state.fallbackIndex))
			state.fallbackIndex++
			classNumber = 4
		case "func":
			num = big.NewInt(int64(state.fallbackIndex))
			state.fallbackIndex++
			classNumber = 5
		case "map":
			num = big.NewInt(int64(state.fallbackIndex))
			state.fallbackIndex++
			classNumber = 6
		case "struct":
			num = big.NewInt(int64(state.getStructTypeNum(typecode)))
			classNumber = 7
		default:
			panic("unknown type kind: " + class)
		}
		if name == "" {
			num.Lsh(num, 5).Or(num, big.NewInt((classNumber<<1)+1))
		} else {
			num = big.NewInt(int64(state.getNonBasicNamedTypeNum(name, num))<<1 | 1)
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

// getBasicNamedTypeNum returns an appropriate (unique) number for the given
// named type. If the name already has a number that number is returned, else a
// new number is returned. The number is always non-zero.
func (state *typeCodeAssignmentState) getBasicNamedTypeNum(name string) int {
	if num, ok := state.namedBasicTypes[name]; ok {
		return num
	}
	num := len(state.namedBasicTypes) + 1
	state.namedBasicTypes[name] = num
	return num
}

// getNonBasicNamedTypeNum returns a number unique for this named type. It tries
// to return the smallest number possible to make encoding of this type code
// easier.
func (state *typeCodeAssignmentState) getNonBasicNamedTypeNum(name string, value *big.Int) int {
	if num, ok := state.namedNonBasicTypes[name]; ok {
		return num
	}
	if !state.needsNamedNonBasicTypesSidetable {
		// Use simple small integers in this case, to make these numbers
		// smaller.
		num := len(state.namedNonBasicTypes) + 1
		state.namedNonBasicTypes[name] = num
		return num
	}
	num := len(state.namedNonBasicTypesSidetable)
	if value.BitLen() > state.uintptrLen || !value.IsUint64() {
		panic("cannot store value in sidetable")
	}
	state.namedNonBasicTypesSidetable = append(state.namedNonBasicTypesSidetable, makeVarint(value.Uint64())...)
	state.namedNonBasicTypes[name] = num
	return num
}

// getStructTypeNum returns the struct type number, which is an index into
// reflect.structTypesSidetable or an unique number for every struct if this
// sidetable is not needed in the to-be-compiled program.
func (state *typeCodeAssignmentState) getStructTypeNum(typecode llvm.Value) int {
	name := typecode.Name()
	if num, ok := state.structTypes[name]; ok {
		// This struct already has an assigned type code.
		return num
	}

	if !state.needsStructTypesSidetable {
		// We don't need struct sidetables, so we can just assign monotonically
		// increasing numbers to each struct type.
		num := len(state.structTypes)
		state.structTypes[name] = num
		return num
	}

	// Get the fields this struct type contains.
	// The struct number will be the start index of
	structTypeGlobal := llvm.ConstExtractValue(typecode.Initializer(), []uint32{0}).Operand(0).Initializer()
	numFields := structTypeGlobal.Type().ArrayLength()

	// The first data that is stored in the struct sidetable is the number of
	// fields this struct contains. This is usually just a single byte because
	// most structs don't contain that many fields, but make it a varint just
	// to be sure.
	buf := makeVarint(uint64(numFields))

	// Iterate over every field in the struct.
	// Every field is stored sequentially in the struct sidetable. Fields can
	// be retrieved from this list of fields at runtime by iterating over all
	// of them until the right field has been found.
	// Perhaps adding some index would speed things up, but it would also make
	// the sidetable bigger.
	for i := 0; i < numFields; i++ {
		// Collect some information about this field.
		field := llvm.ConstExtractValue(structTypeGlobal, []uint32{uint32(i)})

		nameGlobal := llvm.ConstExtractValue(field, []uint32{1})
		if nameGlobal == llvm.ConstPointerNull(nameGlobal.Type()) {
			panic("compiler: no name for this struct field")
		}
		fieldNameBytes := getGlobalBytes(nameGlobal.Operand(0))
		fieldNameNumber := state.getStructNameNumber(fieldNameBytes)

		// See whether this struct field has an associated tag, and if so,
		// store that tag in the tags sidetable.
		tagGlobal := llvm.ConstExtractValue(field, []uint32{2})
		hasTag := false
		tagNumber := 0
		if tagGlobal != llvm.ConstPointerNull(tagGlobal.Type()) {
			hasTag = true
			tagBytes := getGlobalBytes(tagGlobal.Operand(0))
			tagNumber = state.getStructNameNumber(tagBytes)
		}

		// The 'embedded' or 'anonymous' flag for this field.
		embedded := llvm.ConstExtractValue(field, []uint32{3}).ZExtValue() != 0

		// The first byte in the struct types sidetable is a flags byte with
		// two bits in it.
		flagsByte := byte(0)
		if embedded {
			flagsByte |= 1
		}
		if hasTag {
			flagsByte |= 2
		}
		if ast.IsExported(string(fieldNameBytes)) {
			flagsByte |= 4
		}
		buf = append(buf, flagsByte)

		// Get the type number and add it to the buffer.
		// All fields have a type, so include it directly here.
		typeNum := state.getTypeCodeNum(llvm.ConstExtractValue(field, []uint32{0}))
		if typeNum.BitLen() > state.uintptrLen || !typeNum.IsUint64() {
			// TODO: make this a regular error
			panic("struct field has a type code that is too big")
		}
		buf = append(buf, makeVarint(typeNum.Uint64())...)

		// Add the name.
		buf = append(buf, makeVarint(uint64(fieldNameNumber))...)

		// Add the tag, if there is one.
		if hasTag {
			buf = append(buf, makeVarint(uint64(tagNumber))...)
		}
	}

	num := len(state.structTypesSidetable)
	state.structTypes[name] = num
	state.structTypesSidetable = append(state.structTypesSidetable, buf...)
	return num
}

// getStructNameNumber stores this string (name or tag) onto the struct names
// sidetable. The format is a varint of the length of the struct, followed by
// the raw bytes of the name. Multiple identical strings are stored under the
// same name for space efficiency.
func (state *typeCodeAssignmentState) getStructNameNumber(nameBytes []byte) int {
	name := string(nameBytes)
	if n, ok := state.structNames[name]; ok {
		// This name was used before, re-use it now (for space efficiency).
		return n
	}
	// This name is not yet in the names sidetable. Add it now.
	n := len(state.structNamesSidetable)
	state.structNames[name] = n
	state.structNamesSidetable = append(state.structNamesSidetable, makeVarint(uint64(len(nameBytes)))...)
	state.structNamesSidetable = append(state.structNamesSidetable, nameBytes...)
	return n
}

// makeVarint is a small helper function that returns the bytes of the number in
// varint encoding.
func makeVarint(n uint64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	return buf[:binary.PutUvarint(buf, n)]
}
