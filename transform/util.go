package transform

// This file contains utilities used across transforms.

import (
	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

// Check whether all uses of this param as parameter to the call have the given
// flag. In most cases, there will only be one use but a function could take the
// same parameter twice, in which case both must have the flag.
// A flag can be any enum flag, like "readonly".
func hasFlag(call, param llvm.Value, kind string) bool {
	fn := call.CalledValue()
	if fn.IsAFunction().IsNil() {
		// This is not a function but something else, like a function pointer.
		return false
	}
	kindID := llvm.AttributeKindID(kind)
	for i := 0; i < fn.ParamsCount(); i++ {
		if call.Operand(i) != param {
			// This is not the parameter we're checking.
			continue
		}
		index := i + 1 // param attributes start at 1
		attr := fn.GetEnumAttributeAtIndex(index, kindID)
		if attr.IsNil() {
			// At least one parameter doesn't have the flag (there may be
			// multiple).
			return false
		}
	}
	return true
}

// isReadOnly returns true if the given value (which must be of pointer type) is
// never stored to, and false if this cannot be proven.
func isReadOnly(value llvm.Value) bool {
	return isReadOnlyDepth(value, 0, map[llvm.Value]struct{}{})
}

// maxReadOnlyDepth limits how many calls deep isReadOnly looks.
const maxReadOnlyDepth = 4

func isReadOnlyDepth(value llvm.Value, depth int, visited map[llvm.Value]struct{}) bool {
	if _, ok := visited[value]; ok {
		// Already being checked, for example a loop through a PHI.
		return true
	}
	visited[value] = struct{}{}
	for use := range llvmutil.Uses(value) {
		switch {
		case !use.IsAGetElementPtrInst().IsNil(), !use.IsAPHINode().IsNil(),
			!use.IsASelectInst().IsNil(), !use.IsABitCastInst().IsNil(),
			!use.IsAInsertValueInst().IsNil():
			if !isReadOnlyDepth(use, depth, visited) {
				return false
			}
		case !use.IsAExtractValueInst().IsNil():
			// Extracting a length or capacity does not carry the pointer.
			if typeHasPointers(use.Type()) && !isReadOnlyDepth(use, depth, visited) {
				return false
			}
		case !use.IsACallInst().IsNil():
			if !hasFlag(use, value, "readonly") && !callParamIsReadOnly(use, value, depth) {
				return false
			}
		case !use.IsALoadInst().IsNil():
			// Loads are read-only.
		default:
			// Unknown instruction, might not be readonly.
			return false
		}
	}
	return true
}

// callParamIsReadOnly reports whether the callee only reads through the
// parameter that receives value, even if LLVM did not mark it 'readonly'.
// LLVM gives up on slices that are carried through a loop as an aggregate
// PHI (b = b[8:]), so look at the callee's own uses of the parameter.
func callParamIsReadOnly(call, value llvm.Value, depth int) bool {
	callee := call.CalledValue()
	if callee.IsAFunction().IsNil() || callee.IsDeclaration() || depth >= maxReadOnlyDepth {
		return false
	}
	if call.OperandsCount()-1 != callee.ParamsCount() {
		return false
	}
	for i := 0; i < callee.ParamsCount(); i++ {
		if call.Operand(i) == value && !isReadOnlyDepth(callee.Param(i), depth+1, map[llvm.Value]struct{}{}) {
			return false
		}
	}
	return true
}
