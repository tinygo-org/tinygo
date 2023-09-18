package compiler

// This file provides IR transformations necessary for precise and portable
// garbage collectors.

import (
	"go/token"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// trackExpr inserts pointer tracking intrinsics for the GC if the expression is
// one of the expressions that need this.
func (b *builder) trackExpr(expr ssa.Value, value llvm.Value) {
	// There are uses of this expression, Make sure the pointers
	// are tracked during GC.
	switch expr := expr.(type) {
	case *ssa.Alloc, *ssa.MakeChan, *ssa.MakeMap:
		// These values are always of pointer type in IR.
		b.trackPointer(value)
	case *ssa.Call, *ssa.Convert, *ssa.MakeClosure, *ssa.MakeInterface, *ssa.MakeSlice, *ssa.Next:
		if !value.IsNil() {
			b.trackValue(value)
		}
	case *ssa.Select:
		if alloca, ok := b.selectRecvBuf[expr]; ok {
			if alloca.IsAUndefValue().IsNil() {
				b.trackPointer(alloca)
			}
		}
	case *ssa.UnOp:
		switch expr.Op {
		case token.MUL:
			// Pointer dereference.
			b.trackValue(value)
		case token.ARROW:
			// Channel receive operator.
			// It's not necessary to look at commaOk here, because in that
			// case it's just an aggregate and trackValue will extract the
			// pointer in there (if there is one).
			b.trackValue(value)
		}
	case *ssa.BinOp:
		switch expr.Op {
		case token.ADD:
			// String concatenation.
			b.trackValue(value)
		}
	}
}

// trackValue locates pointers in a value (possibly an aggregate) and tracks the
// individual pointers
func (b *builder) trackValue(value llvm.Value) {
	typ := value.Type()
	switch typ.TypeKind() {
	case llvm.PointerTypeKind:
		b.trackPointer(value)
	case llvm.StructTypeKind:
		if !typeHasPointers(typ) {
			return
		}
		numElements := typ.StructElementTypesCount()
		for i := 0; i < numElements; i++ {
			subValue := b.CreateExtractValue(value, i, "")
			b.trackValue(subValue)
		}
	case llvm.ArrayTypeKind:
		if !typeHasPointers(typ) {
			return
		}
		numElements := typ.ArrayLength()
		for i := 0; i < numElements; i++ {
			subValue := b.CreateExtractValue(value, i, "")
			b.trackValue(subValue)
		}
	}
}

// trackPointer creates a call to runtime.trackPointer, bitcasting the poitner
// first if needed. The input value must be of LLVM pointer type.
func (b *builder) trackPointer(value llvm.Value) {
	b.createRuntimeCall("trackPointer", []llvm.Value{value, b.stackChainAlloca}, "")
}

// typeHasPointers returns whether this type is a pointer or contains pointers.
// If the type is an aggregate type, it will check whether there is a pointer
// inside.
func typeHasPointers(t llvm.Type) bool {
	switch t.TypeKind() {
	case llvm.PointerTypeKind:
		return true
	case llvm.StructTypeKind:
		for _, subType := range t.StructElementTypes() {
			if typeHasPointers(subType) {
				return true
			}
		}
		return false
	case llvm.ArrayTypeKind:
		if typeHasPointers(t.ElementType()) {
			return true
		}
		return false
	default:
		return false
	}
}
