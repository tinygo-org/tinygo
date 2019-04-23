package compiler

import (
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

func (c *Compiler) emitAtomic(frame *Frame, call *ssa.CallCommon) (llvm.Value, error) {
	name := call.Value.(*ssa.Function).Name()
	switch name {
	case "AddInt32", "AddInt64", "AddUint32", "AddUint64", "AddUintptr":
		ptr, err := c.parseExpr(frame, call.Args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		val, err := c.parseExpr(frame, call.Args[1])
		if err != nil {
			return llvm.Value{}, err
		}
		if c.GOARCH == "wasm" {
			ptrVal := c.builder.CreateLoad(ptr, "")
			ptrVal = c.builder.CreateAdd(ptrVal, val, "")
			c.builder.CreateStore(ptrVal, ptr)
			return ptrVal, nil
		}
		oldVal := c.builder.CreateAtomicRMW(llvm.AtomicRMWBinOpAdd, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		// Return the new value, not the original value returned by atomicrmw.
		return c.builder.CreateAdd(oldVal, val, ""), nil
	case "SwapInt32", "SwapInt64", "SwapUint32", "SwapUint64", "SwapUintptr", "SwapPointer":
		ptr, err := c.parseExpr(frame, call.Args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		val, err := c.parseExpr(frame, call.Args[1])
		if err != nil {
			return llvm.Value{}, err
		}
		if c.GOARCH == "wasm" {
			old := c.builder.CreateLoad(ptr, "")
			c.builder.CreateStore(val, ptr)
			return old, nil
		}
		isPointer := val.Type().TypeKind() == llvm.PointerTypeKind
		if isPointer {
			// atomicrmw only supports integers, so cast to an integer.
			val = c.builder.CreatePtrToInt(val, c.uintptrType, "")
			ptr = c.builder.CreateBitCast(ptr, llvm.PointerType(val.Type(), 0), "")
		}
		oldVal := c.builder.CreateAtomicRMW(llvm.AtomicRMWBinOpXchg, ptr, val, llvm.AtomicOrderingSequentiallyConsistent, true)
		if isPointer {
			oldVal = c.builder.CreateIntToPtr(oldVal, c.i8ptrType, "")
		}
		return oldVal, nil
	case "CompareAndSwapInt32", "CompareAndSwapInt64", "CompareAndSwapUint32", "CompareAndSwapUint64", "CompareAndSwapUintptr", "CompareAndSwapPointer":
		ptr, err := c.parseExpr(frame, call.Args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		old, err := c.parseExpr(frame, call.Args[1])
		if err != nil {
			return llvm.Value{}, err
		}
		newVal, err := c.parseExpr(frame, call.Args[2])
		if err != nil {
			return llvm.Value{}, err
		}
		if c.GOARCH == "wasm" {
			swapBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, "cas.swap")
			nextBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, "cas.next")
			frame.blockExits[frame.currentBlock] = nextBlock
			val := c.builder.CreateLoad(ptr, "")
			swapped := c.builder.CreateICmp(llvm.IntEQ, val, old, "")
			c.builder.CreateCondBr(swapped, swapBlock, nextBlock)
			c.builder.SetInsertPointAtEnd(swapBlock)
			c.builder.CreateStore(newVal, ptr)
			c.builder.CreateBr(nextBlock)
			c.builder.SetInsertPointAtEnd(nextBlock)
			return swapped, nil
		}
		tuple := c.builder.CreateAtomicCmpXchg(ptr, old, newVal, llvm.AtomicOrderingSequentiallyConsistent, llvm.AtomicOrderingSequentiallyConsistent, true)
		swapped := c.builder.CreateExtractValue(tuple, 1, "")
		return swapped, nil
	case "LoadInt32", "LoadInt64", "LoadUint32", "LoadUint64", "LoadUintptr", "LoadPointer":
		ptr, err := c.parseExpr(frame, call.Args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		val := c.builder.CreateLoad(ptr, "")
		if c.GOARCH == "wasm" {
			return val, nil
		}
		val.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		val.SetAlignment(c.targetData.PrefTypeAlignment(val.Type())) // required
		return val, nil
	case "StoreInt32", "StoreInt64", "StoreUint32", "StoreUint64", "StoreUintptr", "StorePointer":
		ptr, err := c.parseExpr(frame, call.Args[0])
		if err != nil {
			return llvm.Value{}, err
		}
		val, err := c.parseExpr(frame, call.Args[1])
		if err != nil {
			return llvm.Value{}, err
		}
		store := c.builder.CreateStore(val, ptr)
		if c.GOARCH == "wasm" {
			return store, nil
		}
		store.SetOrdering(llvm.AtomicOrderingSequentiallyConsistent)
		store.SetAlignment(c.targetData.PrefTypeAlignment(val.Type())) // required
		return store, nil
	default:
		return llvm.Value{}, c.makeError(call.Pos(), "unknown atomic call: "+name)
	}
}
