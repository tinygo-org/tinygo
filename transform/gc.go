package transform

import (
	"math/big"

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
			uintptrType,         // Number of elements in this frame.
		}
		for _, alloca := range allocas {
			fields = append(fields, alloca.Type().ElementType())
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
		parent := builder.CreateLoad(stackChainStart, "")
		gep := builder.CreateGEP(stackObject, []llvm.Value{
			llvm.ConstInt(ctx.Int32Type(), 0, false),
			llvm.ConstInt(ctx.Int32Type(), 0, false),
		}, "")
		builder.CreateStore(parent, gep)
		stackObjectCast := builder.CreateBitCast(stackObject, stackChainStartType, "")
		builder.CreateStore(stackObjectCast, stackChainStart)

		// Replace all independent allocas with GEPs in the stack object.
		for i, alloca := range allocas {
			gep := builder.CreateGEP(stackObject, []llvm.Value{
				llvm.ConstInt(ctx.Int32Type(), 0, false),
				llvm.ConstInt(ctx.Int32Type(), uint64(2+i), false),
			}, "")
			alloca.ReplaceAllUsesWith(gep)
			alloca.EraseFromParentAsInstruction()
		}

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
			gep := builder.CreateGEP(stackObject, []llvm.Value{
				llvm.ConstInt(ctx.Int32Type(), 0, false),
				llvm.ConstInt(ctx.Int32Type(), uint64(2+len(allocas)+i), false),
			}, "")

			// Store the pointer into the stack slot.
			store := builder.CreateStore(ptr, gep)
			pointerStores[store] = struct{}{}
		}

		// Make sure this stack object is popped from the linked list of stack
		// objects at return.
		for _, ret := range returns {
			inst := ret
			// Try to do the popping of the stack object earlier, by inserting
			// it not right before the return instruction but moving the insert
			// position up.
			// This is necessary so that the GC stack slot pass doesn't
			// interfere with tail calls (in particular, musttail calls).
			for {
				prevInst := llvm.PrevInstruction(inst)
				if prevInst == parent {
					break
				}
				if _, ok := pointerStores[prevInst]; ok {
					// Pop the stack object after the last store instruction.
					// This can probably be made more efficient: storing to the
					// stack chain object and then immediately popping isn't
					// useful.
					break
				}
				if prevInst.IsNil() {
					// Start of basic block. Pop the stack object here.
					break
				}
				if !prevInst.IsAPHINode().IsNil() {
					// Do not insert before a PHI node. PHI nodes must be
					// grouped at the beginning of a basic block before any
					// other instruction.
					break
				}
				inst = prevInst
			}
			builder.SetInsertPointBefore(inst)
			builder.CreateStore(parent, stackChainStart)
		}
	}

	return true
}

// AddGlobalsBitmap performs a few related functions. It is needed for scanning
// globals on platforms where the .data/.bss section is not easily accessible by
// the GC, and thus all globals that contain pointers must be made reachable by
// the GC in some other way.
//
// First, it scans all globals, and bundles all globals that contain a pointer
// into one large global (updating all uses in the process). Then it creates a
// bitmap (bit vector) to locate all the pointers in this large global. This
// bitmap allows the GC to know in advance where exactly all the pointers live
// in the large globals bundle, to avoid false positives.
func AddGlobalsBitmap(mod llvm.Module) bool {
	if mod.NamedGlobal("runtime.trackedGlobalsStart").IsNil() {
		return false // nothing to do: no GC in use
	}

	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	uintptrType := ctx.IntType(targetData.PointerSize() * 8)

	// Collect all globals that contain pointers (and thus must be scanned by
	// the GC).
	var trackedGlobals []llvm.Value
	var trackedGlobalTypes []llvm.Type
	for global := mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if global.IsDeclaration() || global.IsGlobalConstant() {
			continue
		}
		typ := global.Type().ElementType()
		ptrs := getPointerBitmap(targetData, typ, global.Name())
		if ptrs.BitLen() == 0 {
			continue
		}
		trackedGlobals = append(trackedGlobals, global)
		trackedGlobalTypes = append(trackedGlobalTypes, typ)
	}

	// Make a new global that bundles all existing globals, and remove the
	// existing globals. All uses of the previous independent globals are
	// replaced with a GEP into the new globals bundle.
	globalsBundleType := ctx.StructType(trackedGlobalTypes, false)
	globalsBundle := llvm.AddGlobal(mod, globalsBundleType, "tinygo.trackedGlobals")
	globalsBundle.SetLinkage(llvm.InternalLinkage)
	globalsBundle.SetUnnamedAddr(true)
	initializer := llvm.Undef(globalsBundleType)
	for i, global := range trackedGlobals {
		initializer = llvm.ConstInsertValue(initializer, global.Initializer(), []uint32{uint32(i)})
		gep := llvm.ConstGEP(globalsBundle, []llvm.Value{
			llvm.ConstInt(ctx.Int32Type(), 0, false),
			llvm.ConstInt(ctx.Int32Type(), uint64(i), false),
		})
		global.ReplaceAllUsesWith(gep)
		global.EraseFromParentAsGlobal()
	}
	globalsBundle.SetInitializer(initializer)

	// Update trackedGlobalsStart, which points to the globals bundle.
	trackedGlobalsStart := llvm.ConstPtrToInt(globalsBundle, uintptrType)
	mod.NamedGlobal("runtime.trackedGlobalsStart").SetInitializer(trackedGlobalsStart)
	mod.NamedGlobal("runtime.trackedGlobalsStart").SetLinkage(llvm.InternalLinkage)

	// Update trackedGlobalsLength, which contains the length (in words) of the
	// globals bundle.
	alignment := targetData.PrefTypeAlignment(llvm.PointerType(ctx.Int8Type(), 0))
	trackedGlobalsLength := llvm.ConstInt(uintptrType, targetData.TypeAllocSize(globalsBundleType)/uint64(alignment), false)
	mod.NamedGlobal("runtime.trackedGlobalsLength").SetLinkage(llvm.InternalLinkage)
	mod.NamedGlobal("runtime.trackedGlobalsLength").SetInitializer(trackedGlobalsLength)

	// Create a bitmap (a new global) that stores for each word in the globals
	// bundle whether it contains a pointer. This allows globals to be scanned
	// precisely: no non-pointers will be considered pointers if the bit pattern
	// looks like one.
	// This code assumes that pointers are self-aligned. For example, that a
	// 32-bit (4-byte) pointer is also aligned to 4 bytes.
	bitmapBytes := getPointerBitmap(targetData, globalsBundleType, "globals bundle").Bytes()
	bitmapValues := make([]llvm.Value, len(bitmapBytes))
	for i, b := range bitmapBytes {
		bitmapValues[len(bitmapBytes)-i-1] = llvm.ConstInt(ctx.Int8Type(), uint64(b), false)
	}
	bitmapArray := llvm.ConstArray(ctx.Int8Type(), bitmapValues)
	bitmapNew := llvm.AddGlobal(mod, bitmapArray.Type(), "runtime.trackedGlobalsBitmap.tmp")
	bitmapOld := mod.NamedGlobal("runtime.trackedGlobalsBitmap")
	bitmapOld.ReplaceAllUsesWith(llvm.ConstBitCast(bitmapNew, bitmapOld.Type()))
	bitmapNew.SetInitializer(bitmapArray)
	bitmapNew.SetName("runtime.trackedGlobalsBitmap")
	bitmapNew.SetLinkage(llvm.InternalLinkage)

	return true // the IR was changed
}

// getPointerBitmap scans the given LLVM type for pointers and sets bits in a
// bigint at the word offset that contains a pointer. This scan is recursive.
func getPointerBitmap(targetData llvm.TargetData, typ llvm.Type, name string) *big.Int {
	alignment := targetData.PrefTypeAlignment(llvm.PointerType(typ.Context().Int8Type(), 0))
	switch typ.TypeKind() {
	case llvm.IntegerTypeKind, llvm.FloatTypeKind, llvm.DoubleTypeKind:
		return big.NewInt(0)
	case llvm.PointerTypeKind:
		return big.NewInt(1)
	case llvm.StructTypeKind:
		ptrs := big.NewInt(0)
		for i, subtyp := range typ.StructElementTypes() {
			subptrs := getPointerBitmap(targetData, subtyp, name)
			if subptrs.BitLen() == 0 {
				continue
			}
			offset := targetData.ElementOffset(typ, i)
			if offset%uint64(alignment) != 0 {
				panic("precise GC: global contains unaligned pointer: " + name)
			}
			subptrs.Lsh(subptrs, uint(offset)/uint(alignment))
			ptrs.Or(ptrs, subptrs)
		}
		return ptrs
	case llvm.ArrayTypeKind:
		subtyp := typ.ElementType()
		subptrs := getPointerBitmap(targetData, subtyp, name)
		ptrs := big.NewInt(0)
		if subptrs.BitLen() == 0 {
			return ptrs
		}
		elementSize := targetData.TypeAllocSize(subtyp)
		for i := 0; i < typ.ArrayLength(); i++ {
			ptrs.Lsh(ptrs, uint(elementSize)/uint(alignment))
			ptrs.Or(ptrs, subptrs)
		}
		return ptrs
	default:
		panic("unknown type kind of global: " + name)
	}
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
