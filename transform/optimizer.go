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

// OptimizePackage runs optimization passes over the LLVM module for the given
// Go package.
func OptimizePackage(mod llvm.Module, config *compileopts.Config) {
	_, speedLevel, _ := config.OptLevel()

	// Run TinyGo-specific optimization passes.
	if speedLevel > 0 {
		OptimizeMaps(mod)
	}
}

// Optimize runs a number of optimization and transformation passes over the
// given module. Some passes are specific to TinyGo, others are generic LLVM
// passes.
//
// Please note that some optimizations are not optional, thus Optimize must
// always be run before emitting machine code.
func Optimize(mod llvm.Module, config *compileopts.Config) []error {
	optLevel, speedLevel, _ := config.OptLevel()

	// Make sure these functions are kept in tact during TinyGo transformation passes.
	for _, name := range functionsUsedInTransforms {
		fn := mod.NamedFunction(name)
		if fn.IsNil() {
			panic(fmt.Errorf("missing core function %q", name))
		}
		fn.SetLinkage(llvm.ExternalLinkage)
	}

	// run a check of all of our code
	if config.VerifyIR() {
		errs := ircheck.Module(mod)
		if errs != nil {
			return errs
		}
	}

	if speedLevel > 0 {
		// Run some preparatory passes for the Go optimizer.
		po := llvm.NewPassBuilderOptions()
		defer po.Dispose()
		err := mod.RunPasses("globaldce,globalopt,ipsccp,instcombine,adce,function-attrs", llvm.TargetMachine{}, po)
		if err != nil {
			return []error{fmt.Errorf("could not build pass pipeline: %w", err)}
		}

		// Run TinyGo-specific optimization passes.
		OptimizeStringToBytes(mod)
		OptimizeReflectImplements(mod)
		maxStackSize := config.MaxStackAlloc()
		OptimizeAllocs(mod, nil, maxStackSize, nil)
		err = LowerInterfaces(mod, config)
		if err != nil {
			return []error{err}
		}

		errs := LowerInterrupts(mod)
		if len(errs) > 0 {
			return errs
		}

		// After interfaces are lowered, there are many more opportunities for
		// interprocedural optimizations. To get them to work, function
		// attributes have to be updated first.
		err = mod.RunPasses("globaldce,globalopt,ipsccp,instcombine,adce,function-attrs", llvm.TargetMachine{}, po)
		if err != nil {
			return []error{fmt.Errorf("could not build pass pipeline: %w", err)}
		}

		// Run TinyGo-specific interprocedural optimizations.
		OptimizeAllocs(mod, config.Options.PrintAllocs, maxStackSize, func(pos token.Position, msg string) {
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
		errs := LowerInterrupts(mod)
		if len(errs) > 0 {
			return errs
		}

		// Clean up some leftover symbols of the previous transformations.
		po := llvm.NewPassBuilderOptions()
		defer po.Dispose()
		err = mod.RunPasses("globaldce", llvm.TargetMachine{}, po)
		if err != nil {
			return []error{fmt.Errorf("could not build pass pipeline: %w", err)}
		}
	}

	if config.Scheduler() == "none" {
		// Check for any goroutine starts.
		if start := mod.NamedFunction("internal/task.start"); !start.IsNil() && len(getUses(start)) > 0 {
			errs := []error{}
			for _, call := range getUses(start) {
				errs = append(errs, errorAt(call, "attempted to start a goroutine without a scheduler"))
			}
			return errs
		}
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
	for _, name := range functionsUsedInTransforms {
		fn := mod.NamedFunction(name)
		if fn.IsNil() || fn.IsDeclaration() {
			continue
		}
		fn.SetLinkage(llvm.InternalLinkage)
	}

	// Run the default pass pipeline.
	// TODO: set the PrepareForThinLTO flag somehow.
	po := llvm.NewPassBuilderOptions()
	defer po.Dispose()
	passes := fmt.Sprintf("default<%s>", optLevel)
	err := mod.RunPasses(passes, llvm.TargetMachine{}, po)
	if err != nil {
		return []error{fmt.Errorf("could not build pass pipeline: %w", err)}
	}

	hasGCPass := MakeGCStackSlots(mod)
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
