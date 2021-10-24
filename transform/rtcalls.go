package transform

// This file implements several small optimizations of runtime and reflect
// calls.

import (
	"strings"

	"tinygo.org/x/go-llvm"
)

// OptimizeStringToBytes transforms runtime.stringToBytes(...) calls into const
// []byte slices whenever possible. This optimizes the following pattern:
//
//     w.Write([]byte("foo"))
//
// where Write does not store to the slice.
func OptimizeStringToBytes(mod llvm.Module) {
	stringToBytes := mod.NamedFunction("runtime.stringToBytes")
	if stringToBytes.IsNil() {
		// nothing to optimize
		return
	}

	for _, call := range getUses(stringToBytes) {
		strptr := call.Operand(0)
		strlen := call.Operand(1)

		// strptr is always constant because strings are always constant.

		convertedAllUses := true
		for _, use := range getUses(call) {
			if use.IsAExtractValueInst().IsNil() {
				// Expected an extractvalue, but this is something else.
				convertedAllUses = false
				continue
			}
			switch use.Type().TypeKind() {
			case llvm.IntegerTypeKind:
				// A length (len or cap). Propagate the length value.
				use.ReplaceAllUsesWith(strlen)
				use.EraseFromParentAsInstruction()
			case llvm.PointerTypeKind:
				// The string pointer itself.
				if !isReadOnly(use) {
					convertedAllUses = false
					continue
				}
				use.ReplaceAllUsesWith(strptr)
				use.EraseFromParentAsInstruction()
			default:
				// should not happen
				panic("unknown return type of runtime.stringToBytes: " + use.Type().String())
			}
		}
		if convertedAllUses {
			// Call to runtime.stringToBytes can be eliminated: both the input
			// and the output is constant.
			call.EraseFromParentAsInstruction()
		}
	}
}

// OptimizeStringEqual transforms runtime.stringEqual(...) calls into simple
// integer comparisons if at least one of the sides of the comparison is zero.
// Ths converts str == "" into len(str) == 0 and "" == "" into false.
func OptimizeStringEqual(mod llvm.Module) {
	stringEqual := mod.NamedFunction("runtime.stringEqual")
	if stringEqual.IsNil() {
		// nothing to optimize
		return
	}

	builder := mod.Context().NewBuilder()
	defer builder.Dispose()

	for _, call := range getUses(stringEqual) {
		str1len := call.Operand(1)
		str2len := call.Operand(3)

		zero := llvm.ConstInt(str1len.Type(), 0, false)
		if str1len == zero || str2len == zero {
			builder.SetInsertPointBefore(call)
			icmp := builder.CreateICmp(llvm.IntEQ, str1len, str2len, "")
			call.ReplaceAllUsesWith(icmp)
			call.EraseFromParentAsInstruction()
			continue
		}
	}
}

// OptimizeReflectImplements optimizes the following code:
//
//     implements := someType.Implements(someInterfaceType)
//
// where someType is an arbitrary reflect.Type and someInterfaceType is a
// reflect.Type of interface kind, to the following code:
//
//     _, implements := someType.(interfaceType)
//
// if the interface type is known at compile time (that is, someInterfaceType is
// a LLVM constant aggregate). This optimization is especially important for the
// encoding/json package, which uses this method.
//
// As of this writing, the (reflect.Type).Interface method has not yet been
// implemented so this optimization is critical for the encoding/json package.
func OptimizeReflectImplements(mod llvm.Module) {
	implementsSignature := mod.NamedGlobal("reflect/methods.Implements(reflect.Type) bool")
	if implementsSignature.IsNil() {
		return
	}

	builder := mod.Context().NewBuilder()
	defer builder.Dispose()

	// Get a few useful object for use later.
	uintptrType := mod.Context().IntType(llvm.NewTargetData(mod.DataLayout()).PointerSize() * 8)

	// Look up the (reflect.Value).Implements() method.
	var implementsFunc llvm.Value
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		attr := fn.GetStringAttributeAtIndex(-1, "tinygo-invoke")
		if attr.IsNil() {
			continue
		}
		if attr.GetStringValue() == "reflect/methods.Implements(reflect.Type) bool" {
			implementsFunc = fn
			break
		}
	}
	if implementsFunc.IsNil() {
		// Doesn't exist in the program, so nothing to do.
		return
	}

	for _, call := range getUses(implementsFunc) {
		if call.IsACallInst().IsNil() {
			continue
		}
		interfaceTypeBitCast := call.Operand(2)
		if interfaceTypeBitCast.IsAConstantExpr().IsNil() || interfaceTypeBitCast.Opcode() != llvm.BitCast {
			// The asserted interface is not constant, so can't optimize this
			// code.
			continue
		}

		interfaceType := interfaceTypeBitCast.Operand(0)
		if strings.HasPrefix(interfaceType.Name(), "reflect/types.type:named:") {
			// Get the underlying type.
			interfaceType = llvm.ConstExtractValue(interfaceType.Initializer(), []uint32{0})
		}
		if !strings.HasPrefix(interfaceType.Name(), "reflect/types.type:interface:") {
			// This is an error. The Type passed to Implements should be of
			// interface type. Ignore it here (don't report it), it will be
			// reported at runtime.
			continue
		}
		if interfaceType.IsAGlobalVariable().IsNil() {
			// Interface is unknown at compile time. This can't be optimized.
			continue
		}
		typeAssertFunction := llvm.ConstExtractValue(interfaceType.Initializer(), []uint32{4}).Operand(0)

		// Replace Implements call with the type assert call.
		builder.SetInsertPointBefore(call)
		implements := builder.CreateCall(typeAssertFunction, []llvm.Value{
			builder.CreatePtrToInt(call.Operand(0), uintptrType, ""), // typecode to check
		}, "")
		call.ReplaceAllUsesWith(implements)
		call.EraseFromParentAsInstruction()
	}
}
