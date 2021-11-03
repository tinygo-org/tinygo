package compiler

import (
	"strings"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createAtomicOp lowers an atomic library call by lowering it as an LLVM atomic
// operation. It returns the result of the operation and true if the call could
// be lowered inline, and false otherwise.
func (b *builder) createAtomicOp(call *ssa.CallCommon) (llvm.Value, bool) {
	name := call.Value.(*ssa.Function).Name()
	switch name {
	case "AddInt32", "AddInt64", "AddUint32", "AddUint64", "AddUintptr":
		ptr := b.getValue(call.Args[0])
		val := b.getValue(call.Args[1])
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpAdd, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		// Return the new value, not the original value returned by atomicrmw.
		return b.CreateAdd(oldVal, val, ""), true
	case "SwapInt32", "SwapInt64", "SwapUint32", "SwapUint64", "SwapUintptr", "SwapPointer":
		ptr := b.getValue(call.Args[0])
		val := b.getValue(call.Args[1])
		isPointer := val.Type().TypeKind() == llvm.PointerTypeKind
		if isPointer {
			// atomicrmw only supports integers, so cast to an integer.
			val = b.CreatePtrToInt(val, b.uintptrType, "")
			ptr = b.CreateBitCast(ptr, llvm.PointerType(val.Type(), 0), "")
		}
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpXchg, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		if isPointer {
			oldVal = b.CreateIntToPtr(oldVal, b.i8ptrType, "")
		}
		return oldVal, true
	case "CompareAndSwapInt32", "CompareAndSwapInt64", "CompareAndSwapUint32", "CompareAndSwapUint64", "CompareAndSwapUintptr", "CompareAndSwapPointer":
		ptr := b.getValue(call.Args[0])
		old := b.getValue(call.Args[1])
		newVal := b.getValue(call.Args[2])
		if strings.HasSuffix(name, "64") {
			if strings.HasPrefix(b.Triple, "thumb") {
				// Work around a bug in LLVM, at least LLVM 11:
				// https://reviews.llvm.org/D95891
				// Check for thumbv6m, thumbv7, thumbv7em, and perhaps others.
				// See also: https://gcc.gnu.org/onlinedocs/gcc/_005f_005fsync-Builtins.html
				compareAndSwap := b.mod.NamedFunction("__sync_val_compare_and_swap_8")
				if compareAndSwap.IsNil() {
					// Declare the function if it isn't already declared.
					i64Type := b.ctx.Int64Type()
					fnType := llvm.FunctionType(i64Type, []llvm.Type{llvm.PointerType(i64Type, 0), i64Type, i64Type}, false)
					compareAndSwap = llvm.AddFunction(b.mod, "__sync_val_compare_and_swap_8", fnType)
				}
				actualOldValue := b.CreateCall(compareAndSwap, []llvm.Value{ptr, old, newVal}, "")
				// The __sync_val_compare_and_swap_8 function returns the old
				// value. However, we shouldn't return the old value, we should
				// return whether the compare/exchange was successful. This is
				// easily done by comparing the returned (actual) old value with
				// the expected old value passed to
				// __sync_val_compare_and_swap_8.
				swapped := b.CreateICmp(llvm.IntEQ, old, actualOldValue, "")
				return swapped, true
			}
		}
		tuple := b.CreateAtomicCmpXchg(ptr, old, newVal, llvm.AtomicOrderingSequentiallyConsistent, llvm.AtomicOrderingSequentiallyConsistent, true)
		swapped := b.CreateExtractValue(tuple, 1, "")
		return swapped, true
	case "LoadInt32", "LoadInt64", "LoadUint32", "LoadUint64", "LoadUintptr", "LoadPointer":
		ptr := b.getValue(call.Args[0])
		val := b.CreateLoad(ptr, "")
		val.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		val.SetAlignment(b.targetData.PrefTypeAlignment(val.Type())) // required
		return val, true
	case "StoreInt32", "StoreInt64", "StoreUint32", "StoreUint64", "StoreUintptr", "StorePointer":
		ptr := b.getValue(call.Args[0])
		val := b.getValue(call.Args[1])
		store := b.CreateStore(val, ptr)
		store.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		store.SetAlignment(b.targetData.PrefTypeAlignment(val.Type())) // required
		return store, true
	default:
		return llvm.Value{}, false
	}
}
