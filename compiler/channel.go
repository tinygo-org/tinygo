package compiler

// This file lowers channel operations (make/send/recv/close) to runtime calls
// or pseudo-operations that are lowered during goroutine lowering.

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// emitMakeChan returns a new channel value for the given channel type.
func (c *Compiler) emitMakeChan(expr *ssa.MakeChan) (llvm.Value, error) {
	chanType := c.mod.GetTypeByName("runtime.channel")
	size := c.targetData.TypeAllocSize(chanType)
	sizeValue := llvm.ConstInt(c.uintptrType, size, false)
	ptr := c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "chan.alloc")
	ptr = c.builder.CreateBitCast(ptr, llvm.PointerType(chanType, 0), "chan")
	return ptr, nil
}

// emitChanSend emits a pseudo chan send operation. It is lowered to the actual
// channel send operation during goroutine lowering.
func (c *Compiler) emitChanSend(frame *Frame, instr *ssa.Send) {
	valueType := c.getLLVMType(instr.X.Type())
	ch := c.getValue(frame, instr.Chan)
	chanValue := c.getValue(frame, instr.X)
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(chanValue.Type()), false)
	coroutine := c.createRuntimeCall("getCoroutine", nil, "")

	// store value-to-send
	c.builder.SetInsertPointBefore(coroutine.InstructionParent().Parent().EntryBasicBlock().FirstInstruction())
	valueAlloca := c.builder.CreateAlloca(valueType, "chan.value")
	c.builder.SetInsertPointBefore(coroutine)
	c.builder.SetInsertPointAtEnd(coroutine.InstructionParent())
	c.builder.CreateStore(chanValue, valueAlloca)
	valueAllocaCast := c.builder.CreateBitCast(valueAlloca, c.i8ptrType, "chan.value.i8ptr")

	// Do the send.
	c.createRuntimeCall("chanSend", []llvm.Value{coroutine, ch, valueAllocaCast, valueSize}, "")

	// Make sure CoroSplit includes the alloca in the coroutine frame.
	// This is a bit dirty, but it works (at least in LLVM 8).
	valueSizeI64 := llvm.ConstInt(c.ctx.Int64Type(), c.targetData.TypeAllocSize(chanValue.Type()), false)
	c.builder.CreateCall(c.getLifetimeEndFunc(), []llvm.Value{valueSizeI64, valueAllocaCast}, "")
}

// emitChanRecv emits a pseudo chan receive operation. It is lowered to the
// actual channel receive operation during goroutine lowering.
func (c *Compiler) emitChanRecv(frame *Frame, unop *ssa.UnOp) llvm.Value {
	valueType := c.getLLVMType(unop.X.Type().(*types.Chan).Elem())
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(valueType), false)
	ch := c.getValue(frame, unop.X)
	coroutine := c.createRuntimeCall("getCoroutine", nil, "")

	// Allocate memory to receive into.
	c.builder.SetInsertPointBefore(coroutine.InstructionParent().Parent().EntryBasicBlock().FirstInstruction())
	valueAlloca := c.builder.CreateAlloca(valueType, "chan.value")
	c.builder.SetInsertPointBefore(coroutine)
	c.builder.SetInsertPointAtEnd(coroutine.InstructionParent())
	valueAllocaCast := c.builder.CreateBitCast(valueAlloca, c.i8ptrType, "chan.value.i8ptr")

	// Do the receive.
	c.createRuntimeCall("chanRecv", []llvm.Value{coroutine, ch, valueAllocaCast, valueSize}, "")
	received := c.builder.CreateLoad(valueAlloca, "chan.received")
	if unop.CommaOk {
		commaOk := c.createRuntimeCall("getTaskPromiseData", []llvm.Value{coroutine}, "chan.commaOk.wide")
		commaOk = c.builder.CreateTrunc(commaOk, c.ctx.Int1Type(), "chan.commaOk")
		tuple := llvm.Undef(c.ctx.StructType([]llvm.Type{valueType, c.ctx.Int1Type()}, false))
		tuple = c.builder.CreateInsertValue(tuple, received, 0, "")
		tuple = c.builder.CreateInsertValue(tuple, commaOk, 1, "")
		return tuple
	} else {
		return received
	}
}

// emitChanClose closes the given channel.
func (c *Compiler) emitChanClose(frame *Frame, param ssa.Value) {
	valueType := c.getLLVMType(param.Type().(*types.Chan).Elem())
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(valueType), false)
	ch := c.getValue(frame, param)
	c.createRuntimeCall("chanClose", []llvm.Value{ch, valueSize}, "")
}
