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
func (c *Compiler) trackExpr(frame *Frame, expr ssa.Value, value llvm.Value) {
	// There are uses of this expression, Make sure the pointers
	// are tracked during GC.
	switch expr := expr.(type) {
	case *ssa.Alloc, *ssa.MakeChan, *ssa.MakeMap:
		// These values are always of pointer type in IR.
		c.trackPointer(value)
	case *ssa.Call, *ssa.Convert, *ssa.MakeClosure, *ssa.MakeInterface, *ssa.MakeSlice, *ssa.Next:
		if !value.IsNil() {
			c.trackValue(value)
		}
	case *ssa.Select:
		if alloca, ok := frame.selectRecvBuf[expr]; ok {
			if alloca.IsAUndefValue().IsNil() {
				c.trackPointer(alloca)
			}
		}
	case *ssa.UnOp:
		switch expr.Op {
		case token.MUL:
			// Pointer dereference.
			c.trackValue(value)
		case token.ARROW:
			// Channel receive operator.
			// It's not necessary to look at commaOk here, because in that
			// case it's just an aggregate and trackValue will extract the
			// pointer in there (if there is one).
			c.trackValue(value)
		}
	}
}

// trackValue locates pointers in a value (possibly an aggregate) and tracks the
// individual pointers
func (c *Compiler) trackValue(value llvm.Value) {
	typ := value.Type()
	switch typ.TypeKind() {
	case llvm.PointerTypeKind:
		c.trackPointer(value)
	case llvm.StructTypeKind:
		if !typeHasPointers(typ) {
			return
		}
		numElements := typ.StructElementTypesCount()
		for i := 0; i < numElements; i++ {
			subValue := c.builder.CreateExtractValue(value, i, "")
			c.trackValue(subValue)
		}
	case llvm.ArrayTypeKind:
		if !typeHasPointers(typ) {
			return
		}
		numElements := typ.ArrayLength()
		for i := 0; i < numElements; i++ {
			subValue := c.builder.CreateExtractValue(value, i, "")
			c.trackValue(subValue)
		}
	}
}

// trackPointer creates a call to runtime.trackPointer, bitcasting the poitner
// first if needed. The input value must be of LLVM pointer type.
func (c *Compiler) trackPointer(value llvm.Value) {
	if value.Type() != c.i8ptrType {
		value = c.builder.CreateBitCast(value, c.i8ptrType, "")
	}
	c.createRuntimeCall("trackPointer", []llvm.Value{value}, "")
}

// trackPointer creates a call to runtime.trackPointer, bitcasting the poitner
// first if needed. The input value must be of LLVM pointer type.
func (b *builder) trackPointer(value llvm.Value) {
	if value.Type() != b.i8ptrType {
		value = b.CreateBitCast(value, b.i8ptrType, "")
	}
	b.createRuntimeCall("trackPointer", []llvm.Value{value}, "")
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
