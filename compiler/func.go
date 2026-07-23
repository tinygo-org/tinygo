package compiler

// This file implements function values and closures. It may need some lowering
// in a later step, see func-lowering.go.

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// LLVM recursively expands each struct field and array element in parameters
// and results into separate values. It gets very slow with too many values, so
// pass larger aggregates indirectly before LLVM expands them.
const maxDirectAggregateValues = 1024

func (c *compilerContext) getLLVMResultType(sig *types.Signature) llvm.Type {
	switch sig.Results().Len() {
	case 0:
		return c.ctx.VoidType()
	case 1:
		return c.getLLVMType(sig.Results().At(0).Type())
	default:
		results := make([]llvm.Type, sig.Results().Len())
		for i := range results {
			results[i] = c.getLLVMType(sig.Results().At(i).Type())
		}
		return c.ctx.StructType(results, false)
	}
}

func (c *compilerContext) hasIndirectResult(sig *types.Signature) (llvm.Type, bool) {
	resultType := c.getLLVMResultType(sig)
	return resultType, c.isIndirectAggregate(resultType)
}

func (c *compilerContext) isIndirectAggregate(typ llvm.Type) bool {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind, llvm.StructTypeKind:
		_, exceeded := aggregateValueCount(typ, 0)
		return exceeded
	default:
		return false
	}
}

func aggregateValueCount(typ llvm.Type, count uint64) (uint64, bool) {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind:
		length := uint64(typ.ArrayLength())
		if length == 0 {
			return count, false
		}
		elementCount, exceeded := aggregateValueCount(typ.ElementType(), 0)
		if exceeded {
			return count, true
		}
		if elementCount != 0 && length > (maxDirectAggregateValues-count)/elementCount {
			return count, true
		}
		return count + length*elementCount, false
	case llvm.StructTypeKind:
		for _, field := range typ.StructElementTypes() {
			var exceeded bool
			count, exceeded = aggregateValueCount(field, count)
			if exceeded {
				return count, true
			}
		}
		return count, false
	default:
		count++
		return count, count > maxDirectAggregateValues
	}
}

func isLLVMValueType(typ types.Type) bool {
	switch typ := typ.Underlying().(type) {
	case *types.Basic:
		return typ.Kind() != types.Invalid
	case *types.Array:
		return isLLVMValueType(typ.Elem())
	case *types.Struct:
		for field := range typ.Fields() {
			if !isLLVMValueType(field.Type()) {
				return false
			}
		}
		return true
	case *types.Chan, *types.Interface, *types.Map, *types.Pointer, *types.Signature, *types.Slice:
		return true
	default:
		return false
	}
}

// createFuncValue creates a function value from a raw function pointer with no
// context.
func (b *builder) createFuncValue(funcPtr, context llvm.Value, sig *types.Signature) llvm.Value {
	// Closure is: {context, function pointer}
	funcValueType := b.getFuncType(sig)
	funcValue := llvm.Undef(funcValueType)
	funcValue = b.CreateInsertValue(funcValue, context, 0, "")
	funcValue = b.CreateInsertValue(funcValue, funcPtr, 1, "")
	return funcValue
}

// extractFuncScalar returns some scalar that can be used in comparisons. It is
// a cheap operation.
func (b *builder) extractFuncScalar(funcValue llvm.Value) llvm.Value {
	return b.CreateExtractValue(funcValue, 1, "")
}

// extractFuncContext extracts the context pointer from this function value. It
// is a cheap operation.
func (b *builder) extractFuncContext(funcValue llvm.Value) llvm.Value {
	return b.CreateExtractValue(funcValue, 0, "")
}

// decodeFuncValue extracts the context and the function pointer from this func
// value.
func (b *builder) decodeFuncValue(funcValue llvm.Value) (funcPtr, context llvm.Value) {
	context = b.CreateExtractValue(funcValue, 0, "")
	funcPtr = b.CreateExtractValue(funcValue, 1, "")
	return
}

// getFuncType returns the type of a func value given a signature.
func (c *compilerContext) getFuncType(typ *types.Signature) llvm.Type {
	return c.ctx.StructType([]llvm.Type{c.dataPtrType, c.funcPtrType}, false)
}

// getLLVMFunctionType returns a LLVM function type for a given signature.
func (c *compilerContext) getLLVMFunctionType(typ *types.Signature) llvm.Type {
	returnType, indirectResult := c.hasIndirectResult(typ)

	// Get the parameter types.
	var paramTypes []llvm.Type
	if indirectResult {
		// LLVM expands aggregate returns into scalar leaves before deciding
		// whether to pass them indirectly, so a large IR return can exhaust
		// memory. Returning void avoids that expansion and cannot be demoted
		// again. Keep the result pointer first so the context remains last.
		paramTypes = append(paramTypes, c.dataPtrType)
		returnType = c.ctx.VoidType()
	}
	if typ.Recv() != nil {
		recv := c.getLLVMType(typ.Recv().Type())
		if recv.StructName() == "runtime._interface" {
			// This is a call on an interface, not a concrete type.
			// The receiver is not an interface, but a i8* type.
			recv = c.dataPtrType
		}
		for _, info := range c.expandFormalParamType(recv, "", nil) {
			paramTypes = append(paramTypes, info.llvmType)
		}
	}
	for v := range typ.Params().Variables() {
		subType := c.getLLVMType(v.Type())
		for _, info := range c.expandFormalParamType(subType, "", nil) {
			paramTypes = append(paramTypes, info.llvmType)
		}
	}
	// All functions take these parameters at the end.
	paramTypes = append(paramTypes, c.dataPtrType) // context

	// Make a func type out of the signature.
	return llvm.FunctionType(returnType, paramTypes, false)
}

// parseMakeClosure makes a function value (with context) from the given
// closure expression.
func (b *builder) parseMakeClosure(expr *ssa.MakeClosure) (llvm.Value, error) {
	if len(expr.Bindings) == 0 {
		panic("unexpected: MakeClosure without bound variables")
	}
	f := expr.Fn.(*ssa.Function)

	// Collect all bound variables.
	boundVars := make([]llvm.Value, len(expr.Bindings))
	for i, binding := range expr.Bindings {
		// The context stores the bound variables.
		llvmBoundVar := b.getValue(binding, getPos(expr))
		boundVars[i] = llvmBoundVar
	}

	// Store the bound variables in a single object, allocating it on the heap
	// if necessary.
	context := b.emitPointerPack(boundVars, expr.Pos())

	// Create the closure.
	_, fn := b.getFunction(f)
	return b.createFuncValue(fn, context, f.Signature), nil
}
