package transform

// This file contains utilities used across transforms.

import (
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
