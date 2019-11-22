package compiler

// This file implements volatile loads/stores in runtime/volatile.LoadT and
// runtime/volatile.StoreT as compiler builtins.

import (
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

func (c *Compiler) emitVolatileLoad(frame *Frame, instr *ssa.CallCommon) (llvm.Value, error) {
	addr := frame.getValue(instr.Args[0])
	c.emitNilCheck(frame, addr, "deref")
	val := c.builder.CreateLoad(addr, "")
	val.SetVolatile(true)
	return val, nil
}

func (c *Compiler) emitVolatileStore(frame *Frame, instr *ssa.CallCommon) (llvm.Value, error) {
	addr := frame.getValue(instr.Args[0])
	val := frame.getValue(instr.Args[1])
	c.emitNilCheck(frame, addr, "deref")
	store := c.builder.CreateStore(val, addr)
	store.SetVolatile(true)
	return llvm.Value{}, nil
}
