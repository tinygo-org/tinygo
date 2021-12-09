package transform

import (
	"errors"
	"fmt"
	"go/token"
	"os"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler/ircheck"
	"tinygo.org/x/go-llvm"
)

// Optimize runs a number of optimization and transformation passes over the
// given module. Some passes are specific to TinyGo, others are generic LLVM
// passes. You can set a preferred performance (0-3) and size (0-2) level and
// control the limits of the inliner (higher numbers mean more inlining, set it
// to 0 to disable entirely).
//
// Please note that some optimizations are not optional, thus Optimize must
// alwasy be run before emitting machine code. Set all controls (optLevel,
// sizeLevel, inlinerThreshold) to 0 to reduce the number of optimizations to a
// minimum.
func Optimize(mod llvm.Module, config *compileopts.Config, optLevel, sizeLevel int, inlinerThreshold uint) []error {
	builder := llvm.NewPassManagerBuilder()
	defer builder.Dispose()
	builder.SetOptLevel(optLevel)
	builder.SetSizeLevel(sizeLevel)
	if inlinerThreshold != 0 {
		builder.UseInlinerWithThreshold(inlinerThreshold)
	}
	builder.AddCoroutinePassesToExtensionPoints()

	// Make sure these functions are kept in tact during TinyGo transformation passes.
	for _, name := range getFunctionsUsedInTransforms(config) {
		fn := mod.NamedFunction(name)
		if fn.IsNil() {
			panic(fmt.Errorf("missing core function %q", name))
		}
		fn.SetLinkage(llvm.ExternalLinkage)
	}

	if config.PanicStrategy() == "trap" {
		ReplacePanicsWithTrap(mod) // -panic=trap
	}

	// run a check of all of our code
	if config.VerifyIR() {
		errs := ircheck.Module(mod)
		if errs != nil {
			return errs
		}
	}

	if optLevel > 0 {
		// Run some preparatory passes for the Go optimizer.
		goPasses := llvm.NewPassManager()
		defer goPasses.Dispose()
		goPasses.AddGlobalDCEPass()
		goPasses.AddGlobalOptimizerPass()
		goPasses.AddIPSCCPPass()
		goPasses.AddInstructionCombiningPass() // necessary for OptimizeReflectImplements
		goPasses.AddAggressiveDCEPass()
		goPasses.AddFunctionAttrsPass()
		goPasses.Run(mod)

		// Run TinyGo-specific optimization passes.
		OptimizeMaps(mod)
		OptimizeStringToBytes(mod)
		OptimizeReflectImplements(mod)
		OptimizeAllocs(mod, nil, nil)
		err := LowerInterfaces(mod, config)
		if err != nil {
			return []error{err}
		}

		errs := LowerInterrupts(mod)
		if len(errs) > 0 {
			return errs
		}

		if config.FuncImplementation() == "switch" {
			LowerFuncValues(mod)
		}

		// After interfaces are lowered, there are many more opportunities for
		// interprocedural optimizations. To get them to work, function
		// attributes have to be updated first.
		goPasses.Run(mod)

		// Run TinyGo-specific interprocedural optimizations.
		LowerReflect(mod)
		OptimizeAllocs(mod, config.Options.PrintAllocs, func(pos token.Position, msg string) {
			fmt.Fprintln(os.Stderr, pos.String()+": "+msg)
		})
		OptimizeStringToBytes(mod)
		OptimizeStringEqual(mod)

	} else {
		// Must be run at any optimization level.
		err := LowerInterfaces(mod, config)
		if err != nil {
			return []error{err}
		}
		LowerReflect(mod)
		if config.FuncImplementation() == "switch" {
			LowerFuncValues(mod)
		}
		errs := LowerInterrupts(mod)
		if len(errs) > 0 {
			return errs
		}

		// Clean up some leftover symbols of the previous transformations.
		goPasses := llvm.NewPassManager()
		defer goPasses.Dispose()
		goPasses.AddGlobalDCEPass()
		goPasses.Run(mod)
	}

	// Lower async implementations.
	switch config.Scheduler() {
	case "coroutines":
		// Lower async as coroutines.
		err := LowerCoroutines(mod, config.NeedsStackObjects())
		if err != nil {
			return []error{err}
		}
	case "tasks", "asyncify":
		// No transformations necessary.
	case "none":
		// Check for any goroutine starts.
		if start := mod.NamedFunction("internal/task.start"); !start.IsNil() && len(getUses(start)) > 0 {
			errs := []error{}
			for _, call := range getUses(start) {
				errs = append(errs, errorAt(call, "attempted to start a goroutine without a scheduler"))
			}
			return errs
		}
	default:
		return []error{errors.New("invalid scheduler")}
	}

	if config.VerifyIR() {
		if errs := ircheck.Module(mod); errs != nil {
			return errs
		}
	}
	if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
		return []error{errors.New("optimizations caused a verification failure")}
	}

	// After TinyGo-specific transforms have finished, undo exporting these functions.
	for _, name := range getFunctionsUsedInTransforms(config) {
		fn := mod.NamedFunction(name)
		if fn.IsNil() || fn.IsDeclaration() {
			continue
		}
		fn.SetLinkage(llvm.InternalLinkage)
	}

	// Run function passes again, because without it, llvm.coro.size.i32()
	// doesn't get lowered.
	funcPasses := llvm.NewFunctionPassManagerForModule(mod)
	defer funcPasses.Dispose()
	builder.PopulateFunc(funcPasses)
	funcPasses.InitializeFunc()
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		funcPasses.RunFunc(fn)
	}
	funcPasses.FinalizeFunc()

	// Run module passes.
	modPasses := llvm.NewPassManager()
	defer modPasses.Dispose()
	builder.Populate(modPasses)
	modPasses.Run(mod)

	hasGCPass := AddGlobalsBitmap(mod)
	hasGCPass = MakeGCStackSlots(mod) || hasGCPass
	if hasGCPass {
		if err := llvm.VerifyModule(mod, llvm.PrintMessageAction); err != nil {
			return []error{errors.New("GC pass caused a verification failure")}
		}
	}

	return nil
}

// functionsUsedInTransform is a list of function symbols that may be used
// during TinyGo optimization passes so they have to be marked as external
// linkage until all TinyGo passes have finished.
var functionsUsedInTransforms = []string{
	"runtime.alloc",
	"runtime.free",
	"runtime.nilPanic",
}

var taskFunctionsUsedInTransforms = []string{}

// These functions need to be preserved in the IR until after the coroutines
// pass has run.
var coroFunctionsUsedInTransforms = []string{
	"internal/task.start",
	"internal/task.Pause",
	"internal/task.fake",
	"internal/task.Current",
	"internal/task.createTask",
	"(*internal/task.Task).setState",
	"(*internal/task.Task).returnTo",
	"(*internal/task.Task).returnCurrent",
	"(*internal/task.Task).setReturnPtr",
	"(*internal/task.Task).getReturnPtr",
}

// getFunctionsUsedInTransforms gets a list of all special functions that should be preserved during transforms and optimization.
func getFunctionsUsedInTransforms(config *compileopts.Config) []string {
	fnused := functionsUsedInTransforms
	switch config.Scheduler() {
	case "none":
	case "coroutines":
		fnused = append(append([]string{}, fnused...), coroFunctionsUsedInTransforms...)
	case "tasks", "asyncify":
		fnused = append(append([]string{}, fnused...), taskFunctionsUsedInTransforms...)
	default:
		panic(fmt.Errorf("invalid scheduler %q", config.Scheduler()))
	}
	return fnused
}
