package compiler

import (
	"github.com/aykevl/go-llvm"
)

// Run the LLVM optimizer over the module.
// The inliner can be disabled (if necessary) by passing 0 to the inlinerThreshold.
func (c *Compiler) Optimize(optLevel, sizeLevel int, inlinerThreshold uint) {
	builder := llvm.NewPassManagerBuilder()
	defer builder.Dispose()
	builder.SetOptLevel(optLevel)
	builder.SetSizeLevel(sizeLevel)
	if inlinerThreshold != 0 {
		builder.UseInlinerWithThreshold(inlinerThreshold)
	}
	builder.AddCoroutinePassesToExtensionPoints()

	// Run function passes for each function.
	funcPasses := llvm.NewFunctionPassManagerForModule(c.mod)
	defer funcPasses.Dispose()
	builder.PopulateFunc(funcPasses)
	funcPasses.InitializeFunc()
	for fn := c.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		funcPasses.RunFunc(fn)
	}
	funcPasses.FinalizeFunc()

	if optLevel > 0 {
		// Run some preparatory passes for the Go optimizer.
		goPasses := llvm.NewPassManager()
		defer goPasses.Dispose()
		goPasses.AddGlobalOptimizerPass()
		goPasses.AddConstantPropagationPass()
		goPasses.AddAggressiveDCEPass()
		goPasses.AddFunctionAttrsPass()
		goPasses.Run(c.mod)

		// Run Go-specific optimization passes.
		c.OptimizeAllocs()
		c.Verify()
	}

	// Run module passes.
	modPasses := llvm.NewPassManager()
	defer modPasses.Dispose()
	builder.Populate(modPasses)
	modPasses.Run(c.mod)
}

// Basic escape analysis: translate runtime.alloc calls into alloca
// instructions.
func (c *Compiler) OptimizeAllocs() {
	allocator := c.mod.NamedFunction("runtime.alloc")
	if allocator.IsNil() {
		// nothing to optimize
		return
	}

	heapallocs := getUses(allocator)
	for _, heapalloc := range heapallocs {
		nilValue := llvm.Value{}
		if heapalloc.Operand(0).IsAConstant() == nilValue {
			// Do not allocate variable length arrays on the stack.
			continue
		}
		size := heapalloc.Operand(0).ZExtValue()
		if size > 256 {
			// The maximum value for a stack allocation.
			// TODO: tune this, this is just a random value.
			continue
		}

		// In general the pattern is:
		//     %0 = call i8* @runtime.alloc(i32 %size)
		//     %1 = bitcast i8* %0 to type*
		//     (use %1 only)
		// But the bitcast might sometimes be dropped when allocating an *i8.
		// The 'bitcast' variable below is thus usually a bitcast of the
		// heapalloc but not always.
		bitcast := heapalloc // instruction that creates the value
		if uses := getUses(heapalloc); len(uses) == 1 && uses[0].IsABitCastInst() != nilValue {
			// getting only bitcast use
			bitcast = uses[0]
		}
		if !c.doesEscape(bitcast) {
			// Insert alloca in the entry block. Do it here so that mem2reg can
			// promote it to a SSA value.
			fn := bitcast.InstructionParent().Parent()
			c.builder.SetInsertPointBefore(fn.EntryBasicBlock().FirstInstruction())
			allocaType := llvm.ArrayType(llvm.Int8Type(), int(size))
			// TODO: alignment?
			alloca := c.builder.CreateAlloca(allocaType, "stackalloc.alloca")
			stackalloc := c.builder.CreateBitCast(alloca, bitcast.Type(), "stackalloc")
			bitcast.ReplaceAllUsesWith(stackalloc)
			if heapalloc != bitcast {
				bitcast.EraseFromParentAsInstruction()
			}
			heapalloc.EraseFromParentAsInstruction()
		}
	}
}

// Very basic escape analysis.
func (c *Compiler) doesEscape(value llvm.Value) bool {
	uses := getUses(value)
	for _, use := range uses {
		nilValue := llvm.Value{}
		if use.IsAGetElementPtrInst() != nilValue {
			if c.doesEscape(use) {
				return true
			}
		} else if use.IsALoadInst() != nilValue {
			// Load does not escape.
		} else if use.IsAStoreInst() != nilValue {
			// Store only escapes when the value is stored to, not when the
			// value is stored into another value.
			if use.Operand(0) == value {
				return true
			}
		} else if use.IsACallInst() != nilValue {
			// Call only escapes when the (pointer) parameter is not marked
			// "nocapture". This flag means that the parameter does not escape
			// the give function.
			fn := use.CalledValue()
			if fn.IsAFunction() == nilValue {
				// This is not a function but something else, like a function
				// pointer.
				return false
			}
			nocaptureKind := llvm.AttributeKindID("nocapture")
			for i := 0; i < fn.ParamsCount(); i++ {
				if use.Operand(i) != value {
					// This is not the parameter we're checking.
					continue
				}
				index := i + 1 // param attributes start at 1
				nocapture := fn.GetEnumAttributeAtIndex(index, nocaptureKind)
				nilAttribute := llvm.Attribute{}
				if nocapture == nilAttribute {
					// Parameter doesn't have the nocapture flag, so may escape.
					return true
				}
			}
		} else {
			// Unknown instruction, might escape.
			return true
		}
	}

	// does not escape
	return false
}

// Return a list of values (actually, instructions) where this value is used as
// an operand.
func getUses(value llvm.Value) []llvm.Value {
	var uses []llvm.Value
	use := value.FirstUse()
	for !use.IsNil() {
		uses = append(uses, use.User())
		use = use.NextUse()
	}
	return uses
}
