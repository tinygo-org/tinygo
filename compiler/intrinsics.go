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
		params = append(params, b.getValue(param, getPos(b.fn)))
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
		b.getValue(b.fn.Params[0], getPos(b.fn)),
		llvm.ConstInt(b.ctx.Int8Type(), 0, false),
		b.getValue(b.fn.Params[1], getPos(b.fn)),
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
	interfaceValue := b.getValue(b.fn.Params[0], getPos(b.fn))
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
		args[i] = b.getValue(param, getPos(b.fn))
	}
	result := b.CreateCall(llvmFn.GlobalValueType(), llvmFn, args, "")
	b.CreateRet(result)
}

// Implement most math/bits functions.
//
// This implements all the functions that operate on bits. It does not yet
// implement the arithmetic functions (like bits.Add), which also have LLVM
// intrinsics.
func (b *builder) defineMathBitsIntrinsic() bool {
	if b.fn.Pkg.Pkg.Path() != "math/bits" {
		return false
	}
	name := b.fn.Name()
	switch name {
	case "LeadingZeros", "LeadingZeros8", "LeadingZeros16", "LeadingZeros32", "LeadingZeros64",
		"TrailingZeros", "TrailingZeros8", "TrailingZeros16", "TrailingZeros32", "TrailingZeros64":
		b.createFunctionStart(true)
		param := b.getValue(b.fn.Params[0], b.fn.Pos())
		valueType := param.Type()
		var intrinsicName string
		if strings.HasPrefix(name, "Leading") { // LeadingZeros
			intrinsicName = "llvm.ctlz.i" + strconv.Itoa(valueType.IntTypeWidth())
		} else { // TrailingZeros
			intrinsicName = "llvm.cttz.i" + strconv.Itoa(valueType.IntTypeWidth())
		}
		llvmFn := b.mod.NamedFunction(intrinsicName)
		llvmFnType := llvm.FunctionType(valueType, []llvm.Type{valueType, b.ctx.Int1Type()}, false)
		if llvmFn.IsNil() {
			llvmFn = llvm.AddFunction(b.mod, intrinsicName, llvmFnType)
		}
		result := b.createCall(llvmFnType, llvmFn, []llvm.Value{
			param,
			llvm.ConstInt(b.ctx.Int1Type(), 0, false),
		}, "")
		result = b.createZExtOrTrunc(result, b.intType)
		b.CreateRet(result)
		return true
	case "Len", "Len8", "Len16", "Len32", "Len64":
		// bits.Len can be implemented as:
		//     (unsafe.Sizeof(v) * 8) -  bits.LeadingZeros(n)
		// Not sure why this isn't already done in the standard library, as it
		// is much simpler than a lookup table.
		b.createFunctionStart(true)
		param := b.getValue(b.fn.Params[0], b.fn.Pos())
		valueType := param.Type()
		valueBits := valueType.IntTypeWidth()
		intrinsicName := "llvm.ctlz.i" + strconv.Itoa(valueBits)
		llvmFn := b.mod.NamedFunction(intrinsicName)
		llvmFnType := llvm.FunctionType(valueType, []llvm.Type{valueType, b.ctx.Int1Type()}, false)
		if llvmFn.IsNil() {
			llvmFn = llvm.AddFunction(b.mod, intrinsicName, llvmFnType)
		}
		result := b.createCall(llvmFnType, llvmFn, []llvm.Value{
			param,
			llvm.ConstInt(b.ctx.Int1Type(), 0, false),
		}, "")
		result = b.createZExtOrTrunc(result, b.intType)
		maxLen := llvm.ConstInt(b.intType, uint64(valueBits), false) // number of bits in the value
		result = b.CreateSub(maxLen, result, "")
		b.CreateRet(result)
		return true
	case "OnesCount", "OnesCount8", "OnesCount16", "OnesCount32", "OnesCount64":
		b.createFunctionStart(true)
		param := b.getValue(b.fn.Params[0], b.fn.Pos())
		valueType := param.Type()
		intrinsicName := "llvm.ctpop.i" + strconv.Itoa(valueType.IntTypeWidth())
		llvmFn := b.mod.NamedFunction(intrinsicName)
		llvmFnType := llvm.FunctionType(valueType, []llvm.Type{valueType}, false)
		if llvmFn.IsNil() {
			llvmFn = llvm.AddFunction(b.mod, intrinsicName, llvmFnType)
		}
		result := b.createCall(llvmFnType, llvmFn, []llvm.Value{param}, "")
		result = b.createZExtOrTrunc(result, b.intType)
		b.CreateRet(result)
		return true
	case "Reverse", "Reverse8", "Reverse16", "Reverse32", "Reverse64",
		"ReverseBytes", "ReverseBytes16", "ReverseBytes32", "ReverseBytes64":
		b.createFunctionStart(true)
		param := b.getValue(b.fn.Params[0], b.fn.Pos())
		valueType := param.Type()
		var intrinsicName string
		if strings.HasPrefix(name, "ReverseBytes") {
			intrinsicName = "llvm.bswap.i" + strconv.Itoa(valueType.IntTypeWidth())
		} else { // Reverse
			intrinsicName = "llvm.bitreverse.i" + strconv.Itoa(valueType.IntTypeWidth())
		}
		llvmFn := b.mod.NamedFunction(intrinsicName)
		llvmFnType := llvm.FunctionType(valueType, []llvm.Type{valueType}, false)
		if llvmFn.IsNil() {
			llvmFn = llvm.AddFunction(b.mod, intrinsicName, llvmFnType)
		}
		result := b.createCall(llvmFnType, llvmFn, []llvm.Value{param}, "")
		b.CreateRet(result)
		return true
	case "RotateLeft", "RotateLeft8", "RotateLeft16", "RotateLeft32", "RotateLeft64":
		// Warning: the documentation says these functions must be constant time.
		// I do not think LLVM guarantees this, but there's a good chance LLVM
		// already recognized the rotate instruction so it probably won't get
		// any _worse_ by implementing these rotate functions.
		b.createFunctionStart(true)
		x := b.getValue(b.fn.Params[0], b.fn.Pos())
		k := b.getValue(b.fn.Params[1], b.fn.Pos())
		valueType := x.Type()
		intrinsicName := "llvm.fshl.i" + strconv.Itoa(valueType.IntTypeWidth())
		llvmFn := b.mod.NamedFunction(intrinsicName)
		llvmFnType := llvm.FunctionType(valueType, []llvm.Type{valueType, valueType, valueType}, false)
		if llvmFn.IsNil() {
			llvmFn = llvm.AddFunction(b.mod, intrinsicName, llvmFnType)
		}
		k = b.createZExtOrTrunc(k, valueType)
		result := b.createCall(llvmFnType, llvmFn, []llvm.Value{x, x, k}, "")
		b.CreateRet(result)
		return true
	default:
		return false
	}
}
