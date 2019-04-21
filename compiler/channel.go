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
	valueType := c.getLLVMType(expr.Type().(*types.Chan).Elem())
	if c.targetData.TypeAllocSize(valueType) > c.targetData.TypeAllocSize(c.intType) {
		// Values bigger than int overflow the data part of the coroutine.
		// TODO: make the coroutine data part big enough to hold these bigger
		// values.
		return llvm.Value{}, c.makeError(expr.Pos(), "todo: channel with values bigger than int")
	}
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
	valueType := c.getLLVMType(instr.Chan.Type().(*types.Chan).Elem())
	ch := c.getValue(frame, instr.Chan)
	chanValue := c.getValue(frame, instr.X)
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(chanValue.Type()), false)
	valueAlloca := c.builder.CreateAlloca(valueType, "chan.value")
	c.builder.CreateStore(chanValue, valueAlloca)
	valueAllocaCast := c.builder.CreateBitCast(valueAlloca, c.i8ptrType, "chan.value.i8ptr")
	c.createRuntimeCall("chanSendStub", []llvm.Value{llvm.Undef(c.i8ptrType), ch, valueAllocaCast, valueSize}, "")
}

// emitChanRecv emits a pseudo chan receive operation. It is lowered to the
// actual channel receive operation during goroutine lowering.
func (c *Compiler) emitChanRecv(frame *Frame, unop *ssa.UnOp) llvm.Value {
	valueType := c.getLLVMType(unop.X.Type().(*types.Chan).Elem())
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(valueType), false)
	ch := c.getValue(frame, unop.X)
	valueAlloca := c.builder.CreateAlloca(valueType, "chan.value")
	valueAllocaCast := c.builder.CreateBitCast(valueAlloca, c.i8ptrType, "chan.value.i8ptr")
	valueOk := c.builder.CreateAlloca(c.ctx.Int1Type(), "chan.comma-ok.alloca")
	c.createRuntimeCall("chanRecvStub", []llvm.Value{llvm.Undef(c.i8ptrType), ch, valueAllocaCast, valueOk, valueSize}, "")
	received := c.builder.CreateLoad(valueAlloca, "chan.received")
	if unop.CommaOk {
		commaOk := c.builder.CreateLoad(valueOk, "chan.comma-ok")
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
