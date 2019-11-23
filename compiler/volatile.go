package compiler

// This file implements volatile loads/stores in runtime/volatile.LoadT and
// runtime/volatile.StoreT as compiler builtins.

import (
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createVolatileLoad is the implementation of the intrinsic function
// runtime/volatile.LoadT().
func (b *builder) createVolatileLoad(instr *ssa.CallCommon) (llvm.Value, error) {
	addr := b.getValue(instr.Args[0])
	b.createNilCheck(addr, "deref")
	val := b.CreateLoad(addr, "")
	val.SetVolatile(true)
	return val, nil
}

// createVolatileStore is the implementation of the intrinsic function
// runtime/volatile.StoreT().
func (b *builder) createVolatileStore(instr *ssa.CallCommon) (llvm.Value, error) {
	addr := b.getValue(instr.Args[0])
	val := b.getValue(instr.Args[1])
	b.createNilCheck(addr, "deref")
	store := b.CreateStore(val, addr)
	store.SetVolatile(true)
	return llvm.Value{}, nil
}
