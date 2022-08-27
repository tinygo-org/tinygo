package compiler

// This file contains helper functions to create calls to LLVM intrinsics.

import (
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// Define unimplemented intrinsic functions.
//
// Some functions are either normally implemented in Go assembly (like
// sync/atomic functions) or intentionally left undefined to be implemented
// directly in the compiler (like runtime/volatile functions). Either way, look
// for these and implement them if this is the case.
func (b *builder) defineIntrinsicFunction() {
	name := b.fn.RelString(nil)
	switch {
	case name == "runtime.memcpy" || name == "runtime.memmove":
		b.createMemoryCopyImpl()
	case name == "runtime.memzero":
		b.createMemoryZeroImpl()
	case strings.HasPrefix(name, "runtime/volatile.Load"):
		b.createVolatileLoad()
	case strings.HasPrefix(name, "runtime/volatile.Store"):
		b.createVolatileStore()
	case strings.HasPrefix(name, "sync/atomic.") && token.IsExported(b.fn.Name()):
		b.createFunctionStart()
		returnValue := b.createAtomicOp(b.fn.Name())
		if !returnValue.IsNil() {
			b.CreateRet(returnValue)
		} else {
			b.CreateRetVoid()
		}
	}
}

// createMemoryCopyImpl creates a call to a builtin LLVM memcpy or memmove
// function, declaring this function if needed. These calls are treated
// specially by optimization passes possibly resulting in better generated code,
// and will otherwise be lowered to regular libc memcpy/memmove calls.
func (b *builder) createMemoryCopyImpl() {
	b.createFunctionStart()
	fnName := "llvm." + b.fn.Name() + ".p0i8.p0i8.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	llvmFn := b.mod.NamedFunction(fnName)
	if llvmFn.IsNil() {
		fnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.i8ptrType, b.i8ptrType, b.uintptrType, b.ctx.Int1Type()}, false)
		llvmFn = llvm.AddFunction(b.mod, fnName, fnType)
	}
	var params []llvm.Value
	for _, param := range b.fn.Params {
		params = append(params, b.getValue(param))
	}
	params = append(params, llvm.ConstInt(b.ctx.Int1Type(), 0, false))
	b.CreateCall(llvmFn, params, "")
	b.CreateRetVoid()
}

// createMemoryZeroImpl creates calls to llvm.memset.* to zero a block of
// memory, declaring the function if needed. These calls will be lowered to
// regular libc memset calls if they aren't optimized out in a different way.
func (b *builder) createMemoryZeroImpl() {
	b.createFunctionStart()
	fnName := "llvm.memset.p0i8.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	llvmFn := b.mod.NamedFunction(fnName)
	if llvmFn.IsNil() {
		fnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.i8ptrType, b.ctx.Int8Type(), b.uintptrType, b.ctx.Int1Type()}, false)
		llvmFn = llvm.AddFunction(b.mod, fnName, fnType)
	}
	params := []llvm.Value{
		b.getValue(b.fn.Params[0]),
		llvm.ConstInt(b.ctx.Int8Type(), 0, false),
		b.getValue(b.fn.Params[1]),
		llvm.ConstInt(b.ctx.Int1Type(), 0, false),
	}
	b.CreateCall(llvmFn, params, "")
	b.CreateRetVoid()
}

var mathToLLVMMapping = map[string]string{
	"math.Sqrt":  "llvm.sqrt.f64",
	"math.Floor": "llvm.floor.f64",
	"math.Ceil":  "llvm.ceil.f64",
	"math.Trunc": "llvm.trunc.f64",
}

// createMathOp lowers the given call as a LLVM math intrinsic. It returns the
// resulting value.
func (b *builder) createMathOp(call *ssa.CallCommon) llvm.Value {
	llvmName := mathToLLVMMapping[call.StaticCallee().RelString(nil)]
	if llvmName == "" {
		panic("unreachable: unknown math operation") // sanity check
	}
	llvmFn := b.mod.NamedFunction(llvmName)
	if llvmFn.IsNil() {
		// The intrinsic doesn't exist yet, so declare it.
		// At the moment, all supported intrinsics have the form "double
		// foo(double %x)" so we can hardcode the signature here.
		llvmType := llvm.FunctionType(b.ctx.DoubleType(), []llvm.Type{b.ctx.DoubleType()}, false)
		llvmFn = llvm.AddFunction(b.mod, llvmName, llvmType)
	}
	// Create a call to the intrinsic.
	args := make([]llvm.Value, len(call.Args))
	for i, arg := range call.Args {
		args[i] = b.getValue(arg)
	}
	return b.CreateCall(llvmFn, args, "")
}
