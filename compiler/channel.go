package compiler

// This file lowers channel operations (make/send/recv/close) to runtime calls
// or pseudo-operations that are lowered during goroutine lowering.

import (
	"go/types"

	"github.com/aykevl/go-llvm"
	"golang.org/x/tools/go/ssa"
)

// emitMakeChan returns a new channel value for the given channel type.
func (c *Compiler) emitMakeChan(expr *ssa.MakeChan) (llvm.Value, error) {
	valueType, err := c.getLLVMType(expr.Type().(*types.Chan).Elem())
	if err != nil {
		return llvm.Value{}, err
	}
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
func (c *Compiler) emitChanSend(frame *Frame, instr *ssa.Send) error {
	valueType, err := c.getLLVMType(instr.Chan.Type().(*types.Chan).Elem())
	if err != nil {
		return err
	}
	ch, err := c.parseExpr(frame, instr.Chan)
	if err != nil {
		return err
	}
	chanValue, err := c.parseExpr(frame, instr.X)
	if err != nil {
		return err
	}
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(chanValue.Type()), false)
	valueAlloca := c.builder.CreateAlloca(valueType, "chan.value")
	c.builder.CreateStore(chanValue, valueAlloca)
	valueAllocaCast := c.builder.CreateBitCast(valueAlloca, c.i8ptrType, "chan.value.i8ptr")
	c.createRuntimeCall("chanSendStub", []llvm.Value{llvm.Undef(c.i8ptrType), ch, valueAllocaCast, valueSize}, "")
	return nil
}

// emitChanRecv emits a pseudo chan receive operation. It is lowered to the
// actual channel receive operation during goroutine lowering.
func (c *Compiler) emitChanRecv(frame *Frame, unop *ssa.UnOp) (llvm.Value, error) {
	valueType, err := c.getLLVMType(unop.X.Type().(*types.Chan).Elem())
	if err != nil {
		return llvm.Value{}, err
	}
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(valueType), false)
	ch, err := c.parseExpr(frame, unop.X)
	if err != nil {
		return llvm.Value{}, err
	}
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
		return tuple, nil
	} else {
		return received, nil
	}
}

// emitChanClose closes the given channel.
func (c *Compiler) emitChanClose(frame *Frame, param ssa.Value) error {
	valueType, err := c.getLLVMType(param.Type().(*types.Chan).Elem())
	valueSize := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(valueType), false)
	if err != nil {
		return err
	}
	ch, err := c.parseExpr(frame, param)
	if err != nil {
		return err
	}
	c.createRuntimeCall("chanClose", []llvm.Value{ch, valueSize}, "")
	return nil
}
