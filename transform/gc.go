package transform

import (
	"tinygo.org/x/go-llvm"
)

// MakeGCStackSlots converts all calls to runtime.trackPointer to explicit
// stores to stack slots that are scannable by the GC.
func MakeGCStackSlots(mod llvm.Module) bool {
	// Check whether there are allocations at all.
	alloc := mod.NamedFunction("runtime.alloc")
	if alloc.IsNil() {
		// Nothing to. Make sure all remaining bits and pieces for stack
		// chains are neutralized.
		for _, call := range getUses(mod.NamedFunction("runtime.trackPointer")) {
			call.EraseFromParentAsInstruction()
		}
		stackChainStart := mod.NamedGlobal("runtime.stackChainStart")
		if !stackChainStart.IsNil() {
			stackChainStart.SetLinkage(llvm.InternalLinkage)
			stackChainStart.SetInitializer(llvm.ConstNull(stackChainStart.Type().ElementType()))
			stackChainStart.SetGlobalConstant(true)
		}
		return false
	}

	trackPointer := mod.NamedFunction("runtime.trackPointer")
	if trackPointer.IsNil() || trackPointer.FirstUse().IsNil() {
		return false // nothing to do
	}

	ctx := mod.Context()
	builder := ctx.NewBuilder()
	defer builder.Dispose()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	uintptrType := ctx.IntType(targetData.PointerSize() * 8)

	// Look at *all* functions to see whether they are free of function pointer
	// calls.
	// This takes less than 5ms for ~100kB of WebAssembly but would perhaps be
	// faster when written in C++ (to avoid the CGo overhead).
	funcsWithFPCall := map[llvm.Value]struct{}{}
	n := 0
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		n++
		if _, ok := funcsWithFPCall[fn]; ok {
			continue // already found
		}
		done := false
		for bb := fn.FirstBasicBlock(); !bb.IsNil() && !done; bb = llvm.NextBasicBlock(bb) {
			for call := bb.FirstInstruction(); !call.IsNil() && !done; call = llvm.NextInstruction(call) {
				if call.IsACallInst().IsNil() {
					continue // only looking at calls
				}
				called := call.CalledValue()
				if !called.IsAFunction().IsNil() {
					continue // only looking for function pointers
				}
				funcsWithFPCall[fn] = struct{}{}
				markParentFunctions(funcsWithFPCall, fn)
				done = true
			}
		}
	}

	// Determine which functions need stack objects. Many leaf functions don't
	// need it: it only causes overhead for them.
	// Actually, in one test it was only able to eliminate stack object from 12%
	// of functions that had a call to runtime.trackPointer (8 out of 68
	// functions), so this optimization is not as big as it may seem.
	allocatingFunctions := map[llvm.Value]struct{}{} // set of allocating functions

	// Work from runtime.alloc and trace all parents to check which functions do
	// a heap allocation (and thus which functions do not).
	markParentFunctions(allocatingFunctions, alloc)

	// Also trace all functions that call a function pointer.
	for fn := range funcsWithFPCall {
		// Assume that functions that call a function pointer do a heap
		// allocation as a conservative guess because the called function might
		// do a heap allocation.
		allocatingFunctions[fn] = struct{}{}
		markParentFunctions(allocatingFunctions, fn)
	}

	// Collect some variables used below in the loop.
	stackChainStart := mod.NamedGlobal("runtime.stackChainStart")
	if stackChainStart.IsNil() {
		// This may be reached in a weird scenario where we call runtime.alloc but the garbage collector is unreachable.
		// This can be accomplished by allocating 0 bytes.
		// There is no point in tracking anything.
		for _, use := range getUses(trackPointer) {
			use.EraseFromParentAsInstruction()
		}
		return false
	}
	stackChainStart.SetLinkage(llvm.InternalLinkage)
	stackChainStartType := stackChainStart.Type().ElementType()
	stackChainStart.SetInitializer(llvm.ConstNull(stackChainStartType))

	// Iterate until runtime.trackPointer has no uses left.
	for use := trackPointer.FirstUse(); !use.IsNil(); use = trackPointer.FirstUse() {
		// Pick the first use of runtime.trackPointer.
		call := use.User()
		if call.IsACallInst().IsNil() {
			panic("expected runtime.trackPointer use to be a call")
		}

		// Pick the parent function.
		fn := call.InstructionParent().Parent()

		if _, ok := allocatingFunctions[fn]; !ok {
			// This function nor any of the functions it calls (recursively)
			// allocate anything from the heap, so it will not trigger a garbage
			// collection cycle. Thus, it does not need to track local pointer
			// values.
			// This is a useful optimization but not as big as you might guess,
			// as described above (it avoids stack objects for ~12% of
			// functions).
			call.EraseFromParentAsInstruction()
			continue
		}

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
		var pointers []llvm.Value
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
			case llvm.GetElementPtr:
				// Check for all zero offsets.
				// Sometimes LLVM rewrites bitcasts to zero-index GEPs, and we still need to track the GEP.
				n := ptr.OperandsCount()
				var hasOffset bool
				for i := 1; i < n; i++ {
					offset := ptr.Operand(i)
					if offset.IsAConstantInt().IsNil() || offset.ZExtValue() != 0 {
						hasOffset = true
						break
					}
				}

				if hasOffset {
					// These values do not create new values: the values already
					// existed locally in this function so must have been tracked
					// already.
					continue
				}
			case llvm.PHI:
				// While the value may have already been tracked, it may be overwritten in a loop.
				// Therefore, a second copy must be created to ensure that it is tracked over the entirety of its lifetime.
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

			if ptr := stripPointerCasts(ptr); !ptr.IsAAllocaInst().IsNil() {
				// Allocas don't need to be tracked because they are allocated
				// on the C stack which is scanned separately.
				continue
			}
			pointers = append(pointers, ptr)
		}

		if len(pointers) == 0 {
			// This function does not need to keep track of stack pointers.
			continue
		}

		// Determine the type of the required stack slot.
		fields := []llvm.Type{
			stackChainStartType, // Pointer to parent frame.
			uintptrType,         // Number of elements in this frame.
		}
		for _, ptr := range pointers {
			fields = append(fields, ptr.Type())
		}
		stackObjectType := ctx.StructType(fields, false)

		// Create the stack object at the function entry.
		builder.SetInsertPointBefore(fn.EntryBasicBlock().FirstInstruction())
		stackObject := builder.CreateAlloca(stackObjectType, "gc.stackobject")
		initialStackObject := llvm.ConstNull(stackObjectType)
		numSlots := (targetData.TypeAllocSize(stackObjectType) - uint64(targetData.PointerSize())*2) / uint64(targetData.ABITypeAlignment(uintptrType))
		numSlotsValue := llvm.ConstInt(uintptrType, numSlots, false)
		initialStackObject = llvm.ConstInsertValue(initialStackObject, numSlotsValue, []uint32{1})
		builder.CreateStore(initialStackObject, stackObject)

		// Update stack start.
		parent := builder.CreateLoad(stackChainStartType, stackChainStart, "")
		gep := builder.CreateGEP(stackObjectType, stackObject, []llvm.Value{
			llvm.ConstInt(ctx.Int32Type(), 0, false),
			llvm.ConstInt(ctx.Int32Type(), 0, false),
		}, "")
		builder.CreateStore(parent, gep)
		stackObjectCast := builder.CreateBitCast(stackObject, stackChainStartType, "")
		builder.CreateStore(stackObjectCast, stackChainStart)

		// Do a store to the stack object after each new pointer that is created.
		pointerStores := make(map[llvm.Value]struct{})
		for i, ptr := range pointers {
			// Insert the store after the pointer value is created.
			insertionPoint := llvm.NextInstruction(ptr)
			for !insertionPoint.IsAPHINode().IsNil() {
				// PHI nodes are required to be at the start of the block.
				// Insert after the last PHI node.
				insertionPoint = llvm.NextInstruction(insertionPoint)
			}
			builder.SetInsertPointBefore(insertionPoint)

			// Extract a pointer to the appropriate section of the stack object.
			gep := builder.CreateGEP(stackObjectType, stackObject, []llvm.Value{
				llvm.ConstInt(ctx.Int32Type(), 0, false),
				llvm.ConstInt(ctx.Int32Type(), uint64(2+i), false),
			}, "")

			// Store the pointer into the stack slot.
			store := builder.CreateStore(ptr, gep)
			pointerStores[store] = struct{}{}
		}

		// Make sure this stack object is popped from the linked list of stack
		// objects at return.
		for _, ret := range returns {
			// Check for any tail calls at this return.
			prev := llvm.PrevInstruction(ret)
			if !prev.IsNil() && !prev.IsABitCastInst().IsNil() {
				// A bitcast can appear before a tail call, so skip backwards more.
				prev = llvm.PrevInstruction(prev)
			}
			if !prev.IsNil() && !prev.IsACallInst().IsNil() {
				// This is no longer a tail call.
				prev.SetTailCall(false)
			}
			builder.SetInsertPointBefore(ret)
			builder.CreateStore(parent, stackChainStart)
		}
	}

	return true
}

// markParentFunctions traverses all parent function calls (recursively) and
// adds them to the set of marked functions. It only considers function calls:
// any other uses of such a function is ignored.
func markParentFunctions(marked map[llvm.Value]struct{}, fn llvm.Value) {
	worklist := []llvm.Value{fn}
	for len(worklist) != 0 {
		fn := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		for _, use := range getUses(fn) {
			if use.IsACallInst().IsNil() || use.CalledValue() != fn {
				// Not the parent function.
				continue
			}
			parent := use.InstructionParent().Parent()
			if _, ok := marked[parent]; !ok {
				marked[parent] = struct{}{}
				worklist = append(worklist, parent)
			}
		}
	}
}
