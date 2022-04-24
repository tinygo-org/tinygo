package compiler

import (
	"fmt"
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
	case "llvm_AddInt32", "llvm_AddInt64", "llvm_AddUint32", "llvm_AddUint64", "llvm_AddUintptr":
		ptr := b.getValue(call.Args[0])
		val := b.getValue(call.Args[1])
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
			oldVal := b.createCall(fn, []llvm.Value{ptr, val}, "")
			// Return the new value, not the original value returned.
			return b.CreateAdd(oldVal, val, ""), true
		}
		oldVal := b.CreateAtomicRMW(llvm.AtomicRMWBinOpAdd, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		// Return the new value, not the original value returned by atomicrmw.
		return b.CreateAdd(oldVal, val, ""), true
	case "llvm_SwapInt32", "llvm_SwapInt64", "llvm_SwapUint32", "llvm_SwapUint64", "llvm_SwapUintptr", "llvm_SwapPointer":
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
	case "llvm_CompareAndSwapInt32", "llvm_CompareAndSwapInt64", "llvm_CompareAndSwapUint32", "llvm_CompareAndSwapUint64", "llvm_CompareAndSwapUintptr", "llvm_CompareAndSwapPointer":
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
	case "llvm_LoadInt32", "llvm_LoadInt64", "llvm_LoadUint32", "llvm_LoadUint64", "llvm_LoadUintptr", "llvm_LoadPointer":
		ptr := b.getValue(call.Args[0])
		val := b.CreateLoad(ptr, "")
		val.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		val.SetAlignment(b.targetData.PrefTypeAlignment(val.Type())) // required
		return val, true
	case "llvm_StoreInt32", "llvm_StoreInt64", "llvm_StoreUint32", "llvm_StoreUint64", "llvm_StoreUintptr", "llvm_StorePointer":
		ptr := b.getValue(call.Args[0])
		val := b.getValue(call.Args[1])
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
			return b.createCall(fn, []llvm.Value{ptr, val, llvm.ConstInt(b.uintptrType, 5, false)}, ""), true
		}
		store := b.CreateStore(val, ptr)
		store.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		store.SetAlignment(b.targetData.PrefTypeAlignment(val.Type())) // required
		return store, true
	default:
		return llvm.Value{}, false
	}
}
