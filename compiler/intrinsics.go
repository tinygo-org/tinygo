package compiler

// This file contains helper functions to create calls to LLVM intrinsics.

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
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
	case name == "runtime.KeepAlive":
		b.createKeepAliveImpl()
	case strings.HasPrefix(name, "runtime/volatile.Load"):
		b.createVolatileLoad()
	case strings.HasPrefix(name, "runtime/volatile.Store"):
		b.createVolatileStore()
	case strings.HasPrefix(name, "sync/atomic.") && token.IsExported(b.fn.Name()):
		b.createFunctionStart(true)
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
	b.createFunctionStart(true)
	fnName := "llvm." + b.fn.Name() + ".p0.p0.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	if llvmutil.Major() < 15 { // compatibility with LLVM 14
		fnName = "llvm." + b.fn.Name() + ".p0i8.p0i8.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	}
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
	b.CreateCall(llvmFn.GlobalValueType(), llvmFn, params, "")
	b.CreateRetVoid()
}

// createMemoryZeroImpl creates calls to llvm.memset.* to zero a block of
// memory, declaring the function if needed. These calls will be lowered to
// regular libc memset calls if they aren't optimized out in a different way.
func (b *builder) createMemoryZeroImpl() {
	b.createFunctionStart(true)
	fnName := "llvm.memset.p0.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	if llvmutil.Major() < 15 { // compatibility with LLVM 14
		fnName = "llvm.memset.p0i8.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	}
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
	b.CreateCall(llvmFn.GlobalValueType(), llvmFn, params, "")
	b.CreateRetVoid()
}

// createKeepAlive creates the runtime.KeepAlive function. It is implemented
// using inline assembly.
func (b *builder) createKeepAliveImpl() {
	b.createFunctionStart(true)

	// Get the underlying value of the interface value.
	interfaceValue := b.getValue(b.fn.Params[0])
	pointerValue := b.CreateExtractValue(interfaceValue, 1, "")

	// Create an equivalent of the following C code, which is basically just a
	// nop but ensures the pointerValue is kept alive:
	//
	//     __asm__ __volatile__("" : : "r"(pointerValue))
	//
	// It should be portable to basically everything as the "r" register type
	// exists basically everywhere.
	asmType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.i8ptrType}, false)
	asmFn := llvm.InlineAsm(asmType, "", "r", true, false, 0, false)
	b.createCall(asmType, asmFn, []llvm.Value{pointerValue}, "")

	b.CreateRetVoid()
}

var mathToLLVMMapping = map[string]string{
	"math.Ceil":  "llvm.ceil.f64",
	"math.Exp":   "llvm.exp.f64",
	"math.Exp2":  "llvm.exp2.f64",
	"math.Floor": "llvm.floor.f64",
	"math.Log":   "llvm.log.f64",
	"math.Sqrt":  "llvm.sqrt.f64",
	"math.Trunc": "llvm.trunc.f64",
}

// defineMathOp defines a math function body as a call to a LLVM intrinsic,
// instead of the regular Go implementation. This allows LLVM to reason about
// the math operation and (depending on the architecture) allows it to lower the
// operation to very fast floating point instructions. If this is not possible,
// LLVM will emit a call to a libm function that implements the same operation.
//
// One example of an optimization that LLVM can do is to convert
// float32(math.Sqrt(float64(v))) to a 32-bit floating point operation, which is
// beneficial on architectures where 64-bit floating point operations are (much)
// more expensive than 32-bit ones.
func (b *builder) defineMathOp() {
	b.createFunctionStart(true)
	llvmName := mathToLLVMMapping[b.fn.RelString(nil)]
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
	args := make([]llvm.Value, len(b.fn.Params))
	for i, param := range b.fn.Params {
		args[i] = b.getValue(param)
	}
	result := b.CreateCall(llvmFn.GlobalValueType(), llvmFn, args, "")
	b.CreateRet(result)
}
