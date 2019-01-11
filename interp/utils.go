package interp

import (
	"github.com/aykevl/go-llvm"
)

// Return a list of values (actually, instructions) where this value is used as
// an operand.
func getUses(value llvm.Value) []llvm.Value {
	var uses []llvm.Value
	use := value.FirstUse()
	for !use.IsNil() {
		uses = append(uses, use.User())
		use = use.NextUse()
	}
	return uses
}

// Return a zero LLVM value for any LLVM type. Setting this value as an
// initializer has the same effect as setting 'zeroinitializer' on a value.
// Sadly, I haven't found a way to do it directly with the Go API but this works
// just fine.
func getZeroValue(typ llvm.Type) llvm.Value {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind:
		subTyp := typ.ElementType()
		subVal := getZeroValue(subTyp)
		vals := make([]llvm.Value, typ.ArrayLength())
		for i := range vals {
			vals[i] = subVal
		}
		return llvm.ConstArray(subTyp, vals)
	case llvm.FloatTypeKind, llvm.DoubleTypeKind:
		return llvm.ConstFloat(typ, 0.0)
	case llvm.IntegerTypeKind:
		return llvm.ConstInt(typ, 0, false)
	case llvm.PointerTypeKind:
		return llvm.ConstPointerNull(typ)
	case llvm.StructTypeKind:
		types := typ.StructElementTypes()
		vals := make([]llvm.Value, len(types))
		for i, subTyp := range types {
			val := getZeroValue(subTyp)
			vals[i] = val
		}
		if typ.StructName() != "" {
			return llvm.ConstNamedStruct(typ, vals)
		} else {
			return typ.Context().ConstStruct(vals, false)
		}
	case llvm.VectorTypeKind:
		zero := getZeroValue(typ.ElementType())
		vals := make([]llvm.Value, typ.VectorSize())
		for i := range vals {
			vals[i] = zero
		}
		return llvm.ConstVector(vals, false)
	default:
		panic("interp: unknown LLVM type: " + typ.String())
	}
}

// getStringBytes loads the byte slice of a Go string represented as a
// {ptr, len} pair.
func getStringBytes(strPtr Value, strLen llvm.Value) []byte {
	if !strLen.IsConstant() {
		panic("getStringBytes with a non-constant length")
	}
	buf := make([]byte, strLen.ZExtValue())
	for i := range buf {
		c := strPtr.GetElementPtr([]uint32{uint32(i)}).Load()
		buf[i] = byte(c.ZExtValue())
	}
	return buf
}

// getLLVMIndices converts an []uint32 into an []llvm.Value, for use in
// llvm.ConstGEP.
func getLLVMIndices(int32Type llvm.Type, indices []uint32) []llvm.Value {
	llvmIndices := make([]llvm.Value, len(indices))
	for i, index := range indices {
		llvmIndices[i] = llvm.ConstInt(int32Type, uint64(index), false)
	}
	return llvmIndices
}

// Return true if this type is a scalar value (integer or floating point), false
// otherwise.
func isScalar(t llvm.Type) bool {
	switch t.TypeKind() {
	case llvm.IntegerTypeKind, llvm.FloatTypeKind, llvm.DoubleTypeKind:
		return true
	default:
		return false
	}
}
