package compiler

import (
	"fmt"
	"strings"

	"tinygo.org/x/go-llvm"
)

// createAtomicOp lowers a sync/atomic function by lowering it as an LLVM atomic
// operation. It returns the result of the operation, or a zero llvm.Value if
// the result is void.
func (b *builder) createAtomicOp(name string) llvm.Value {
	switch name {
	case "AddInt32", "AddInt64", "AddUint32", "AddUint64", "AddUintptr":
		ptr := b.getValue(b.fn.Params[0])
		val := b.getValue(b.fn.Params[1])
		if strings.HasPrefix(b.Triple, "avr") {
			// AtomicRMW does not work on AVR as intended:
			// - There are some register allocation issues (fixed by https://reviews.llvm.org/D97127 which is not yet in a usable LLVM release)
			// - The result is the new value instead of the old value
			vType := val.Type()
			name := fmt.Sprintf("__sync_fetch_and_add_%d", vType.IntTypeWidth()/8)
			fn := b.mod.NamedFunction(name)
			if fn.IsNil() {
				fn = llvm.AddFunction(b.mod, name, llvm.FunctionType(vType, []llvm.Type{ptr.Type(), vType}, false))
			}
			oldVal := b.createCall(fn.GlobalValueType(), fn, []llvm.Value{ptr, val}, "")
			// Return the new value, not the original value returned.
			return b.CreateAdd(oldVal, val, "")
		}
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpAdd, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		// Return the new value, not the original value returned by atomicrmw.
		return b.CreateAdd(oldVal, val, "")
	case "SwapInt32", "SwapInt64", "SwapUint32", "SwapUint64", "SwapUintptr", "SwapPointer":
		ptr := b.getValue(b.fn.Params[0])
		val := b.getValue(b.fn.Params[1])
		isPointer := val.Type().TypeKind() == llvm.PointerTypeKind
		if isPointer {
			// atomicrmw only supports integers, so cast to an integer.
			// TODO: this is fixed in LLVM 15.
			val = b.CreatePtrToInt(val, b.uintptrType, "")
			ptr = b.CreateBitCast(ptr, llvm.PointerType(val.Type(), 0), "")
		}
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpXchg, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		if isPointer {
			oldVal = b.CreateIntToPtr(oldVal, b.i8ptrType, "")
		}
		return oldVal
	case "CompareAndSwapInt32", "CompareAndSwapInt64", "CompareAndSwapUint32", "CompareAndSwapUint64", "CompareAndSwapUintptr", "CompareAndSwapPointer":
		ptr := b.getValue(b.fn.Params[0])
		old := b.getValue(b.fn.Params[1])
		newVal := b.getValue(b.fn.Params[2])
		tuple := b.CreateAtomicCmpXchg(ptr, old, newVal, llvm.AtomicOrderingSequentiallyConsistent, llvm.AtomicOrderingSequentiallyConsistent, true)
		swapped := b.CreateExtractValue(tuple, 1, "")
		return swapped
	case "LoadInt32", "LoadInt64", "LoadUint32", "LoadUint64", "LoadUintptr", "LoadPointer":
		ptr := b.getValue(b.fn.Params[0])
		val := b.CreateLoad(ptr, "")
		val.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		val.SetAlignment(b.targetData.PrefTypeAlignment(val.Type())) // required
		return val
	case "StoreInt32", "StoreInt64", "StoreUint32", "StoreUint64", "StoreUintptr", "StorePointer":
		ptr := b.getValue(b.fn.Params[0])
		val := b.getValue(b.fn.Params[1])
		if strings.HasPrefix(b.Triple, "avr") {
			// SelectionDAGBuilder is currently missing the "are unaligned atomics allowed" check for stores.
			vType := val.Type()
			isPointer := vType.TypeKind() == llvm.PointerTypeKind
			if isPointer {
				// libcalls only supports integers, so cast to an integer.
				vType = b.uintptrType
				val = b.CreatePtrToInt(val, vType, "")
				ptr = b.CreateBitCast(ptr, llvm.PointerType(vType, 0), "")
			}
			name := fmt.Sprintf("__atomic_store_%d", vType.IntTypeWidth()/8)
			fn := b.mod.NamedFunction(name)
			if fn.IsNil() {
				fn = llvm.AddFunction(b.mod, name, llvm.FunctionType(vType, []llvm.Type{ptr.Type(), vType, b.uintptrType}, false))
			}
			b.createCall(fn.GlobalValueType(), fn, []llvm.Value{ptr, val, llvm.ConstInt(b.uintptrType, 5, false)}, "")
			return llvm.Value{}
		}
		store := b.CreateStore(val, ptr)
		store.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		store.SetAlignment(b.targetData.PrefTypeAlignment(val.Type())) // required
		return llvm.Value{}
	default:
		b.addError(b.fn.Pos(), "unknown atomic operation: "+b.fn.Name())
		return llvm.Value{}
	}
}
