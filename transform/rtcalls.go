package transform

// This file implements several small optimizations of runtime and reflect
// calls.

import "tinygo.org/x/go-llvm"

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

// OptimizeStringFromBytes transforms temporary strings created from []byte
// slices into direct uses of the slice data when no instruction between the
// conversion and use can mutate the slice.
func OptimizeStringFromBytes(mod llvm.Module) {
	stringFromBytes := mod.NamedFunction("runtime.stringFromBytes")
	if stringFromBytes.IsNil() {
		// nothing to optimize
		return
	}

	stringEqual := mod.NamedFunction("runtime.stringEqual")
	if stringEqual.IsNil() {
		// nothing to optimize
		return
	}

	// String comparisons only read their operands, so they are safe between a
	// conversion and another supported use of that conversion.
	safeCalls := map[llvm.Value]struct{}{}
	for _, call := range getUses(stringEqual) {
		safeCalls[call] = struct{}{}
	}

	// Rewrite each supported use independently, and remove the conversion only
	// when no unconverted uses remain.
	for _, call := range getUses(stringFromBytes) {
		for _, extract := range getUses(call) {
			if extract.IsAExtractValueInst().IsNil() {
				continue
			}
			indices := extract.Indices()
			if len(indices) != 1 || indices[0] != 0 {
				continue
			}
			for _, use := range getUses(extract) {
				if _, ok := safeCalls[use]; !ok {
					continue
				}
				if !isSafeStringFromBytesUse(call, stringFromBytes, use, safeCalls) {
					continue
				}
				replaceStringFromBytesCompareUse(use, extract, call, stringFromBytes)
			}
		}
		removeDeadStringFromBytes(call)
	}
}

func replaceStringFromBytesCompareUse(compare, ptrExtract, call, stringFromBytes llvm.Value) {
	for _, pair := range [][2]int{{0, 1}, {2, 3}} {
		if compare.Operand(pair[0]) != ptrExtract {
			continue
		}
		lenExtract, ok := getStringFromBytesExtract(compare.Operand(pair[1]), stringFromBytes, 1)
		if !ok || lenExtract.Operand(0) != call {
			continue
		}
		compare.SetOperand(pair[0], call.Operand(0))
		compare.SetOperand(pair[1], call.Operand(1))
	}
}

func getStringFromBytesExtract(value, stringFromBytes llvm.Value, index uint64) (llvm.Value, bool) {
	if value.IsAExtractValueInst().IsNil() {
		return llvm.Value{}, false
	}
	indices := value.Indices()
	if len(indices) != 1 || indices[0] != uint32(index) {
		return llvm.Value{}, false
	}
	call := value.Operand(0)
	if call.IsACallInst().IsNil() {
		return llvm.Value{}, false
	}
	called := call.CalledValue()
	if called.IsNil() || called != stringFromBytes {
		return llvm.Value{}, false
	}
	return value, true
}

func isStringFromBytesCall(value, stringFromBytes llvm.Value) bool {
	if value.IsACallInst().IsNil() {
		return false
	}
	called := value.CalledValue()
	return !called.IsNil() && called == stringFromBytes
}

// isSafeStringFromBytesUse reports whether replacing the copied string with the
// source slice preserves the bytes observed by use.
func isSafeStringFromBytesUse(call, stringFromBytes, use llvm.Value, allowedCalls map[llvm.Value]struct{}) bool {
	if call.InstructionParent() != use.InstructionParent() {
		return false
	}
	for inst := llvm.NextInstruction(call); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		if inst == use {
			return true
		}
		if !isSafeStringFromBytesInterveningInstruction(inst, stringFromBytes, allowedCalls) {
			return false
		}
	}
	return false
}

func isSafeStringFromBytesInterveningInstruction(inst, stringFromBytes llvm.Value, allowedCalls map[llvm.Value]struct{}) bool {
	if _, ok := allowedCalls[inst]; ok {
		return true
	}
	switch {
	case !inst.IsAExtractValueInst().IsNil():
		return true
	case isTrackPointerCall(inst):
		return true
	case isStringFromBytesCall(inst, stringFromBytes):
		return true
	default:
		return false
	}
}

func removeDeadStringFromBytes(call llvm.Value) {
	for _, use := range getUses(call) {
		if use.IsAExtractValueInst().IsNil() {
			return
		}
		for _, extractUse := range getUses(use) {
			if !isTrackPointerCall(extractUse) {
				return
			}
		}
	}
	for _, use := range getUses(call) {
		for _, extractUse := range getUses(use) {
			extractUse.EraseFromParentAsInstruction()
		}
		use.EraseFromParentAsInstruction()
	}
	if !hasUses(call) {
		call.EraseFromParentAsInstruction()
	}
}

func isTrackPointerCall(value llvm.Value) bool {
	if value.IsACallInst().IsNil() {
		return false
	}
	called := value.CalledValue()
	return !called.IsNil() && called.Name() == "runtime.trackPointer"
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
