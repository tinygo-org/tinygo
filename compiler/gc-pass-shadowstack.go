package compiler

// This file implements a compiler pass to move GC pointers to a "shadow stack"
// that can easily be scanned by a garbage collector, even without platform
// support.
// For more information, see:
//    https://llvm.org/docs/GarbageCollection.html#the-shadow-stack-gc

import (
	"github.com/aykevl/go-llvm"
)

// AddGCRoots moves pointer values to shadow stack frames when this function (or
// any function it calls) may allocate something. This allows the GC to scan the
// stack in a highly portable way.
func (c *Compiler) AddGCRoots() {
	alloc := c.mod.NamedFunction("runtime.alloc")
	if alloc.IsNil() {
		return
	}

	// Find all functions that do memory allocation.
	worklist := []llvm.Value{alloc}
	allocSet := make(map[llvm.Value]struct{})
	allocList := make([]llvm.Value, 0, 4)
	for len(worklist) != 0 {
		// Pick the topmost.
		f := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		if _, ok := allocSet[f]; ok {
			continue // already added to list
		}
		// Add to set of allocating functions.
		allocSet[f] = struct{}{}
		allocList = append(allocList, f)

		// Add all callees to the worklist.
		for _, use := range getUses(f) {
			if use.IsACallInst().IsNil() {
				// TODO: function pointers
				panic("allocating function " + f.Name() + " used as function pointer")
			}
			parent := use.InstructionParent().Parent()
			for i := 0; i < use.OperandsCount()-1; i++ {
				if use.Operand(i) == f {
					// TODO: function pointers
					panic("allocating function " + f.Name() + " used as function pointer in " + parent.Name())
				}
			}
			worklist = append(worklist, parent)
		}
	}

	i8ptrPtrType := llvm.PointerType(c.i8ptrType, 0)
	gcrootType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{i8ptrPtrType, c.i8ptrType}, false)
	gcroot := llvm.AddFunction(c.mod, "llvm.gcroot", gcrootType)

	// Process every function that needs to save pointers to the shadow stack.
	for _, fn := range allocList {
		if fn == alloc {
			// runtime.alloc itself should not be treated this way, it is a
			// special case.
			continue
		}

		// Check all instructions in this function and see whether the value
		// needs to be kept on the shadow stack.
		var values []llvm.Value // values to be kept in the shadow stack
		for bb := fn.EntryBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				if !typeHasPointer(inst.Type()) {
					// This instruction does not result in a pointer value.
					continue
				}
				// Check whether any of the uses may occur after a call to
				// runtime.alloca. For example, if there are no call
				// instructions between the definition and the use, then the
				// pointer does not have to be stored in the shadow stack.
				for _, use := range getUses(inst) {
					if crossesAllocatingInst(inst, use, allocSet) {
						values = append(values, inst)
						break
					}
				}
			}
		}
		if len(values) == 0 {
			// The children of this function do allocations, but there is
			// nothing to keep in a stack frame for this function.
			continue
		}
		fn.SetGC("shadow-stack")

		// Convert all values to be kept in the shadow stack to actually be in
		// the shadow stack.
		firstInst := fn.EntryBasicBlock().FirstInstruction()
		for _, value := range values {
			valueUses := getUses(value)
			c.builder.SetInsertPointBefore(firstInst)
			alloca := c.builder.CreateAlloca(value.Type(), "gcroot.value")
			c.builder.SetInsertPointBefore(llvm.NextInstruction(value))
			c.builder.CreateStore(value, alloca)
			metadata := c.gcTypeMetadata(alloca.Type().ElementType())
			allocaCast := alloca
			if alloca.Type() != i8ptrPtrType {
				allocaCast = c.builder.CreateBitCast(alloca, i8ptrPtrType, "")
			}
			c.builder.CreateCall(gcroot, []llvm.Value{allocaCast, metadata}, "")
			for _, use := range valueUses {
				c.builder.SetInsertPointBefore(use)
				load := c.builder.CreateLoad(alloca, "")
				for i := 0; i < use.OperandsCount(); i++ {
					if use.Operand(i) == value {
						use.SetOperand(i, load)
					}
				}
			}
		}
	}
	println(c.IR())
}

// typeHasPointer returns true if (and only if) the given type contains a
// pointer value.
func typeHasPointer(typ llvm.Type) bool {
	switch typ.TypeKind() {
	case llvm.PointerTypeKind:
		return true
	case llvm.ArrayTypeKind, llvm.VectorTypeKind:
		return typeHasPointer(typ.ElementType())
	case llvm.StructTypeKind:
		return false
		for _, subtyp := range typ.StructElementTypes() {
			if typeHasPointer(subtyp) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// gcTypeMetadata returns a pointer value to be used in the llvm.gcroot
// intrinsic. It is either a null pointer or a number which is the number of
// words in the stack slot for this value.
func (c *Compiler) gcTypeMetadata(typ llvm.Type) llvm.Value {
	if typ.TypeKind() == llvm.PointerTypeKind {
		// Simple pointer. This is a common case, so signal this fact by setting
		// the pointer to null.
		return llvm.ConstPointerNull(c.i8ptrType)
	}
	if typ.TypeKind() == llvm.StructTypeKind {
		// Check for structs that only contain a pointer at the start.
		// We can pretend that such structs are a simple pointer, as the GC only
		// needs to read the first word.
		subTypes := typ.StructElementTypes()
		onlyFirstPointer := subTypes[0].TypeKind() == llvm.PointerTypeKind
		for _, subType := range subTypes[1:] {
			if typeHasPointer(subType) {
				onlyFirstPointer = false
			}
		}
		if onlyFirstPointer {
			// Types like string and slice.
			return llvm.ConstPointerNull(c.i8ptrType)
		}
	}
	allocaSize := c.targetData.TypeAllocSize(typ)
	pointerAlignment := uint64(c.targetData.PrefTypeAlignment(c.i8ptrType))
	numWords := allocaSize / pointerAlignment
	// TODO: only return the number until all pointers are included in this
	// struct, not more.
	metadata := llvm.ConstIntToPtr(llvm.ConstInt(c.uintptrType, numWords, false), c.i8ptrType)
	return metadata
}

// crossesAllocatingInst returns true if the given value may be used across a
// call to runtime.alloc. This check is very conservative.
func crossesAllocatingInst(from, to llvm.Value, allocSet map[llvm.Value]struct{}) bool {
	if from.InstructionParent() != to.InstructionParent() {
		// Don't try to check the CFG, conservatively assume there is an alloca
		// in between these instructions.
		return true
	}
	for inst := llvm.NextInstruction(from); inst != to; inst = llvm.NextInstruction(inst) {
		if inst.IsACallInst().IsNil() {
			// Not a call instruction thus not an alloca instruction.
			continue
		}
		if _, ok := allocSet[inst.CalledValue()]; ok {
			// This call is to a function that may do an allocation, or is even
			// runtime.alloc itself.
			// TODO: function pointers
			return true
		}
	}
	return false
}
