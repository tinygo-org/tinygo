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
	chanType := c.getLLVMType(c.getRuntimeType("channel"))
	size := c.targetData.TypeAllocSize(chanType)
	sizeValue := llvm.ConstInt(c.uintptrType, size, false)
	ptr := c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "chan.alloc")
	ptr = c.builder.CreateBitCast(ptr, llvm.PointerType(chanType, 0), "chan")
	return ptr, nil
}

// emitChanSend emits a pseudo chan send operation. It is lowered to the actual
// channel send operation during goroutine lowering.
func (c *Compiler) emitChanSend(frame *Frame, instr *ssa.Send) {
	ch := c.getValue(frame, instr.Chan)
	chanValue := c.getValue(frame, instr.X)

	// store value-to-send
	valueType := c.getLLVMType(instr.X.Type())
	valueAlloca, valueAllocaCast, valueAllocaSize := c.createTemporaryAlloca(valueType, "chan.value")
	c.builder.CreateStore(chanValue, valueAlloca)

	// Do the send.
	coroutine := c.createRuntimeCall("getCoroutine", nil, "")
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(chanValue.Type()), false)
	c.createRuntimeCall("chanSend", []llvm.Value{coroutine, ch, valueAllocaCast, valueSize}, "")

	// End the lifetime of the alloca.
	// This also works around a bug in CoroSplit, at least in LLVM 8:
	// https://bugs.llvm.org/show_bug.cgi?id=41742
	c.emitLifetimeEnd(valueAllocaCast, valueAllocaSize)
}

// emitChanRecv emits a pseudo chan receive operation. It is lowered to the
// actual channel receive operation during goroutine lowering.
func (c *Compiler) emitChanRecv(frame *Frame, unop *ssa.UnOp) llvm.Value {
	valueType := c.getLLVMType(unop.X.Type().(*types.Chan).Elem())
	ch := c.getValue(frame, unop.X)

	// Allocate memory to receive into.
	valueAlloca, valueAllocaCast, valueAllocaSize := c.createTemporaryAlloca(valueType, "chan.value")

	// Do the receive.
	coroutine := c.createRuntimeCall("getCoroutine", nil, "")
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(valueType), false)
	c.createRuntimeCall("chanRecv", []llvm.Value{coroutine, ch, valueAllocaCast, valueSize}, "")
	received := c.builder.CreateLoad(valueAlloca, "chan.received")
	c.emitLifetimeEnd(valueAllocaCast, valueAllocaSize)

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
