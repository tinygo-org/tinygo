package compiler

import (
	"errors"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

// Run the LLVM optimizer over the module.
// The inliner can be disabled (if necessary) by passing 0 to the inlinerThreshold.
func (c *Compiler) Optimize(optLevel, sizeLevel int, inlinerThreshold uint) error {
	builder := llvm.NewPassManagerBuilder()
	defer builder.Dispose()
	builder.SetOptLevel(optLevel)
	builder.SetSizeLevel(sizeLevel)
	if inlinerThreshold != 0 {
		builder.UseInlinerWithThreshold(inlinerThreshold)
	}
	builder.AddCoroutinePassesToExtensionPoints()

	if c.PanicStrategy == "trap" {
		c.replacePanicsWithTrap() // -panic=trap
	}

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
		transform.OptimizeMaps(c.mod)
		c.OptimizeStringToBytes()
		transform.OptimizeAllocs(c.mod)
		c.LowerInterfaces()
		c.LowerFuncValues()

		// After interfaces are lowered, there are many more opportunities for
		// interprocedural optimizations. To get them to work, function
		// attributes have to be updated first.
		goPasses.Run(c.mod)

		// Run TinyGo-specific interprocedural optimizations.
		transform.OptimizeAllocs(c.mod)
		c.OptimizeStringToBytes()

		// Lower runtime.isnil calls to regular nil comparisons.
		isnil := c.mod.NamedFunction("runtime.isnil")
		if !isnil.IsNil() {
			for _, use := range getUses(isnil) {
				c.builder.SetInsertPointBefore(use)
				ptr := use.Operand(0)
				if !ptr.IsABitCastInst().IsNil() {
					ptr = ptr.Operand(0)
				}
				nilptr := llvm.ConstPointerNull(ptr.Type())
				icmp := c.builder.CreateICmp(llvm.IntEQ, ptr, nilptr, "")
				use.ReplaceAllUsesWith(icmp)
				use.EraseFromParentAsInstruction()
			}
		}

		err := c.LowerGoroutines()
		if err != nil {
			return err
		}
	} else {
		// Must be run at any optimization level.
		c.LowerInterfaces()
		c.LowerFuncValues()
		err := c.LowerGoroutines()
		if err != nil {
			return err
		}
	}
	if err := c.Verify(); err != nil {
		return errors.New("optimizations caused a verification failure")
	}

	if sizeLevel >= 2 {
		// Set the "optsize" attribute to make slightly smaller binaries at the
		// cost of some performance.
		kind := llvm.AttributeKindID("optsize")
		attr := c.ctx.CreateEnumAttribute(kind, 0)
		for fn := c.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
			fn.AddFunctionAttr(attr)
		}
	}

	// After TinyGo-specific transforms have finished, undo exporting these functions.
	for _, name := range functionsUsedInTransforms {
		fn := c.mod.NamedFunction(name)
		if fn.IsNil() {
			continue
		}
		fn.SetLinkage(llvm.InternalLinkage)
	}

	// Run function passes again, because without it, llvm.coro.size.i32()
	// doesn't get lowered.
	for fn := c.mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		funcPasses.RunFunc(fn)
	}
	funcPasses.FinalizeFunc()

	// Run module passes.
	modPasses := llvm.NewPassManager()
	defer modPasses.Dispose()
	builder.Populate(modPasses)
	modPasses.Run(c.mod)

	hasGCPass := c.addGlobalsBitmap()
	hasGCPass = c.makeGCStackSlots() || hasGCPass
	if hasGCPass {
		if err := c.Verify(); err != nil {
			return errors.New("GC pass caused a verification failure")
		}
	}

	return nil
}

// Replace panic calls with calls to llvm.trap, to reduce code size. This is the
// -panic=trap intrinsic.
func (c *Compiler) replacePanicsWithTrap() {
	trap := c.mod.NamedFunction("llvm.trap")
	for _, name := range []string{"runtime._panic", "runtime.runtimePanic"} {
		fn := c.mod.NamedFunction(name)
		if fn.IsNil() {
			continue
		}
		for _, use := range getUses(fn) {
			if use.IsACallInst().IsNil() || use.CalledValue() != fn {
				panic("expected use of a panic function to be a call")
			}
			c.builder.SetInsertPointBefore(use)
			c.builder.CreateCall(trap, nil, "")
		}
	}
}

// Transform runtime.stringToBytes(...) calls into const []byte slices whenever
// possible. This optimizes the following pattern:
//     w.Write([]byte("foo"))
// where Write does not store to the slice.
func (c *Compiler) OptimizeStringToBytes() {
	stringToBytes := c.mod.NamedFunction("runtime.stringToBytes")
	if stringToBytes.IsNil() {
		// nothing to optimize
		return
	}

	for _, call := range getUses(stringToBytes) {
		strptr := call.Operand(0)
		strlen := call.Operand(1)

		// strptr is always constant because strings are always constant.

		convertedAllUses := true
		for _, use := range getUses(call) {
			nilValue := llvm.Value{}
			if use.IsAExtractValueInst() == nilValue {
				convertedAllUses = false
				continue
			}
			switch use.Type().TypeKind() {
			case llvm.IntegerTypeKind:
				// A length (len or cap). Propagate the length value.
				use.ReplaceAllUsesWith(strlen)
				use.EraseFromParentAsInstruction()
			case llvm.PointerTypeKind:
				// The string pointer itself.
				if !c.isReadOnly(use) {
					convertedAllUses = false
					continue
				}
				use.ReplaceAllUsesWith(strptr)
				use.EraseFromParentAsInstruction()
			default:
				// should not happen
				panic("unknown return type of runtime.stringToBytes: " + use.Type().String())
			}
		}
		if convertedAllUses {
			// Call to runtime.stringToBytes can be eliminated: both the input
			// and the output is constant.
			call.EraseFromParentAsInstruction()
		}
	}
}

// Check whether the given value (which is of pointer type) is never stored to.
func (c *Compiler) isReadOnly(value llvm.Value) bool {
	uses := getUses(value)
	for _, use := range uses {
		nilValue := llvm.Value{}
		if use.IsAGetElementPtrInst() != nilValue {
			if !c.isReadOnly(use) {
				return false
			}
		} else if use.IsACallInst() != nilValue {
			if !c.hasFlag(use, value, "readonly") {
				return false
			}
		} else {
			// Unknown instruction, might not be readonly.
			return false
		}
	}
	return true
}

// Check whether all uses of this param as parameter to the call have the given
// flag. In most cases, there will only be one use but a function could take the
// same parameter twice, in which case both must have the flag.
// A flag can be any enum flag, like "readonly".
func (c *Compiler) hasFlag(call, param llvm.Value, kind string) bool {
	fn := call.CalledValue()
	nilValue := llvm.Value{}
	if fn.IsAFunction() == nilValue {
		// This is not a function but something else, like a function pointer.
		return false
	}
	kindID := llvm.AttributeKindID(kind)
	for i := 0; i < fn.ParamsCount(); i++ {
		if call.Operand(i) != param {
			// This is not the parameter we're checking.
			continue
		}
		index := i + 1 // param attributes start at 1
		attr := fn.GetEnumAttributeAtIndex(index, kindID)
		nilAttribute := llvm.Attribute{}
		if attr == nilAttribute {
			// At least one parameter doesn't have the flag (there may be
			// multiple).
			return false
		}
	}
	return true
}
