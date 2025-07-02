package compiler

import (
	"tinygo.org/x/go-llvm"
)

// createAtomicOp lowers a sync/atomic function by lowering it as an LLVM atomic
// operation. It returns the result of the operation, or a zero llvm.Value if
// the result is void.
func (b *builder) createAtomicOp(name string) llvm.Value {
	switch name {
	case "AddInt32", "AddInt64", "AddUint32", "AddUint64", "AddUintptr":
		ptr := b.getValue(b.fn.Params[0], getPos(b.fn))
		val := b.getValue(b.fn.Params[1], getPos(b.fn))
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpAdd, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		// Return the new value, not the original value returned by atomicrmw.
		return b.CreateAdd(oldVal, val, "")
	case "AndInt32", "AndInt64", "AndUint32", "AndUint64", "AndUintptr":
		ptr := b.getValue(b.fn.Params[0], getPos(b.fn))
		val := b.getValue(b.fn.Params[1], getPos(b.fn))
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpAnd, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		return oldVal
	case "OrInt32", "OrInt64", "OrUint32", "OrUint64", "OrUintptr":
		ptr := b.getValue(b.fn.Params[0], getPos(b.fn))
		val := b.getValue(b.fn.Params[1], getPos(b.fn))
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpOr, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		return oldVal
	case "SwapInt32", "SwapInt64", "SwapUint32", "SwapUint64", "SwapUintptr", "SwapPointer":
		ptr := b.getValue(b.fn.Params[0], getPos(b.fn))
		val := b.getValue(b.fn.Params[1], getPos(b.fn))
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpXchg, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		return oldVal
	case "CompareAndSwapInt32", "CompareAndSwapInt64", "CompareAndSwapUint32", "CompareAndSwapUint64", "CompareAndSwapUintptr", "CompareAndSwapPointer":
		ptr := b.getValue(b.fn.Params[0], getPos(b.fn))
		old := b.getValue(b.fn.Params[1], getPos(b.fn))
		newVal := b.getValue(b.fn.Params[2], getPos(b.fn))
		tuple := b.CreateAtomicCmpXchg(ptr, old, newVal, llvm.AtomicOrderingSequentiallyConsistent, llvm.AtomicOrderingSequentiallyConsistent, true)
		swapped := b.CreateExtractValue(tuple, 1, "")
		return swapped
	case "LoadInt32", "LoadInt64", "LoadUint32", "LoadUint64", "LoadUintptr", "LoadPointer":
		ptr := b.getValue(b.fn.Params[0], getPos(b.fn))
		val := b.CreateLoad(b.getLLVMType(b.fn.Signature.Results().At(0).Type()), ptr, "")
		val.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		val.SetAlignment(b.targetData.PrefTypeAlignment(val.Type())) // required
		return val
	case "StoreInt32", "StoreInt64", "StoreUint32", "StoreUint64", "StoreUintptr", "StorePointer":
		ptr := b.getValue(b.fn.Params[0], getPos(b.fn))
		val := b.getValue(b.fn.Params[1], getPos(b.fn))
		store := b.CreateStore(val, ptr)
		store.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		store.SetAlignment(b.targetData.PrefTypeAlignment(val.Type())) // required
		return llvm.Value{}
	default:
		b.addError(b.fn.Pos(), "unknown atomic operation: "+b.fn.Name())
		return llvm.Value{}
	}
}
