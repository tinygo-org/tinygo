package compiler

// This file provides IR transformations necessary for precise and portable
// garbage collectors.

import (
	"go/token"
	"math/big"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// needsStackObjects returns true if the compiler should insert stack objects
// that can be traced by the garbage collector.
func (c *Compiler) needsStackObjects() bool {
	if c.selectGC() != "conservative" {
		return false
	}
	for _, tag := range c.BuildTags {
		if tag == "cortexm" || tag == "tinygo.riscv" {
			return false
		}
	}

	return true
}

// trackExpr inserts pointer tracking intrinsics for the GC if the expression is
// one of the expressions that need this.
func (c *Compiler) trackExpr(frame *Frame, expr ssa.Value, value llvm.Value) {
	// There are uses of this expression, Make sure the pointers
	// are tracked during GC.
	switch expr := expr.(type) {
	case *ssa.Alloc, *ssa.MakeChan, *ssa.MakeMap:
		// These values are always of pointer type in IR.
		c.trackPointer(value)
	case *ssa.Call, *ssa.Convert, *ssa.MakeClosure, *ssa.MakeInterface, *ssa.MakeSlice, *ssa.Next:
		if !value.IsNil() {
			c.trackValue(value)
		}
	case *ssa.Select:
		if alloca, ok := frame.selectRecvBuf[expr]; ok {
			if alloca.IsAUndefValue().IsNil() {
				c.trackPointer(alloca)
			}
		}
	case *ssa.UnOp:
		switch expr.Op {
		case token.MUL:
			// Pointer dereference.
			c.trackValue(value)
		case token.ARROW:
			// Channel receive operator.
			// It's not necessary to look at commaOk here, because in that
			// case it's just an aggregate and trackValue will extract the
			// pointer in there (if there is one).
			c.trackValue(value)
		}
	}
}

// trackValue locates pointers in a value (possibly an aggregate) and tracks the
// individual pointers
func (c *Compiler) trackValue(value llvm.Value) {
	typ := value.Type()
	switch typ.TypeKind() {
	case llvm.PointerTypeKind:
		c.trackPointer(value)
	case llvm.StructTypeKind:
		if !typeHasPointers(typ) {
			return
		}
		numElements := typ.StructElementTypesCount()
		for i := 0; i < numElements; i++ {
			subValue := c.builder.CreateExtractValue(value, i, "")
			c.trackValue(subValue)
		}
	case llvm.ArrayTypeKind:
		if !typeHasPointers(typ) {
			return
		}
		numElements := typ.ArrayLength()
		for i := 0; i < numElements; i++ {
			subValue := c.builder.CreateExtractValue(value, i, "")
			c.trackValue(subValue)
		}
	}
}

// trackPointer creates a call to runtime.trackPointer, bitcasting the poitner
// first if needed. The input value must be of LLVM pointer type.
func (c *Compiler) trackPointer(value llvm.Value) {
	if value.Type() != c.i8ptrType {
		value = c.builder.CreateBitCast(value, c.i8ptrType, "")
	}
	c.createRuntimeCall("trackPointer", []llvm.Value{value}, "")
}

// typeHasPointers returns whether this type is a pointer or contains pointers.
// If the type is an aggregate type, it will check whether there is a pointer
// inside.
func typeHasPointers(t llvm.Type) bool {
	switch t.TypeKind() {
	case llvm.PointerTypeKind:
		return true
	case llvm.StructTypeKind:
		for _, subType := range t.StructElementTypes() {
			if typeHasPointers(subType) {
				return true
			}
		}
		return false
	case llvm.ArrayTypeKind:
		if typeHasPointers(t.ElementType()) {
			return true
		}
		return false
	default:
		return false
	}
}

// makeGCStackSlots converts all calls to runtime.trackPointer to explicit
// stores to stack slots that are scannable by the GC.
func (c *Compiler) makeGCStackSlots() bool {
	if c.mod.NamedFunction("runtime.alloc").IsNil() {
		// Nothing to. Make sure all remaining bits and pieces for stack
		// chains are neutralized.
		for _, call := range getUses(c.mod.NamedFunction("runtime.trackPointer")) {
			call.EraseFromParentAsInstruction()
		}
		stackChainStart := c.mod.NamedGlobal("runtime.stackChainStart")
		if !stackChainStart.IsNil() {
			stackChainStart.SetInitializer(c.getZeroValue(stackChainStart.Type().ElementType()))
			stackChainStart.SetGlobalConstant(true)
		}
	}

	trackPointer := c.mod.NamedFunction("runtime.trackPointer")
	if trackPointer.IsNil() || trackPointer.FirstUse().IsNil() {
		return false // nothing to do
	}

	// Collect some variables used below in the loop.
	stackChainStart := c.mod.NamedGlobal("runtime.stackChainStart")
	if stackChainStart.IsNil() {
		panic("stack chain start not found!")
	}
	stackChainStartType := stackChainStart.Type().ElementType()
	stackChainStart.SetInitializer(c.getZeroValue(stackChainStartType))

	// Iterate until runtime.trackPointer has no uses left.
	for use := trackPointer.FirstUse(); !use.IsNil(); use = trackPointer.FirstUse() {
		// Pick the first use of runtime.trackPointer.
		call := use.User()
		if call.IsACallInst().IsNil() {
			panic("expected runtime.trackPointer use to be a call")
		}

		// Pick the parent function.
		fn := call.InstructionParent().Parent()

		// Find all calls to runtime.trackPointer in this function.
		var calls []llvm.Value
		var returns []llvm.Value
		for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				switch inst.InstructionOpcode() {
				case llvm.Call:
					if inst.CalledValue() == trackPointer {
						calls = append(calls, inst)
					}
				case llvm.Ret:
					returns = append(returns, inst)
				}
			}
		}

		// Determine what to do with each call.
		var allocas, pointers []llvm.Value
		for _, call := range calls {
			ptr := call.Operand(0)
			call.EraseFromParentAsInstruction()
			if ptr.IsAInstruction().IsNil() {
				continue
			}

			// Some trivial optimizations.
			if ptr.IsAInstruction().IsNil() {
				continue
			}
			switch ptr.InstructionOpcode() {
			case llvm.PHI, llvm.GetElementPtr:
				// These values do not create new values: the values already
				// existed locally in this function so must have been tracked
				// already.
				continue
			case llvm.ExtractValue, llvm.BitCast:
				// These instructions do not create new values, but their
				// original value may not be tracked. So keep tracking them for
				// now.
				// With more analysis, it should be possible to optimize a
				// significant chunk of these away.
			case llvm.Call, llvm.Load, llvm.IntToPtr:
				// These create new values so must be stored locally. But
				// perhaps some of these can be fused when they actually refer
				// to the same value.
			default:
				// Ambiguous. These instructions are uncommon, but perhaps could
				// be optimized if needed.
			}

			if !ptr.IsAAllocaInst().IsNil() {
				if typeHasPointers(ptr.Type().ElementType()) {
					allocas = append(allocas, ptr)
				}
			} else {
				pointers = append(pointers, ptr)
			}
		}

		if len(allocas) == 0 && len(pointers) == 0 {
			// This function does not need to keep track of stack pointers.
			continue
		}

		// Determine the type of the required stack slot.
		fields := []llvm.Type{
			stackChainStartType, // Pointer to parent frame.
			c.uintptrType,       // Number of elements in this frame.
		}
		for _, alloca := range allocas {
			fields = append(fields, alloca.Type().ElementType())
		}
		for _, ptr := range pointers {
			fields = append(fields, ptr.Type())
		}
		stackObjectType := c.ctx.StructType(fields, false)

		// Create the stack object at the function entry.
		c.builder.SetInsertPointBefore(fn.EntryBasicBlock().FirstInstruction())
		stackObject := c.builder.CreateAlloca(stackObjectType, "gc.stackobject")
		initialStackObject := c.getZeroValue(stackObjectType)
		numSlots := (c.targetData.TypeAllocSize(stackObjectType) - c.targetData.TypeAllocSize(c.i8ptrType)*2) / uint64(c.targetData.ABITypeAlignment(c.uintptrType))
		numSlotsValue := llvm.ConstInt(c.uintptrType, numSlots, false)
		initialStackObject = llvm.ConstInsertValue(initialStackObject, numSlotsValue, []uint32{1})
		c.builder.CreateStore(initialStackObject, stackObject)

		// Update stack start.
		parent := c.builder.CreateLoad(stackChainStart, "")
		gep := c.builder.CreateGEP(stackObject, []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
		}, "")
		c.builder.CreateStore(parent, gep)
		stackObjectCast := c.builder.CreateBitCast(stackObject, stackChainStartType, "")
		c.builder.CreateStore(stackObjectCast, stackChainStart)

		// Replace all independent allocas with GEPs in the stack object.
		for i, alloca := range allocas {
			gep := c.builder.CreateGEP(stackObject, []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				llvm.ConstInt(c.ctx.Int32Type(), uint64(2+i), false),
			}, "")
			alloca.ReplaceAllUsesWith(gep)
			alloca.EraseFromParentAsInstruction()
		}

		// Do a store to the stack object after each new pointer that is created.
		for i, ptr := range pointers {
			c.builder.SetInsertPointBefore(llvm.NextInstruction(ptr))
			gep := c.builder.CreateGEP(stackObject, []llvm.Value{
				llvm.ConstInt(c.ctx.Int32Type(), 0, false),
				llvm.ConstInt(c.ctx.Int32Type(), uint64(2+len(allocas)+i), false),
			}, "")
			c.builder.CreateStore(ptr, gep)
		}

		// Make sure this stack object is popped from the linked list of stack
		// objects at return.
		for _, ret := range returns {
			c.builder.SetInsertPointBefore(ret)
			c.builder.CreateStore(parent, stackChainStart)
		}
	}

	return true
}

func (c *Compiler) addGlobalsBitmap() bool {
	if c.mod.NamedGlobal("runtime.trackedGlobalsStart").IsNil() {
		return false // nothing to do: no GC in use
	}

	var trackedGlobals []llvm.Value
	var trackedGlobalTypes []llvm.Type
	for global := c.mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if global.IsDeclaration() {
			continue
		}
		typ := global.Type().ElementType()
		ptrs := c.getPointerBitmap(typ, global.Name())
		if ptrs.BitLen() == 0 {
			continue
		}
		trackedGlobals = append(trackedGlobals, global)
		trackedGlobalTypes = append(trackedGlobalTypes, typ)
	}

	globalsBundleType := c.ctx.StructType(trackedGlobalTypes, false)
	globalsBundle := llvm.AddGlobal(c.mod, globalsBundleType, "tinygo.trackedGlobals")
	globalsBundle.SetLinkage(llvm.InternalLinkage)
	globalsBundle.SetUnnamedAddr(true)
	initializer := llvm.Undef(globalsBundleType)
	for i, global := range trackedGlobals {
		initializer = llvm.ConstInsertValue(initializer, global.Initializer(), []uint32{uint32(i)})
		gep := llvm.ConstGEP(globalsBundle, []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false),
		})
		global.ReplaceAllUsesWith(gep)
		global.EraseFromParentAsGlobal()
	}
	globalsBundle.SetInitializer(initializer)

	trackedGlobalsStart := llvm.ConstPtrToInt(globalsBundle, c.uintptrType)
	c.mod.NamedGlobal("runtime.trackedGlobalsStart").SetInitializer(trackedGlobalsStart)

	alignment := c.targetData.PrefTypeAlignment(c.i8ptrType)
	trackedGlobalsLength := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(globalsBundleType)/uint64(alignment), false)
	c.mod.NamedGlobal("runtime.trackedGlobalsLength").SetInitializer(trackedGlobalsLength)

	bitmapBytes := c.getPointerBitmap(globalsBundleType, "globals bundle").Bytes()
	bitmapValues := make([]llvm.Value, len(bitmapBytes))
	for i, b := range bitmapBytes {
		bitmapValues[len(bitmapBytes)-i-1] = llvm.ConstInt(c.ctx.Int8Type(), uint64(b), false)
	}
	bitmapArray := llvm.ConstArray(llvm.ArrayType(c.ctx.Int8Type(), len(bitmapBytes)), bitmapValues)
	bitmapNew := llvm.AddGlobal(c.mod, bitmapArray.Type(), "runtime.trackedGlobalsBitmap.tmp")
	bitmapOld := c.mod.NamedGlobal("runtime.trackedGlobalsBitmap")
	bitmapOld.ReplaceAllUsesWith(bitmapNew)
	bitmapNew.SetInitializer(bitmapArray)
	bitmapNew.SetName("runtime.trackedGlobalsBitmap")

	return true // the IR was changed
}

func (c *Compiler) getPointerBitmap(typ llvm.Type, name string) *big.Int {
	alignment := c.targetData.PrefTypeAlignment(c.i8ptrType)
	switch typ.TypeKind() {
	case llvm.IntegerTypeKind, llvm.FloatTypeKind, llvm.DoubleTypeKind:
		return big.NewInt(0)
	case llvm.PointerTypeKind:
		return big.NewInt(1)
	case llvm.StructTypeKind:
		ptrs := big.NewInt(0)
		for i, subtyp := range typ.StructElementTypes() {
			subptrs := c.getPointerBitmap(subtyp, name)
			if subptrs.BitLen() == 0 {
				continue
			}
			offset := c.targetData.ElementOffset(typ, i)
			if offset%uint64(alignment) != 0 {
				panic("precise GC: global contains unaligned pointer: " + name)
			}
			subptrs.Lsh(subptrs, uint(offset)/uint(alignment))
			ptrs.Or(ptrs, subptrs)
		}
		return ptrs
	case llvm.ArrayTypeKind:
		subtyp := typ.ElementType()
		subptrs := c.getPointerBitmap(subtyp, name)
		ptrs := big.NewInt(0)
		if subptrs.BitLen() == 0 {
			return ptrs
		}
		elementSize := c.targetData.TypeAllocSize(subtyp)
		for i := 0; i < typ.ArrayLength(); i++ {
			ptrs.Lsh(ptrs, uint(elementSize)/uint(alignment))
			ptrs.Or(ptrs, subptrs)
		}
		return ptrs
	default:
		panic("unknown type kind of global: " + name)
	}
}
