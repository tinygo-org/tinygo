package transform

import (
	"iter"
	"slices"

	"tinygo.org/x/go-llvm"
)

// getUses returns the values (actually, instructions) where this value is used
// as an operand. Use getAllUses if the loop body changes the use list.
func getUses(value llvm.Value) iter.Seq[llvm.Value] {
	return func(yield func(llvm.Value) bool) {
		if value.IsNil() {
			return
		}
		for use := value.FirstUse(); !use.IsNil(); use = use.NextUse() {
			if !yield(use.User()) {
				return
			}
		}
	}
}

// getAllUses returns the uses of value as a slice.
func getAllUses(value llvm.Value) []llvm.Value {
	return slices.Collect(getUses(value))
}

// hasUses returns whether the given value has any uses. It is equivalent to
// getAllUses(value) != nil but faster.
func hasUses(value llvm.Value) bool {
	if value.IsNil() {
		return false
	}
	return !value.FirstUse().IsNil()
}

// typeHasPointers returns whether a value of this type can hold a pointer.
//
// Vectors report false. That is precise for the IR TinyGo produces: nothing
// lowers a Go value into a vector of pointers, and OptimizeAllocs runs long
// before any vectorizer. It is also why escape analysis may rely on this: a
// pointer can only reach a vector lane through insertelement, which
// valueEscapesAt treats as an escape. A pass that starts feeding pointer
// vectors to either caller needs a case here first.
func typeHasPointers(typ llvm.Type) bool {
	switch typ.TypeKind() {
	case llvm.PointerTypeKind:
		return true
	case llvm.StructTypeKind:
		if slices.ContainsFunc(typ.StructElementTypes(), typeHasPointers) {
			return true
		}
	case llvm.ArrayTypeKind:
		return typ.ArrayLength() != 0 && typeHasPointers(typ.ElementType())
	}
	return false
}

// elementType returns the type of element i of the struct or array type typ.
func elementType(typ llvm.Type, i int) llvm.Type {
	switch typ.TypeKind() {
	case llvm.StructTypeKind:
		return typ.StructTypeAtIndex(i)
	default:
		return typ.ElementType()
	}
}

// stripPointerCasts strips instruction pointer casts (getelementptr and
// bitcast) and returns the original value without the casts.
func stripPointerCasts(value llvm.Value) llvm.Value {
	if !value.IsAConstantExpr().IsNil() {
		switch value.Opcode() {
		case llvm.GetElementPtr, llvm.BitCast:
			return stripPointerCasts(value.Operand(0))
		}
	}
	if !value.IsAInstruction().IsNil() {
		switch value.InstructionOpcode() {
		case llvm.GetElementPtr, llvm.BitCast:
			return stripPointerCasts(value.Operand(0))
		}
	}
	return value
}

// memoryEffects holds two bits, ref and mod, per memory location. See
// https://github.com/llvm/llvm-project/blob/94ebcfd16dac67486bae624f74e1c5c789448bae/llvm/include/llvm/Support/ModRef.h#L62
type memoryEffects uint64

var memoryAttrKind = llvm.AttributeKindID("memory")

// readMemoryEffects returns the memory effects of fn. A function without a
// memory attribute may read and write any memory.
func readMemoryEffects(fn llvm.Value) memoryEffects {
	attr := fn.GetEnumFunctionAttribute(memoryAttrKind)
	if attr.IsNil() {
		return ^memoryEffects(0)
	}
	return memoryEffects(attr.GetEnumValue())
}

// onlyAccessesArgMem returns whether the function reads or writes only memory
// that its pointer arguments point to.
func (effects memoryEffects) onlyAccessesArgMem() bool {
	return effects>>2 == 0
}

// onlyReadsMemory returns whether the function writes no memory.
func (effects memoryEffects) onlyReadsMemory() bool {
	return effects&0xaaaaaaaa == 0
}
