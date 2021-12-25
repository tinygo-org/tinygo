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

// createMemoryZeroCall creates calls to llvm.memset.* to zero a block of
// memory, declaring the function if needed. These calls will be lowered to
// regular libc memset calls if they aren't optimized out in a different way.
func (b *builder) createMemoryZeroCall(args []ssa.Value) (llvm.Value, error) {
	fnName := "llvm.memset.p0i8.i" + strconv.Itoa(b.uintptrType.IntTypeWidth())
	llvmFn := b.mod.NamedFunction(fnName)
	if llvmFn.IsNil() {
		fnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.i8ptrType, b.ctx.Int8Type(), b.uintptrType, b.ctx.Int1Type()}, false)
		llvmFn = llvm.AddFunction(b.mod, fnName, fnType)
	}
	params := []llvm.Value{
		b.getValue(args[0]),
		llvm.ConstInt(b.ctx.Int8Type(), 0, false),
		b.getValue(args[1]),
		llvm.ConstInt(b.ctx.Int1Type(), 0, false),
	}
	b.CreateCall(llvmFn, params, "")
	return llvm.Value{}, nil
}

var mathToLLVMMapping = map[string]string{
	"math.Sqrt":  "llvm.sqrt.f64",
	"math.Floor": "llvm.floor.f64",
	"math.Ceil":  "llvm.ceil.f64",
	"math.Trunc": "llvm.trunc.f64",
}

// createMathOp tries to lower the given call as a LLVM math intrinsic, if
// possible. It returns the call result if possible, and a boolean whether it
// succeeded. If it doesn't succeed, the architecture doesn't support the given
// intrinsic.
func (b *builder) createMathOp(call *ssa.CallCommon) (llvm.Value, bool) {
	// Check whether this intrinsic is supported on the given GOARCH.
	// If it is unsupported, this can have two reasons:
	//
	//  1. LLVM can expand the intrinsic inline (using float instructions), but
	//     the result doesn't pass the tests of the math package.
	//  2. LLVM cannot expand the intrinsic inline, will therefore lower it as a
	//     libm function call, but the libm function call also fails the math
	//     package tests.
	//
	// Whatever the implementation, it must pass the tests in the math package
	// so unfortunately only the below intrinsic+architecture combinations are
	// supported.
	name := call.StaticCallee().RelString(nil)
	switch name {
	case "math.Ceil", "math.Floor", "math.Trunc":
		if b.GOARCH != "wasm" && b.GOARCH != "arm64" {
			return llvm.Value{}, false
		}
	case "math.Sqrt":
		if b.GOARCH != "wasm" && b.GOARCH != "amd64" && b.GOARCH != "386" {
			return llvm.Value{}, false
		}
	default:
		return llvm.Value{}, false // only the above functions are supported.
	}

	llvmFn := b.mod.NamedFunction(mathToLLVMMapping[name])
	if llvmFn.IsNil() {
		// The intrinsic doesn't exist yet, so declare it.
		// At the moment, all supported intrinsics have the form "double
		// foo(double %x)" so we can hardcode the signature here.
		llvmType := llvm.FunctionType(b.ctx.DoubleType(), []llvm.Type{b.ctx.DoubleType()}, false)
		llvmFn = llvm.AddFunction(b.mod, mathToLLVMMapping[name], llvmType)
	}
	// Create a call to the intrinsic.
	args := make([]llvm.Value, len(call.Args))
	for i, arg := range call.Args {
		args[i] = b.getValue(arg)
	}
	return b.CreateCall(llvmFn, args, ""), true
}

// fillCompDef fills the body of a compiler-defined function.
func (b *builder) fillCompDef(cd compdef) {
	// Create a basic block.
	b.SetInsertPointAtEnd(b.ctx.AddBasicBlock(b.llvmFn, ""))

	// Evaluate the function on the params.
	results := cd(b, b.llvmFn.Params()...)
	if len(results) == 0 {
		b.CreateRetVoid()
		return
	}

	// Pack the result and return it.
	var result llvm.Value
	if len(results) == 1 {
		result = results[0]
	} else {
		tuple := llvm.Undef(b.llvmFn.Type().ElementType().ReturnType())
		for i, res := range results {
			tuple = b.CreateInsertValue(tuple, res, i, "insert."+res.Name())
		}
		result = tuple
	}
	b.CreateRet(result)

	// Add LLVM inline hint to the function.
	inline := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("inlinehint"), 0)
	b.llvmFn.AddFunctionAttr(inline)
}

// compDefs are compiler-defined functions.
// These functions use types or operations which cannot be constructed inside TinyGo code.
var compDefs = map[string]compdef{
	// used by math/bits.Mul64 (does 128-bit math)
	"tinygo.math.mul64": (*builder).createMul64,

	// used by math/bits.Div64 and math/bits.Rem64 (does 128-bit math)
	"tinygo.math.div64": (*builder).createDiv64,
}

type compdef func(*builder, ...llvm.Value) []llvm.Value

func (b *builder) createMul64(in ...llvm.Value) []llvm.Value {
	i64 := b.ctx.Int64Type()
	i128 := b.ctx.IntType(128)

	// Zero-extend the arguments.
	x := b.CreateZExt(in[0], i128, "x.zext")
	y := b.CreateZExt(in[1], i128, "y.zext")

	// Compute the full 128-bit product.
	product := b.CreateMul(x, y, "product")

	// Seperate the high and low halves.
	hi := b.CreateLShr(product, llvm.ConstInt(i128, 64, false), "hi.shift")
	hi = b.CreateTrunc(hi, i64, "hi")
	lo := b.CreateTrunc(product, i64, "lo")

	return []llvm.Value{hi, lo}
}

func (b *builder) createDiv64(in ...llvm.Value) []llvm.Value {
	i64 := b.ctx.Int64Type()
	i128 := b.ctx.IntType(128)

	// Merge the high and low portions of the dividend.
	dividend := b.CreateOr(
		b.CreateShl(
			b.CreateZExt(in[0], i128, "hi.zext"),
			llvm.ConstInt(i128, 64, false),
			"hi.shift"),
		b.CreateZExt(in[1], i128, "lo.zext"),
		"dividend")

	// Zero-extend the divisor.
	divisor := b.CreateZExt(in[2], i128, "y.zext")

	// Compute the full 128-bit quotient and remainder.
	quotient := b.CreateUDiv(dividend, divisor, "quotient.128")
	remainder := b.CreateURem(dividend, divisor, "remainder.128")

	// Truncate the quotient and remainder.
	quotient = b.CreateTrunc(quotient, i64, "quotient")
	remainder = b.CreateTrunc(remainder, i64, "remainder")

	return []llvm.Value{quotient, remainder}
}
