package compiler

import (
	"errors"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

// Run the LLVM optimizer over the module.
// The inliner can be disabled (if necessary) by passing 0 to the inlinerThreshold.
func (c *Compiler) Optimize(optLevel, sizeLevel int, inlinerThreshold uint) []error {
	builder := llvm.NewPassManagerBuilder()
	defer builder.Dispose()
	builder.SetOptLevel(optLevel)
	builder.SetSizeLevel(sizeLevel)
	if inlinerThreshold != 0 {
		builder.UseInlinerWithThreshold(inlinerThreshold)
	}
	builder.AddCoroutinePassesToExtensionPoints()

	if c.PanicStrategy() == "trap" {
		transform.ReplacePanicsWithTrap(c.mod) // -panic=trap
	}

	// run a check of all of our code
	if c.VerifyIR() {
		errs := c.checkModule()
		if errs != nil {
			return errs
		}
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
		goPasses.AddGlobalDCEPass()
		goPasses.AddGlobalOptimizerPass()
		goPasses.AddConstantPropagationPass()
		goPasses.AddAggressiveDCEPass()
		goPasses.AddFunctionAttrsPass()
		goPasses.Run(c.mod)

		// Run Go-specific optimization passes.
		transform.OptimizeMaps(c.mod)
		transform.OptimizeStringToBytes(c.mod)
		transform.OptimizeAllocs(c.mod)
		transform.LowerInterfaces(c.mod)

		errs := transform.LowerInterruptRegistrations(c.mod)
		if len(errs) > 0 {
			return errs
		}

		if c.funcImplementation() == funcValueSwitch {
			transform.LowerFuncValues(c.mod)
		}

		// After interfaces are lowered, there are many more opportunities for
		// interprocedural optimizations. To get them to work, function
		// attributes have to be updated first.
		goPasses.Run(c.mod)

		// Run TinyGo-specific interprocedural optimizations.
		transform.OptimizeAllocs(c.mod)
		transform.OptimizeStringToBytes(c.mod)

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
			return []error{err}
		}
	} else {
		// Must be run at any optimization level.
		transform.LowerInterfaces(c.mod)
		if c.funcImplementation() == funcValueSwitch {
			transform.LowerFuncValues(c.mod)
		}
		err := c.LowerGoroutines()
		if err != nil {
			return []error{err}
		}
		errs := transform.LowerInterruptRegistrations(c.mod)
		if len(errs) > 0 {
			return errs
		}
	}
	if c.VerifyIR() {
		if errs := c.checkModule(); errs != nil {
			return errs
		}
	}
	if err := c.Verify(); err != nil {
		return []error{errors.New("optimizations caused a verification failure")}
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
	for _, name := range c.getFunctionsUsedInTransforms() {
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

	hasGCPass := transform.AddGlobalsBitmap(c.mod)
	hasGCPass = transform.MakeGCStackSlots(c.mod) || hasGCPass
	if hasGCPass {
		if err := c.Verify(); err != nil {
			return []error{errors.New("GC pass caused a verification failure")}
		}
	}

	return nil
}
