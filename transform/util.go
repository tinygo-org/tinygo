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

type readOnlyHandling int

const (
	CheckReadOnlyFlag readOnlyHandling = iota
	CheckUseOnly
	CheckAllOperands
)

// Instructions in this map are all read-only.  The boolean value serves
// a purpose inside isReadOnly() to determine how to calculate if a certain
// use of a value is read-only.
var readOnlyOpcodes = map[llvm.Opcode]readOnlyHandling{
	llvm.Load:  CheckUseOnly,
	llvm.Store: CheckAllOperands,

	llvm.Call: CheckReadOnlyFlag,
	llvm.Ret:  CheckUseOnly,

	llvm.ICmp: CheckUseOnly,
	llvm.FCmp: CheckUseOnly,

	llvm.PHI: CheckUseOnly,

	llvm.GetElementPtr:  CheckUseOnly,
	llvm.ExtractValue:   CheckUseOnly,
	llvm.InsertValue:    CheckUseOnly,
	llvm.ExtractElement: CheckUseOnly,

	llvm.Add:  CheckUseOnly,
	llvm.FAdd: CheckUseOnly,
	llvm.Sub:  CheckUseOnly,
	llvm.FSub: CheckUseOnly,
	llvm.Mul:  CheckUseOnly,
	llvm.FMul: CheckUseOnly,
	llvm.UDiv: CheckUseOnly,
	llvm.SDiv: CheckUseOnly,
	llvm.FDiv: CheckUseOnly,
	llvm.URem: CheckUseOnly,
	llvm.SRem: CheckUseOnly,
	llvm.FRem: CheckUseOnly,

	llvm.Shl:  CheckUseOnly,
	llvm.LShr: CheckUseOnly,
	llvm.AShr: CheckUseOnly,
	llvm.And:  CheckUseOnly,
	llvm.Or:   CheckUseOnly,
	llvm.Xor:  CheckUseOnly,

	llvm.ZExt: CheckUseOnly,
	llvm.SExt: CheckUseOnly,

	llvm.PtrToInt: CheckUseOnly,
	llvm.IntToPtr: CheckUseOnly,
	llvm.BitCast:  CheckUseOnly,
}

// isReadOnly returns true if the given value (which must be of pointer type) is
// never stored to, and false if this cannot be proven.
func isReadOnly(value llvm.Value) bool {
	visited := map[llvm.Value]bool{}
	return isReadOnlyInternal(visited, value)
}

func isReadOnlyInternal(visited map[llvm.Value]bool, value llvm.Value) bool {
	for _, use := range getUses(value) {
		if _, ok := visited[value]; ok {
			continue
		}
		visited[value] = true

		handling, ok := readOnlyOpcodes[use.InstructionOpcode()]
		if !ok {
			return false
		}
		switch handling {
		case CheckUseOnly:
			if !isReadOnlyInternal(visited, use) {
				return false
			}
		case CheckReadOnlyFlag:
			if !hasFlag(use, value, "readonly") {
				return false
			}
		case CheckAllOperands:
			for i := 0; i < use.OperandsCount(); i++ {
				if !isReadOnlyInternal(visited, use.Operand(i)) {
					return false
				}
			}
		}
	}

	return true
}
