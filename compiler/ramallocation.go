package compiler

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

func (b *builder) createRAMAllocation(instr *ssa.CallCommon) error {
	// Get the size, which must be a compile-time constant.
	size, ok := instr.Args[0].(*ssa.Const)
	if !ok {
		return b.makeError(instr.Pos(), "Persistent RAM allocation is not a constant")
	}

	// We create a buffer 'constant' in LLVM terms as an array of desired size.  Since
	// the Go runtime code can't know this type (array dimensions), we also have a pointer
	// that points at the allocated buffer.  In order to prevent the symbols from being
	// optimized out, the pointer `_persist_ptr` must be referenced from the runtime
	// code.

	// _persist$buf is the allocated space.
	bufType := llvm.ArrayType(llvm.Int8Type(), int(size.Int64()))
	buf := llvm.AddGlobal(b.mod, bufType, "_persist$buf")
	buf.SetLinkage(llvm.ExternalLinkage)
	buf.SetGlobalConstant(true)
	buf.SetSection(".persist")
	buf.SetUnnamedAddr(true)
	buf.SetInitializer(llvm.ConstNull(bufType))

	// _persist_ptr is the pointer to the allocated space, create if doesn't exist and
	// either way, set initializer to load the address of the allocated space.
	ptrType := b.getLLVMType(types.Typ[types.Uintptr])
	ptr := b.mod.NamedGlobal("_persist_ptr")
	if ptr.IsNil() {
		ptr = llvm.AddGlobal(b.mod, ptrType, "_persist_ptr")
	}
	ptr.SetLinkage(llvm.ExternalLinkage)
	ptr.SetGlobalConstant(true)
	ptr.SetInitializer(llvm.ConstPtrToInt(buf, ptrType))

	return nil
}
