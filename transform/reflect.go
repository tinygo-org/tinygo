package transform

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
	"sort"
	"strings"

	"tinygo.org/x/go-llvm"
)

// A list of basic types and their numbers. This list should be kept in sync
// with the list of Kind constants of type.go in the reflect package.
var basicTypes = map[string]int64{
	"bool":           1,
	"int":            2,
	"int8":           3,
	"int16":          4,
	"int32":          5,
	"int64":          6,
	"uint":           7,
	"uint8":          8,
	"uint16":         9,
	"uint32":         10,
	"uint64":         11,
	"uintptr":        12,
	"float32":        13,
	"float64":        14,
	"complex64":      15,
	"complex128":     16,
	"string":         17,
	"unsafe.Pointer": 18,
}

// A list of non-basic types. Adding 19 to this number will give the Kind as
// used in src/reflect/types.go, and it must be kept in sync with that list.
var nonBasicTypes = map[string]int64{
	"chan":      0,
	"interface": 1,
	"pointer":   2,
	"slice":     3,
	"array":     4,
	"func":      5,
	"map":       6,
	"struct":    7,
}

// typeCodeAssignmentState keeps some global state around for type code
// assignments, used to assign one unique type code to each Go type.
type typeCodeAssignmentState struct {
	// Builder used purely for constant operations (because LLVM 15 removed many
	// llvm.Const* functions).
	builder llvm.Builder

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

	// Map of array types to their type code.
	arrayTypes               map[string]int
	arrayTypesSidetable      []byte
	needsArrayTypesSidetable bool

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
	namedNonBasicTypesSidetable []uint64

	// This indicates whether namedNonBasicTypesSidetable needs to be created at
	// all. If it is false, namedNonBasicTypesSidetable will contain simple
	// monotonically increasing numbers.
	needsNamedNonBasicTypesSidetable bool
}

// LowerReflect is used to assign a type code to each type in the program
// that is ever stored in an interface. It tries to use the smallest possible
// numbers to make the code that works with interfaces as small as possible.
func LowerReflect(mod llvm.Module) {
	// if reflect were not used, we could skip generating the sidetable
	// this does not help in practice, and is difficult to do correctly

	// Obtain slice of all types in the program.
	type typeInfo struct {
		typecode llvm.Value
		name     string
		numUses  int
	}
	var types []*typeInfo
	for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if strings.HasPrefix(global.Name(), "reflect/types.type:") {
			types = append(types, &typeInfo{
				typecode: global,
				name:     global.Name(),
				numUses:  len(getUses(global)),
			})
		}
	}

	// Sort the slice in a way that often used types are assigned a type code
	// first.
	sort.Slice(types, func(i, j int) bool {
		if types[i].numUses != types[j].numUses {
			return types[i].numUses < types[j].numUses
		}
		// It would make more sense to compare the name in the other direction,
		// but for some reason that increases binary size. Could be a fluke, but
		// could also have some good reason (and possibly hint at a small
		// optimization).
		return types[i].name > types[j].name
	})

	// Assign typecodes the way the reflect package expects.
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	uintptrType := mod.Context().IntType(targetData.PointerSize() * 8)
	state := typeCodeAssignmentState{
		builder:                          mod.Context().NewBuilder(),
		fallbackIndex:                    1,
		uintptrLen:                       targetData.PointerSize() * 8,
		namedBasicTypes:                  make(map[string]int),
		namedNonBasicTypes:               make(map[string]int),
		arrayTypes:                       make(map[string]int),
		structTypes:                      make(map[string]int),
		structNames:                      make(map[string]int),
		needsNamedNonBasicTypesSidetable: len(getUses(mod.NamedGlobal("reflect.namedNonBasicTypesSidetable"))) != 0,
		needsStructTypesSidetable:        len(getUses(mod.NamedGlobal("reflect.structTypesSidetable"))) != 0,
		needsStructNamesSidetable:        len(getUses(mod.NamedGlobal("reflect.structNamesSidetable"))) != 0,
		needsArrayTypesSidetable:         len(getUses(mod.NamedGlobal("reflect.arrayTypesSidetable"))) != 0,
	}
	defer state.builder.Dispose()
	for _, t := range types {
		num := state.getTypeCodeNum(t.typecode)
		if num.BitLen() > state.uintptrLen || !num.IsUint64() {
			// TODO: support this in some way, using a side table for example.
			// That's less efficient but better than not working at all.
			// Particularly important on systems with 16-bit pointers (e.g.
			// AVR).
			panic("compiler: could not store type code number inside interface type code")
		}

		// Replace each use of the type code global with the constant type code.
		for _, use := range getUses(t.typecode) {
			if use.IsAConstantExpr().IsNil() {
				continue
			}
			typecode := llvm.ConstInt(uintptrType, num.Uint64(), false)
			switch use.Opcode() {
			case llvm.PtrToInt:
				// Already of the correct type.
			case llvm.BitCast:
				// Could happen when stored in an interface (which is of type
				// i8*).
				typecode = llvm.ConstIntToPtr(typecode, use.Type())
			default:
				panic("unexpected constant expression")
			}
			use.ReplaceAllUsesWith(typecode)
		}
	}

	// Only create this sidetable when it is necessary.
	if state.needsNamedNonBasicTypesSidetable {
		global := replaceGlobalIntWithArray(mod, "reflect.namedNonBasicTypesSidetable", state.namedNonBasicTypesSidetable)
		global.SetLinkage(llvm.InternalLinkage)
		global.SetUnnamedAddr(true)
		global.SetGlobalConstant(true)
	}
	if state.needsArrayTypesSidetable {
		global := replaceGlobalIntWithArray(mod, "reflect.arrayTypesSidetable", state.arrayTypesSidetable)
		global.SetLinkage(llvm.InternalLinkage)
		global.SetUnnamedAddr(true)
		global.SetGlobalConstant(true)
	}
	if state.needsStructTypesSidetable {
		global := replaceGlobalIntWithArray(mod, "reflect.structTypesSidetable", state.structTypesSidetable)
		global.SetLinkage(llvm.InternalLinkage)
		global.SetUnnamedAddr(true)
		global.SetGlobalConstant(true)
	}
	if state.needsStructNamesSidetable {
		global := replaceGlobalIntWithArray(mod, "reflect.structNamesSidetable", state.structNamesSidetable)
		global.SetLinkage(llvm.InternalLinkage)
		global.SetUnnamedAddr(true)
		global.SetGlobalConstant(true)
	}

	// Remove most objects created for interface and reflect lowering.
	// They would normally be removed anyway in later passes, but not always.
	// It also cleans up the IR for testing.
	for _, typ := range types {
		initializer := typ.typecode.Initializer()
		references := state.builder.CreateExtractValue(initializer, 0, "")
		typ.typecode.SetInitializer(llvm.ConstNull(initializer.Type()))
		if strings.HasPrefix(typ.name, "reflect/types.type:struct:") {
			// Structs have a 'references' field that is not a typecode but
			// a pointer to a runtime.structField array and therefore a
			// bitcast. This global should be erased separately, otherwise
			// typecode objects cannot be erased.
			structFields := references.Operand(0)
			structFields.EraseFromParentAsGlobal()
		}
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
		typecode = state.builder.CreateExtractValue(typecode.Initializer(), 0, "")
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
		var classNumber int64
		if n, ok := nonBasicTypes[class]; ok {
			classNumber = n
		} else {
			panic("unknown type kind: " + class)
		}
		var num *big.Int
		lowBits := (classNumber << 1) + 1 // the 5 low bits of the typecode
		if name == "" {
			num = state.getNonBasicTypeCode(class, typecode)
		} else {
			// We must return a named type here. But first check whether it
			// has already been defined.
			if index, ok := state.namedNonBasicTypes[name]; ok {
				num := big.NewInt(int64(index))
				num.Lsh(num, 5).Or(num, big.NewInt((classNumber<<1)+1+(1<<4)))
				return num
			}
			lowBits |= 1 << 4 // set the 'n' bit (see above)
			if !state.needsNamedNonBasicTypesSidetable {
				// Use simple small integers in this case, to make these numbers
				// smaller.
				index := len(state.namedNonBasicTypes) + 1
				state.namedNonBasicTypes[name] = index
				num = big.NewInt(int64(index))
			} else {
				// We need to store full type information.
				// First allocate a number in the named non-basic type
				// sidetable.
				index := len(state.namedNonBasicTypesSidetable)
				state.namedNonBasicTypesSidetable = append(state.namedNonBasicTypesSidetable, 0)
				state.namedNonBasicTypes[name] = index
				// Get the typecode of the underlying type (which could be the
				// element type in the case of pointers, for example).
				num = state.getNonBasicTypeCode(class, typecode)
				if num.BitLen() > state.uintptrLen || !num.IsUint64() {
					panic("cannot store value in sidetable")
				}
				// Now update the side table with the number we just
				// determined. We need this multi-step approach to avoid stack
				// overflow due to adding types recursively in the case of
				// linked lists (a pointer which points to a struct that
				// contains that same pointer).
				state.namedNonBasicTypesSidetable[index] = num.Uint64()
				num = big.NewInt(int64(index))
			}
		}
		// Concatenate the 'num' and 'lowBits' bitstrings.
		num.Lsh(num, 5).Or(num, big.NewInt(lowBits))
		return num
	}
}

// getNonBasicTypeCode is used by getTypeCodeNum. It returns the upper bits of
// the type code used there in the type code.
func (state *typeCodeAssignmentState) getNonBasicTypeCode(class string, typecode llvm.Value) *big.Int {
	switch class {
	case "chan", "pointer", "slice":
		// Prefix-style type kinds. The upper bits contain the element type.
		sub := state.builder.CreateExtractValue(typecode.Initializer(), 0, "")
		return state.getTypeCodeNum(sub)
	case "array":
		// An array is basically a pair of (typecode, length) stored in a
		// sidetable.
		return big.NewInt(int64(state.getArrayTypeNum(typecode)))
	case "struct":
		// More complicated type kind. The upper bits contain the index to the
		// struct type in the struct types sidetable.
		return big.NewInt(int64(state.getStructTypeNum(typecode)))
	default:
		// Type has not yet been implemented, so fall back by using a unique
		// number.
		num := big.NewInt(int64(state.fallbackIndex))
		state.fallbackIndex++
		return num
	}
}

// getClassAndValueFromTypeCode takes a typecode (a llvm.Value of type
// runtime.typecodeID), looks at the name, and extracts the typecode class and
// value from it. For example, for a typecode with the following name:
//
//	reflect/types.type:pointer:named:reflect.ValueError
//
// It extracts:
//
//	class = "pointer"
//	value = "named:reflect.ValueError"
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

// getArrayTypeNum returns the array type number, which is an index into the
// reflect.arrayTypesSidetable or a unique number for this type if this table is
// not used.
func (state *typeCodeAssignmentState) getArrayTypeNum(typecode llvm.Value) int {
	name := typecode.Name()
	if num, ok := state.arrayTypes[name]; ok {
		// This array type already has an entry in the sidetable. Don't store
		// it twice.
		return num
	}

	if !state.needsArrayTypesSidetable {
		// We don't need array sidetables, so we can just assign monotonically
		// increasing numbers to each array type.
		num := len(state.arrayTypes)
		state.arrayTypes[name] = num
		return num
	}

	elemTypeCode := state.builder.CreateExtractValue(typecode.Initializer(), 0, "")
	elemTypeNum := state.getTypeCodeNum(elemTypeCode)
	if elemTypeNum.BitLen() > state.uintptrLen || !elemTypeNum.IsUint64() {
		// TODO: make this a regular error
		panic("array element type has a type code that is too big")
	}

	// The array side table is a sequence of {element type, array length}.
	arrayLength := state.builder.CreateExtractValue(typecode.Initializer(), 1, "").ZExtValue()
	buf := makeVarint(elemTypeNum.Uint64())
	buf = append(buf, makeVarint(arrayLength)...)

	index := len(state.arrayTypesSidetable)
	state.arrayTypes[name] = index
	state.arrayTypesSidetable = append(state.arrayTypesSidetable, buf...)
	return index
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
	structTypeGlobal := state.builder.CreateExtractValue(typecode.Initializer(), 0, "").Operand(0).Initializer()
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
		field := state.builder.CreateExtractValue(structTypeGlobal, i, "")

		nameGlobal := state.builder.CreateExtractValue(field, 1, "")
		if nameGlobal == llvm.ConstPointerNull(nameGlobal.Type()) {
			panic("compiler: no name for this struct field")
		}
		fieldNameBytes := getGlobalBytes(nameGlobal.Operand(0), state.builder)
		fieldNameNumber := state.getStructNameNumber(fieldNameBytes)

		// See whether this struct field has an associated tag, and if so,
		// store that tag in the tags sidetable.
		tagGlobal := state.builder.CreateExtractValue(field, 2, "")
		hasTag := false
		tagNumber := 0
		if tagGlobal != llvm.ConstPointerNull(tagGlobal.Type()) {
			hasTag = true
			tagBytes := getGlobalBytes(tagGlobal.Operand(0), state.builder)
			tagNumber = state.getStructNameNumber(tagBytes)
		}

		// The 'embedded' or 'anonymous' flag for this field.
		embedded := state.builder.CreateExtractValue(field, 3, "").ZExtValue() != 0

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
		typeNum := state.getTypeCodeNum(state.builder.CreateExtractValue(field, 0, ""))
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
