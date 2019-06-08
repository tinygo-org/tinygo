package compiler

// This file lowers channel operations (make/send/recv/close) to runtime calls
// or pseudo-operations that are lowered during goroutine lowering.

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// emitMakeChan returns a new channel value for the given channel type.
func (c *Compiler) emitMakeChan(expr *ssa.MakeChan) (llvm.Value, error) {
	chanType := c.getLLVMType(expr.Type())
	size := c.targetData.TypeAllocSize(chanType.ElementType())
	sizeValue := llvm.ConstInt(c.uintptrType, size, false)
	ptr := c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "chan.alloc")
	ptr = c.builder.CreateBitCast(ptr, chanType, "chan")
	// Set the elementSize field
	elementSizePtr := c.builder.CreateGEP(ptr, []llvm.Value{
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
	}, "")
	elementSize := c.targetData.TypeAllocSize(c.getLLVMType(expr.Type().(*types.Chan).Elem()))
	if elementSize > 0xffff {
		return ptr, c.makeError(expr.Pos(), fmt.Sprintf("element size is %d bytes, which is bigger than the maximum of %d bytes", elementSize, 0xffff))
	}
	elementSizeValue := llvm.ConstInt(c.ctx.Int16Type(), elementSize, false)
	c.builder.CreateStore(elementSizeValue, elementSizePtr)
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
	c.createRuntimeCall("chanSend", []llvm.Value{coroutine, ch, valueAllocaCast}, "")

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
	c.createRuntimeCall("chanRecv", []llvm.Value{coroutine, ch, valueAllocaCast}, "")
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
	ch := c.getValue(frame, param)
	c.createRuntimeCall("chanClose", []llvm.Value{ch}, "")
}
