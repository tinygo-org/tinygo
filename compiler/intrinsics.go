package compiler

// This file contains helper functions to create calls to LLVM intrinsics.

import (
	"strconv"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createMemoryCopyCall creates a call to a builtin LLVM memcpy or memmove
// function, declaring this function if needed. These calls are treated
// specially by optimization passes possibly resulting in better generated code,
// and will otherwise be lowered to regular libc memcpy/memmove calls.
func (b *builder) createMemoryCopyCall(fn *ssa.Function, args []ssa.Value) (llvm.Value, error) {
	fnName := "llvm." + fn.Name() + ".p0i8.p0i8.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	llvmFn := b.mod.NamedFunction(fnName)
	if llvmFn.IsNil() {
		fnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.i8ptrType, b.i8ptrType, b.uintptrType, b.ctx.Int1Type()}, false)
		llvmFn = llvm.AddFunction(b.mod, fnName, fnType)
	}
	var params []llvm.Value
	for _, param := range args {
		params = append(params, b.getValue(param))
	}
	params = append(params, llvm.ConstInt(b.ctx.Int1Type(), 0, false))
	b.CreateCall(llvmFn, params, "")
	return llvm.Value{}, nil
}
