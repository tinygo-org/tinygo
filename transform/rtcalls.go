package transform

// This file implements several small optimizations of runtime and reflect
// calls.

import (
	"tinygo.org/x/go-llvm"
)

// OptimizeStringToBytes transforms runtime.stringToBytes(...) calls into const
// []byte slices whenever possible. This optimizes the following pattern:
//
//	w.Write([]byte("foo"))
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

		var pointerUses []llvm.Value
		canConvertPointer := true
		for _, use := range getUses(call) {
			if use.IsAExtractValueInst().IsNil() {
				// Expected an extractvalue, but this is something else.
				canConvertPointer = false
				break
			}
			switch use.Type().TypeKind() {
			case llvm.IntegerTypeKind:
				// A length (len or cap). Propagate the length value.
				// This can always be done because the byte slice is always the
				// same length as the original string.
				use.ReplaceAllUsesWith(strlen)
				use.EraseFromParentAsInstruction()
			case llvm.PointerTypeKind:
				// The string pointer itself.
				if !isReadOnly(use) {
					// There is a store to the byte slice. This means that none
					// of the pointer uses can't be propagated.
					canConvertPointer = false
					break
				}
				// It may be that the pointer value can be propagated, if all of
				// the pointer uses are readonly.
				pointerUses = append(pointerUses, use)
			default:
				// should not happen
				panic("unknown return type of runtime.stringToBytes: " + use.Type().String())
			}
		}
		if canConvertPointer {
			// All pointer uses are readonly, so they can be converted.
			for _, use := range pointerUses {
				use.ReplaceAllUsesWith(strptr)
				use.EraseFromParentAsInstruction()
			}

			// Call to runtime.stringToBytes can be eliminated: both the input
			// and the output is constant.
			call.EraseFromParentAsInstruction()
		}
	}
}
