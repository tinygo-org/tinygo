package compiler

// This file implements function values and closures. It may need some lowering
// in a later step, see func-lowering.go.

import (
	"fmt"
	"go/types"
	"sort"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// LLVM recursively expands each struct field and array element into separate
// values. Pass larger aggregates indirectly before LLVM expands them.
const maxDirectAggregateValues = 1024

// The WebAssembly JavaScript API limits function types to 1000 parameters.
// Apply the same internal ABI cap on every target.
const maxFunctionParams = 1000

type functionABIParam struct {
	llvmType  llvm.Type
	indirect  bool
	leafCount uint64
}

type functionABI struct {
	resultType     llvm.Type
	indirectResult bool
	params         []functionABIParam
}

type functionABIKey struct {
	signature         *types.Signature
	exported          bool
	interfaceReceiver bool
	budgetReceiverPtr bool
	extraParams       uint64
}

func (c *compilerContext) getFunctionABI(sig *types.Signature, exported bool) functionABI {
	budgetReceiverPtr := sig.Recv() != nil && !exported
	extraParams := uint64(0)
	if budgetReceiverPtr {
		// Keep ordinary parameter decisions identical between concrete method
		// calls and interface invokes.
		extraParams++ // interface typecode
	}
	return c.getFunctionABIWithReceiver(sig, exported, false, budgetReceiverPtr, extraParams)
}

func (c *compilerContext) getInterfaceFunctionABI(sig *types.Signature) functionABI {
	return c.getFunctionABIWithReceiver(sig, false, true, false, 1)
}

// getFunctionABIWithReceiver lowers the fewest aggregate parameters necessary
// to keep the complete scalarized signature within the internal ABI cap.
func (c *compilerContext) getFunctionABIWithReceiver(sig *types.Signature, exported, interfaceReceiver, budgetReceiverPtr bool, extraParams uint64) functionABI {
	key := functionABIKey{sig, exported, interfaceReceiver, budgetReceiverPtr, extraParams}
	if abi, ok := c.functionABIs[key]; ok {
		return abi
	}

	abi := functionABI{}
	abi.resultType, abi.indirectResult = c.hasIndirectResult(sig)
	if exported {
		abi.indirectResult = false
	}

	for i, param := range getParams(sig) {
		llvmType := c.getLLVMType(param.Type())
		if i == 0 && interfaceReceiver {
			llvmType = c.dataPtrType
		}
		leafCount, exceeded := aggregateValueCountLimit(llvmType, 0, maxFunctionParams)
		if exceeded {
			leafCount = maxFunctionParams + 1
		}
		abi.params = append(abi.params, functionABIParam{
			llvmType:  llvmType,
			indirect:  !exported && c.isIndirectAggregate(llvmType),
			leafCount: leafCount,
		})
	}

	if exported {
		c.functionABIs[key] = abi
		return abi
	}

	count := uint64(1) + extraParams // context and synthetic parameters
	if abi.indirectResult {
		count++
	} else if aggregateValueCountExceeds(abi.resultType, 1) {
		count++
	}

	var candidates []int
	for i, param := range abi.params {
		if param.indirect {
			count++
			continue
		}
		if i == 0 && budgetReceiverPtr {
			count++
			continue
		}
		count += param.leafCount
		switch param.llvmType.TypeKind() {
		case llvm.ArrayTypeKind, llvm.StructTypeKind:
			if param.leafCount > 1 {
				candidates = append(candidates, i)
			}
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return abi.params[candidates[i]].leafCount > abi.params[candidates[j]].leafCount
	})
	// Minimize ABI changes by lowering the largest aggregates first.
	for _, i := range candidates {
		if count <= maxFunctionParams {
			break
		}
		abi.params[i].indirect = true
		count -= abi.params[i].leafCount - 1
	}
	if budgetReceiverPtr && !abi.params[0].indirect {
		concreteCount := count - extraParams - 1 + abi.params[0].leafCount
		if concreteCount > maxFunctionParams {
			abi.params[0].indirect = true
		}
	}

	c.functionABIs[key] = abi
	return abi
}

// ValidateWasmFunctionParameters checks the final LLVM module before the
// WebAssembly backend expands aggregate parameters into scalar values.
func ValidateWasmFunctionParameters(mod llvm.Module) error {
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() && fn.FirstUse().IsNil() {
			continue
		}
		count := uint64(0)
		if returnType := fn.GlobalValueType().ReturnType(); returnType.TypeKind() == llvm.ArrayTypeKind || returnType.TypeKind() == llvm.StructTypeKind {
			if _, indirect := aggregateValueCountLimit(returnType, 0, 1); indirect {
				count++
			}
		}
		for _, paramType := range fn.GlobalValueType().ParamTypes() {
			var exceeded bool
			count, exceeded = aggregateValueCountLimit(paramType, count, maxFunctionParams)
			if exceeded {
				return fmt.Errorf("function %s has more than %d WebAssembly parameters after ABI lowering", fn.Name(), maxFunctionParams)
			}
		}
	}
	return nil
}

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
	return aggregateValueCountExceeds(typ, maxDirectAggregateValues)
}

func aggregateValueCountExceeds(typ llvm.Type, limit uint64) bool {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind, llvm.StructTypeKind:
		_, exceeded := aggregateValueCountLimit(typ, 0, limit)
		return exceeded
	default:
		return false
	}
}

func aggregateValueCount(typ llvm.Type, count uint64) (uint64, bool) {
	return aggregateValueCountLimit(typ, count, maxDirectAggregateValues)
}

func aggregateValueCountLimit(typ llvm.Type, count, limit uint64) (uint64, bool) {
	switch typ.TypeKind() {
	case llvm.ArrayTypeKind:
		length := uint64(typ.ArrayLength())
		if length == 0 {
			return count, false
		}
		elementCount, exceeded := aggregateValueCountLimit(typ.ElementType(), 0, limit)
		if exceeded {
			return count, true
		}
		if elementCount != 0 && length > (limit-count)/elementCount {
			return count, true
		}
		return count + length*elementCount, false
	case llvm.StructTypeKind:
		for _, field := range typ.StructElementTypes() {
			var exceeded bool
			count, exceeded = aggregateValueCountLimit(field, count, limit)
			if exceeded {
				return count, true
			}
		}
		return count, false
	default:
		count++
		return count, count > limit
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
	var abi functionABI
	if typ.Recv() != nil && c.getLLVMType(typ.Recv().Type()).StructName() == "runtime._interface" {
		abi = c.getFunctionABIWithReceiver(typ, false, true, false, 0)
	} else {
		abi = c.getFunctionABI(typ, false)
	}
	returnType := abi.resultType

	// Get the parameter types.
	var paramTypes []llvm.Type
	if abi.indirectResult {
		// LLVM expands aggregate returns into scalar leaves before deciding
		// whether to pass them indirectly, so a large IR return can exhaust
		// memory. Returning void avoids that expansion and cannot be demoted
		// again. Keep the result pointer first so the context remains last.
		paramTypes = append(paramTypes, c.dataPtrType)
		returnType = c.ctx.VoidType()
	}
	for _, param := range abi.params {
		if param.indirect {
			paramTypes = append(paramTypes, c.dataPtrType)
		} else {
			for _, info := range c.expandDirectFormalParamType(param.llvmType, "", nil) {
				paramTypes = append(paramTypes, info.llvmType)
			}
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
