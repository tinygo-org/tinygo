package compiler

// This file implements function values and closures. It may need some lowering
// in a later step, see func-lowering.go.

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

type funcValueImplementation int

const (
	funcValueNone funcValueImplementation = iota

	// A func value is implemented as a pair of pointers:
	//     {context, function pointer}
	// where the context may be a pointer to a heap-allocated struct containing
	// the free variables, or it may be undef if the function being pointed to
	// doesn't need a context. The function pointer is a regular function
	// pointer.
	funcValueDoubleword

	// As funcValueDoubleword, but with the function pointer replaced by a
	// unique ID per function signature. Function values are called by using a
	// switch statement and choosing which function to call.
	funcValueSwitch
)

// funcImplementation picks an appropriate func value implementation for the
// target.
func (c *Compiler) funcImplementation() funcValueImplementation {
	if c.GOARCH == "wasm" {
		return funcValueSwitch
	} else {
		return funcValueDoubleword
	}
}

// createFuncValue creates a function value from a raw function pointer with no
// context.
func (c *Compiler) createFuncValue(funcPtr, context llvm.Value, sig *types.Signature) llvm.Value {
	var funcValueScalar llvm.Value
	switch c.funcImplementation() {
	case funcValueDoubleword:
		// Closure is: {context, function pointer}
		funcValueScalar = funcPtr
	case funcValueSwitch:
		sigGlobal := c.getFuncSignature(sig)
		funcValueWithSignatureGlobalName := funcPtr.Name() + "$withSignature"
		funcValueWithSignatureGlobal := c.mod.NamedGlobal(funcValueWithSignatureGlobalName)
		if funcValueWithSignatureGlobal.IsNil() {
			funcValueWithSignatureType := c.mod.GetTypeByName("runtime.funcValueWithSignature")
			funcValueWithSignature := llvm.ConstNamedStruct(funcValueWithSignatureType, []llvm.Value{
				llvm.ConstPtrToInt(funcPtr, c.uintptrType),
				sigGlobal,
			})
			funcValueWithSignatureGlobal = llvm.AddGlobal(c.mod, funcValueWithSignatureType, funcValueWithSignatureGlobalName)
			funcValueWithSignatureGlobal.SetInitializer(funcValueWithSignature)
			funcValueWithSignatureGlobal.SetGlobalConstant(true)
			funcValueWithSignatureGlobal.SetLinkage(llvm.InternalLinkage)
		}
		funcValueScalar = llvm.ConstPtrToInt(funcValueWithSignatureGlobal, c.uintptrType)
	default:
		panic("unimplemented func value variant")
	}
	funcValueType := c.getFuncType(sig)
	funcValue := llvm.Undef(funcValueType)
	funcValue = c.builder.CreateInsertValue(funcValue, context, 0, "")
	funcValue = c.builder.CreateInsertValue(funcValue, funcValueScalar, 1, "")
	return funcValue
}

// getFuncSignature returns a global for identification of a particular function
// signature. It is used in runtime.funcValueWithSignature and in calls to
// getFuncPtr.
func (c *Compiler) getFuncSignature(sig *types.Signature) llvm.Value {
	typeCodeName := getTypeCodeName(sig)
	sigGlobalName := "reflect/types.type:" + typeCodeName
	sigGlobal := c.mod.NamedGlobal(sigGlobalName)
	if sigGlobal.IsNil() {
		sigGlobal = llvm.AddGlobal(c.mod, c.ctx.Int8Type(), sigGlobalName)
		sigGlobal.SetInitializer(llvm.Undef(c.ctx.Int8Type()))
		sigGlobal.SetGlobalConstant(true)
		sigGlobal.SetLinkage(llvm.InternalLinkage)
	}
	return sigGlobal
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
func (c *Compiler) decodeFuncValue(funcValue llvm.Value, sig *types.Signature) (funcPtr, context llvm.Value, err error) {
	context = c.builder.CreateExtractValue(funcValue, 0, "")
	switch c.funcImplementation() {
	case funcValueDoubleword:
		funcPtr = c.builder.CreateExtractValue(funcValue, 1, "")
	case funcValueSwitch:
		llvmSig := c.getRawFuncType(sig)
		sigGlobal := c.getFuncSignature(sig)
		funcPtr = c.createRuntimeCall("getFuncPtr", []llvm.Value{funcValue, sigGlobal}, "")
		funcPtr = c.builder.CreateIntToPtr(funcPtr, llvmSig, "")
	default:
		panic("unimplemented func value variant")
	}
	return
}

// getFuncType returns the type of a func value given a signature.
func (c *Compiler) getFuncType(typ *types.Signature) llvm.Type {
	switch c.funcImplementation() {
	case funcValueDoubleword:
		rawPtr := c.getRawFuncType(typ)
		return c.ctx.StructType([]llvm.Type{c.i8ptrType, rawPtr}, false)
	case funcValueSwitch:
		return c.mod.GetTypeByName("runtime.funcValue")
	default:
		panic("unimplemented func value variant")
	}
}

// getRawFuncType returns a LLVM function pointer type for a given signature.
func (c *Compiler) getRawFuncType(typ *types.Signature) llvm.Type {
	// Get the return type.
	var returnType llvm.Type
	switch typ.Results().Len() {
	case 0:
		// No return values.
		returnType = c.ctx.VoidType()
	case 1:
		// Just one return value.
		returnType = c.getLLVMType(typ.Results().At(0).Type())
	default:
		// Multiple return values. Put them together in a struct.
		// This appears to be the common way to handle multiple return values in
		// LLVM.
		members := make([]llvm.Type, typ.Results().Len())
		for i := 0; i < typ.Results().Len(); i++ {
			members[i] = c.getLLVMType(typ.Results().At(i).Type())
		}
		returnType = c.ctx.StructType(members, false)
	}

	// Get the parameter types.
	var paramTypes []llvm.Type
	if typ.Recv() != nil {
		recv := c.getLLVMType(typ.Recv().Type())
		if recv.StructName() == "runtime._interface" {
			// This is a call on an interface, not a concrete type.
			// The receiver is not an interface, but a i8* type.
			recv = c.i8ptrType
		}
		paramTypes = append(paramTypes, c.expandFormalParamType(recv)...)
	}
	for i := 0; i < typ.Params().Len(); i++ {
		subType := c.getLLVMType(typ.Params().At(i).Type())
		paramTypes = append(paramTypes, c.expandFormalParamType(subType)...)
	}
	// All functions take these parameters at the end.
	paramTypes = append(paramTypes, c.i8ptrType) // context
	paramTypes = append(paramTypes, c.i8ptrType) // parent coroutine

	// Make a func type out of the signature.
	return llvm.PointerType(llvm.FunctionType(returnType, paramTypes, false), c.funcPtrAddrSpace)
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

	// Create the closure.
	return c.createFuncValue(f.LLVMFn, context, f.Signature), nil
}
