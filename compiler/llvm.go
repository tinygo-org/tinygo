package compiler

import (
	"errors"
	"reflect"

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

// createEntryBlockAlloca creates a new alloca in the entry block, even though
// the IR builder is located elsewhere. It assumes that the insert point is
// at the end of the current block.
func (c *Compiler) createEntryBlockAlloca(t llvm.Type, name string) llvm.Value {
	currentBlock := c.builder.GetInsertBlock()
	entryBlock := currentBlock.Parent().EntryBasicBlock()
	if entryBlock.FirstInstruction().IsNil() {
		c.builder.SetInsertPointAtEnd(entryBlock)
	} else {
		c.builder.SetInsertPointBefore(entryBlock.FirstInstruction())
	}
	alloca := c.builder.CreateAlloca(t, name)
	c.builder.SetInsertPointAtEnd(currentBlock)
	return alloca
}

// createTemporaryAlloca creates a new alloca in the entry block and adds
// lifetime start infromation in the IR signalling that the alloca won't be used
// before this point.
//
// This is useful for creating temporary allocas for intrinsics. Don't forget to
// end the lifetime using emitLifetimeEnd after you're done with it.
func (c *Compiler) createTemporaryAlloca(t llvm.Type, name string) (alloca, bitcast, size llvm.Value) {
	alloca = c.createEntryBlockAlloca(t, name)
	bitcast = c.builder.CreateBitCast(alloca, c.i8ptrType, name+".bitcast")
	size = llvm.ConstInt(c.ctx.Int64Type(), c.targetData.TypeAllocSize(t), false)
	c.builder.CreateCall(c.getLifetimeStartFunc(), []llvm.Value{size, bitcast}, "")
	return
}

// emitLifetimeEnd signals the end of an (alloca) lifetime by calling the
// llvm.lifetime.end intrinsic. It is commonly used together with
// createTemporaryAlloca.
func (c *Compiler) emitLifetimeEnd(ptr, size llvm.Value) {
	c.builder.CreateCall(c.getLifetimeEndFunc(), []llvm.Value{size, ptr}, "")
}

// getLifetimeStartFunc returns the llvm.lifetime.start intrinsic and creates it
// first if it doesn't exist yet.
func (c *Compiler) getLifetimeStartFunc() llvm.Value {
	fn := c.mod.NamedFunction("llvm.lifetime.start.p0i8")
	if fn.IsNil() {
		fnType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.ctx.Int64Type(), c.i8ptrType}, false)
		fn = llvm.AddFunction(c.mod, "llvm.lifetime.start.p0i8", fnType)
	}
	return fn
}

// getLifetimeEndFunc returns the llvm.lifetime.end intrinsic and creates it
// first if it doesn't exist yet.
func (c *Compiler) getLifetimeEndFunc() llvm.Value {
	fn := c.mod.NamedFunction("llvm.lifetime.end.p0i8")
	if fn.IsNil() {
		fnType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.ctx.Int64Type(), c.i8ptrType}, false)
		fn = llvm.AddFunction(c.mod, "llvm.lifetime.end.p0i8", fnType)
	}
	return fn
}

// splitBasicBlock splits a LLVM basic block into two parts. All instructions
// after afterInst are moved into a new basic block (created right after the
// current one) with the given name.
func (c *Compiler) splitBasicBlock(afterInst llvm.Value, insertAfter llvm.BasicBlock, name string) llvm.BasicBlock {
	oldBlock := afterInst.InstructionParent()
	newBlock := c.ctx.InsertBasicBlock(insertAfter, name)
	var nextInstructions []llvm.Value // values to move

	// Collect to-be-moved instructions.
	inst := afterInst
	for {
		inst = llvm.NextInstruction(inst)
		if inst.IsNil() {
			break
		}
		nextInstructions = append(nextInstructions, inst)
	}

	// Move instructions.
	c.builder.SetInsertPointAtEnd(newBlock)
	for _, inst := range nextInstructions {
		inst.RemoveFromParentAsInstruction()
		c.builder.Insert(inst)
	}

	// Find PHI nodes to update.
	var phiNodes []llvm.Value // PHI nodes to update
	for bb := insertAfter.Parent().FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			if inst.IsAPHINode().IsNil() {
				continue
			}
			needsUpdate := false
			incomingCount := inst.IncomingCount()
			for i := 0; i < incomingCount; i++ {
				if inst.IncomingBlock(i) == oldBlock {
					needsUpdate = true
					break
				}
			}
			if !needsUpdate {
				// PHI node has no incoming edge from the old block.
				continue
			}
			phiNodes = append(phiNodes, inst)
		}
	}

	// Update PHI nodes.
	for _, phi := range phiNodes {
		c.builder.SetInsertPointBefore(phi)
		newPhi := c.builder.CreatePHI(phi.Type(), "")
		incomingCount := phi.IncomingCount()
		incomingVals := make([]llvm.Value, incomingCount)
		incomingBlocks := make([]llvm.BasicBlock, incomingCount)
		for i := 0; i < incomingCount; i++ {
			value := phi.IncomingValue(i)
			block := phi.IncomingBlock(i)
			if block == oldBlock {
				block = newBlock
			}
			incomingVals[i] = value
			incomingBlocks[i] = block
		}
		newPhi.AddIncoming(incomingVals, incomingBlocks)
		phi.ReplaceAllUsesWith(newPhi)
		phi.EraseFromParentAsInstruction()
	}

	return newBlock
}

// makeGlobalArray creates a new LLVM global with the given name and integers as
// contents, and returns the global.
// Note that it is left with the default linkage etc., you should set
// linkage/constant/etc properties yourself.
func (c *Compiler) makeGlobalArray(bufItf interface{}, name string, elementType llvm.Type) llvm.Value {
	buf := reflect.ValueOf(bufItf)
	globalType := llvm.ArrayType(elementType, buf.Len())
	global := llvm.AddGlobal(c.mod, globalType, name)
	value := llvm.Undef(globalType)
	for i := 0; i < buf.Len(); i++ {
		ch := buf.Index(i).Uint()
		value = llvm.ConstInsertValue(value, llvm.ConstInt(elementType, ch, false), []uint32{uint32(i)})
	}
	global.SetInitializer(value)
	return global
}

// getGlobalBytes returns the slice contained in the array of the provided
// global. It can recover the bytes originally created using makeGlobalArray, if
// makeGlobalArray was given a byte slice.
func getGlobalBytes(global llvm.Value) []byte {
	value := global.Initializer()
	buf := make([]byte, value.Type().ArrayLength())
	for i := range buf {
		buf[i] = byte(llvm.ConstExtractValue(value, []uint32{uint32(i)}).ZExtValue())
	}
	return buf
}

// replaceGlobalByteWithArray replaces a global integer type in the module with
// an integer array, using a GEP to make the types match. It is a convenience
// function used for creating reflection sidetables, for example.
func (c *Compiler) replaceGlobalIntWithArray(name string, buf interface{}) llvm.Value {
	oldGlobal := c.mod.NamedGlobal(name)
	global := c.makeGlobalArray(buf, name+".tmp", oldGlobal.Type().ElementType())
	gep := llvm.ConstGEP(global, []llvm.Value{
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		llvm.ConstInt(c.ctx.Int32Type(), 0, false),
	})
	oldGlobal.ReplaceAllUsesWith(gep)
	oldGlobal.EraseFromParentAsGlobal()
	global.SetName(name)
	return global
}

// decomposeValue breaks an aggregate value into primitive values.
func (c *Compiler) decomposeValue(v llvm.Value) []llvm.Value {
	t := v.Type()
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		n := t.StructElementTypesCount()
		fields := make([]llvm.Value, 0, n)
		for i := 0; i < n; i++ {
			fields = append(fields, c.decomposeValue(c.builder.CreateExtractValue(v, i, ""))...)
		}
		return fields
	case llvm.ArrayTypeKind:
		n := t.ArrayLength()
		elems := make([]llvm.Value, 0, n)
		for i := 0; i < n; i++ {
			elems = append(elems, c.decomposeValue(c.builder.CreateExtractValue(v, i, ""))...)
		}
		return elems
	case llvm.VectorTypeKind:
		n := t.VectorSize()
		elems := make([]llvm.Value, 0, n)
		for i := 0; i < n; i++ {
			elems = append(elems, c.decomposeValue(c.builder.CreateExtractElement(v, llvm.ConstInt(c.intType, uint64(i), false), ""))...)
		}
		return elems
	}
	return []llvm.Value{v}
}

// composeValue composes a value of the specified type from the set of primitive values provided.
// Returns the constructed value and remaining values.
// Appropriate conversions are added as needed.
func (c *Compiler) composeValue(t llvm.Type, vals ...llvm.Value) (llvm.Value, []llvm.Value, error) {
	switch t.TypeKind() {
	case llvm.StructTypeKind:
		fieldTypes := t.StructElementTypes()
		fields := make([]llvm.Value, len(fieldTypes))
		for i, t := range fieldTypes {
			var err error
			fields[i], vals, err = c.composeValue(t, vals...)
			if err != nil {
				return llvm.Value{}, nil, err
			}
		}
		s := llvm.ConstNull(t)
		for i, v := range fields {
			s = c.builder.CreateInsertValue(s, v, i, "")
		}
		return s, vals, nil
	case llvm.ArrayTypeKind:
		elemType := t.ElementType()
		elems := make([]llvm.Value, t.ArrayLength())
		for i := range elems {
			var err error
			elems[i], vals, err = c.composeValue(elemType, vals...)
			if err != nil {
				return llvm.Value{}, nil, err
			}
		}
		a := llvm.ConstNull(t)
		for i, e := range elems {
			a = c.builder.CreateInsertValue(a, e, i, "")
		}
		return a, vals, nil
	case llvm.VectorTypeKind:
		elemType := t.ElementType()
		elems := make([]llvm.Value, t.VectorSize())
		for i := range elems {
			var err error
			elems[i], vals, err = c.composeValue(elemType, vals...)
			if err != nil {
				return llvm.Value{}, nil, err
			}
		}
		v := llvm.ConstNull(t)
		for i, e := range elems {
			v = c.builder.CreateInsertElement(v, e, llvm.ConstInt(c.intType, uint64(i), false), "")
		}
		return v, vals, nil
	}

	if len(vals) < 1 {
		return llvm.Value{}, nil, errors.New("insufficient values to compose type")
	}

	if vals[0].Type() == t {
		// type already matches
		return vals[0], vals[1:], nil
	}

	return c.builder.CreateBitCast(vals[0], t, ""), vals[1:], nil
}

// deepCast converts the given value to the given type, recursively if necessary.
func (c *Compiler) deepCast(v llvm.Value, t llvm.Type) (llvm.Value, error) {
	if v.Type() == t {
		return v, nil
	}
	vals := c.decomposeValue(v)
	res, extra, err := c.composeValue(t, vals...)
	if err != nil {
		return llvm.Value{}, err
	}
	if len(extra) > 0 {
		return llvm.Value{}, errors.New("too many values")
	}
	return res, nil
}
