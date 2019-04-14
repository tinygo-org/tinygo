package compiler

// This file implements function values and closures. A func value is
// implemented as a pair of pointers: {context, function pointer}, where the
// context may be a pointer to a heap-allocated struct containing the free
// variables, or it may be undef if the function being pointed to doesn't need a
// context.

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createFuncValue creates a function value from a raw function pointer with no
// context.
func (c *Compiler) createFuncValue(funcPtr llvm.Value) (llvm.Value, error) {
	// Closure is: {context, function pointer}
	return c.ctx.ConstStruct([]llvm.Value{
		llvm.Undef(c.i8ptrType),
		funcPtr,
	}, false), nil
}

// extractFuncScalar returns some scalar that can be used in comparisons. It is
// a cheap operation.
func (c *Compiler) extractFuncScalar(funcValue llvm.Value) llvm.Value {
	return c.builder.CreateExtractValue(funcValue, 1, "")
}

// extractFuncContext extracts the context pointer from this function value. It
// is a cheap operation.
func (c *Compiler) extractFuncContext(funcValue llvm.Value) llvm.Value {
	return c.builder.CreateExtractValue(funcValue, 0, "")
}

// decodeFuncValue extracts the context and the function pointer from this func
// value. This may be an expensive operation.
func (c *Compiler) decodeFuncValue(funcValue llvm.Value) (funcPtr, context llvm.Value) {
	context = c.builder.CreateExtractValue(funcValue, 0, "")
	funcPtr = c.builder.CreateExtractValue(funcValue, 1, "")
	return
}

// getFuncType returns the type of a func value given a signature.
func (c *Compiler) getFuncType(typ *types.Signature) (llvm.Type, error) {
	rawPtr, err := c.getRawFuncType(typ)
	if err != nil {
		return llvm.Type{}, err
	}
	return c.ctx.StructType([]llvm.Type{c.i8ptrType, rawPtr}, false), nil
}

// getRawFuncType returns a LLVM function pointer type for a given signature.
func (c *Compiler) getRawFuncType(typ *types.Signature) (llvm.Type, error) {
	// Get the return type.
	var err error
	var returnType llvm.Type
	switch typ.Results().Len() {
	case 0:
		// No return values.
		returnType = c.ctx.VoidType()
	case 1:
		// Just one return value.
		returnType, err = c.getLLVMType(typ.Results().At(0).Type())
		if err != nil {
			return llvm.Type{}, err
		}
	default:
		// Multiple return values. Put them together in a struct.
		// This appears to be the common way to handle multiple return values in
		// LLVM.
		members := make([]llvm.Type, typ.Results().Len())
		for i := 0; i < typ.Results().Len(); i++ {
			returnType, err := c.getLLVMType(typ.Results().At(i).Type())
			if err != nil {
				return llvm.Type{}, err
			}
			members[i] = returnType
		}
		returnType = c.ctx.StructType(members, false)
	}

	// Get the parameter types.
	var paramTypes []llvm.Type
	if typ.Recv() != nil {
		recv, err := c.getLLVMType(typ.Recv().Type())
		if err != nil {
			return llvm.Type{}, err
		}
		if recv.StructName() == "runtime._interface" {
			// This is a call on an interface, not a concrete type.
			// The receiver is not an interface, but a i8* type.
			recv = c.i8ptrType
		}
		paramTypes = append(paramTypes, c.expandFormalParamType(recv)...)
	}
	for i := 0; i < typ.Params().Len(); i++ {
		subType, err := c.getLLVMType(typ.Params().At(i).Type())
		if err != nil {
			return llvm.Type{}, err
		}
		paramTypes = append(paramTypes, c.expandFormalParamType(subType)...)
	}
	// All functions take these parameters at the end.
	paramTypes = append(paramTypes, c.i8ptrType) // context
	paramTypes = append(paramTypes, c.i8ptrType) // parent coroutine

	// Make a func type out of the signature.
	return llvm.PointerType(llvm.FunctionType(returnType, paramTypes, false), c.funcPtrAddrSpace), nil
}

// parseMakeClosure makes a function value (with context) from the given
// closure expression.
func (c *Compiler) parseMakeClosure(frame *Frame, expr *ssa.MakeClosure) (llvm.Value, error) {
	if len(expr.Bindings) == 0 {
		panic("unexpected: MakeClosure without bound variables")
	}
	f := c.ir.GetFunction(expr.Fn.(*ssa.Function))

	// Collect all bound variables.
	boundVars := make([]llvm.Value, 0, len(expr.Bindings))
	boundVarTypes := make([]llvm.Type, 0, len(expr.Bindings))
	for _, binding := range expr.Bindings {
		// The context stores the bound variables.
		llvmBoundVar, err := c.parseExpr(frame, binding)
		if err != nil {
			return llvm.Value{}, err
		}
		boundVars = append(boundVars, llvmBoundVar)
		boundVarTypes = append(boundVarTypes, llvmBoundVar.Type())
	}
	contextType := c.ctx.StructType(boundVarTypes, false)

	// Allocate memory for the context.
	contextAlloc := llvm.Value{}
	contextHeapAlloc := llvm.Value{}
	if c.targetData.TypeAllocSize(contextType) <= c.targetData.TypeAllocSize(c.i8ptrType) {
		// Context fits in a pointer - e.g. when it is a pointer. Store it
		// directly in the stack after a convert.
		// Because contextType is a struct and we have to cast it to a *i8,
		// store it in an alloca first for bitcasting (store+bitcast+load).
		contextAlloc = c.builder.CreateAlloca(contextType, "")
	} else {
		// Context is bigger than a pointer, so allocate it on the heap.
		size := c.targetData.TypeAllocSize(contextType)
		sizeValue := llvm.ConstInt(c.uintptrType, size, false)
		contextHeapAlloc = c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "")
		contextAlloc = c.builder.CreateBitCast(contextHeapAlloc, llvm.PointerType(contextType, 0), "")
	}

	// Store all bound variables in the alloca or heap pointer.
	for i, boundVar := range boundVars {
		indices := []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false),
		}
		gep := c.builder.CreateInBoundsGEP(contextAlloc, indices, "")
		c.builder.CreateStore(boundVar, gep)
	}

	context := llvm.Value{}
	if c.targetData.TypeAllocSize(contextType) <= c.targetData.TypeAllocSize(c.i8ptrType) {
		// Load value (as *i8) from the alloca.
		contextAlloc = c.builder.CreateBitCast(contextAlloc, llvm.PointerType(c.i8ptrType, 0), "")
		context = c.builder.CreateLoad(contextAlloc, "")
	} else {
		// Get the original heap allocation pointer, which already is an
		// *i8.
		context = contextHeapAlloc
	}

	// Get the function signature type, which is a closure type.
	// A closure is a tuple of {context, function pointer}.
	typ, err := c.getFuncType(f.Signature)
	if err != nil {
		return llvm.Value{}, err
	}

	// Create the closure, which is a struct: {context, function pointer}.
	closure, err := c.getZeroValue(typ)
	if err != nil {
		return llvm.Value{}, err
	}
	closure = c.builder.CreateInsertValue(closure, f.LLVMFn, 1, "")
	closure = c.builder.CreateInsertValue(closure, context, 0, "")
	return closure, nil
}
