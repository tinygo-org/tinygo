package transform

import (
	"slices"

	"tinygo.org/x/go-llvm"
)

// Return a list of values (actually, instructions) where this value is used as
// an operand.
func getUses(value llvm.Value) []llvm.Value {
	if value.IsNil() {
		return nil
	}
	var uses []llvm.Value
	use := value.FirstUse()
	for !use.IsNil() {
		uses = append(uses, use.User())
		use = use.NextUse()
	}
	return uses
}

// hasUses returns whether the given value has any uses. It is equivalent to
// getUses(value) != nil but faster.
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
