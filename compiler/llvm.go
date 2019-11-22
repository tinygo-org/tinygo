package compiler

import (
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// This file contains helper functions for LLVM that are not exposed in the Go
// bindings.

// Return a list of values (actually, instructions) where this value is used as
// an operand.
func getUses(value llvm.Value) []llvm.Value {
	if value.IsNil() {
		return nil
	}
	var uses []llvm.Value
	use := value.FirstUse()
	for !use.IsNil() {
		uses = append(uses, use.User())
		use = use.NextUse()
	}
	return uses
}

// createTemporaryAlloca creates a new alloca in the entry block and adds
// lifetime start infromation in the IR signalling that the alloca won't be used
// before this point.
//
// This is useful for creating temporary allocas for intrinsics. Don't forget to
// end the lifetime using emitLifetimeEnd after you're done with it.
func (c *Compiler) createTemporaryAlloca(t llvm.Type, name string) (alloca, bitcast, size llvm.Value) {
	return llvmutil.CreateTemporaryAlloca(c.builder, c.mod, t, name)
}

// emitLifetimeEnd signals the end of an (alloca) lifetime by calling the
// llvm.lifetime.end intrinsic. It is commonly used together with
// createTemporaryAlloca.
func (c *Compiler) emitLifetimeEnd(ptr, size llvm.Value) {
	llvmutil.EmitLifetimeEnd(c.builder, c.mod, ptr, size)
}

// emitPointerPack packs the list of values into a single pointer value using
// bitcasts, or else allocates a value on the heap if it cannot be packed in the
// pointer value directly. It returns the pointer with the packed data.
func (c *Compiler) emitPointerPack(values []llvm.Value) llvm.Value {
	return llvmutil.EmitPointerPack(c.builder, c.mod, c.Config, values)
}

// emitPointerUnpack extracts a list of values packed using emitPointerPack.
func (c *Compiler) emitPointerUnpack(ptr llvm.Value, valueTypes []llvm.Type) []llvm.Value {
	return llvmutil.EmitPointerUnpack(c.builder, c.mod, ptr, valueTypes)
}

// emitPointerUnpack extracts a list of values packed using emitPointerPack.
func (b *builder) emitPointerUnpack(ptr llvm.Value, valueTypes []llvm.Type) []llvm.Value {
	return llvmutil.EmitPointerUnpack(b.Builder, b.mod, ptr, valueTypes)
}

// makeGlobalArray creates a new LLVM global with the given name and integers as
// contents, and returns the global.
// Note that it is left with the default linkage etc., you should set
// linkage/constant/etc properties yourself.
func (c *compilerContext) makeGlobalArray(buf []byte, name string, elementType llvm.Type) llvm.Value {
	globalType := llvm.ArrayType(elementType, len(buf))
	global := llvm.AddGlobal(c.mod, globalType, name)
	value := llvm.Undef(globalType)
	for i := 0; i < len(buf); i++ {
		ch := uint64(buf[i])
		value = llvm.ConstInsertValue(value, llvm.ConstInt(elementType, ch, false), []uint32{uint32(i)})
	}
	global.SetInitializer(value)
	return global
}
