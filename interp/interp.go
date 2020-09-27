// Package interp is a partial evaluator of code run at package init time. See
// the README in this package for details.
package interp

import (
	"fmt"
	"os"
	"strings"
	"time"

	"tinygo.org/x/go-llvm"
)

// Enable extra checks, which should be disabled by default.
// This may help track down bugs by adding a few more sanity checks.
const checks = true

// runner contains all state related to one interp run.
type runner struct {
	mod           llvm.Module
	targetData    llvm.TargetData
	builder       llvm.Builder
	pointerSize   uint32                   // cached pointer size from the TargetData
	debug         bool                     // log debug messages
	pkgName       string                   // package name of the currently executing package
	functionCache map[llvm.Value]*function // cache of compiled functions
	objects       []object                 // slice of objects in memory
	globals       map[llvm.Value]int       // map from global to index in objects slice
	start         time.Time
	callsExecuted uint64
}

// Run evaluates runtime.initAll function as much as possible at compile time.
// Set debug to true if it should print output while running.
func Run(mod llvm.Module, debug bool) error {
	r := runner{
		mod:           mod,
		targetData:    llvm.NewTargetData(mod.DataLayout()),
		debug:         debug,
		functionCache: make(map[llvm.Value]*function),
		objects:       []object{{}},
		globals:       make(map[llvm.Value]int),
		start:         time.Now(),
	}
	r.pointerSize = uint32(r.targetData.PointerSize())

	initAll := mod.NamedFunction("runtime.initAll")
	bb := initAll.EntryBasicBlock()

	// Create a builder, to insert instructions that could not be evaluated at
	// compile time.
	r.builder = mod.Context().NewBuilder()
	defer r.builder.Dispose()

	// Create a dummy alloca in the entry block that we can set the insert point
	// to. This is necessary because otherwise we might be removing the
	// instruction (init call) that we are removing after successful
	// interpretation.
	r.builder.SetInsertPointBefore(bb.FirstInstruction())
	dummy := r.builder.CreateAlloca(r.mod.Context().Int8Type(), "dummy")
	r.builder.SetInsertPointBefore(dummy)
	defer dummy.EraseFromParentAsInstruction()

	// Get a list if init calls. A runtime.initAll might look something like this:
	// func initAll() {
	//     unsafe.init()
	//     machine.init()
	//     runtime.init()
	// }
	// This function gets a list of these call instructions.
	var initCalls []llvm.Value
	for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		if inst == dummy {
			continue
		}
		if !inst.IsAReturnInst().IsNil() {
			break // ret void
		}
		if inst.IsACallInst().IsNil() || inst.CalledValue().IsAFunction().IsNil() {
			return errorAt(inst, "interp: expected all instructions in "+initAll.Name()+" to be direct calls")
		}
		initCalls = append(initCalls, inst)
	}

	// Run initializers for each package. Once the package initializer is
	// finished, the call to the package initializer can be removed.
	for _, call := range initCalls {
		initName := call.CalledValue().Name()
		if !strings.HasSuffix(initName, ".init") {
			return errorAt(call, "interp: expected all instructions in "+initAll.Name()+" to be *.init() calls")
		}
		r.pkgName = initName[:len(initName)-len(".init")]
		fn := call.CalledValue()
		if r.debug {
			fmt.Fprintln(os.Stderr, "call:", fn.Name())
		}
		_, mem, callErr := r.run(r.getFunction(fn), nil, nil, "    ")
		if callErr != nil {
			if isRecoverableError(callErr.Err) {
				if r.debug {
					fmt.Fprintln(os.Stderr, "not interpretring", r.pkgName, "because of error:", callErr.Err)
				}
				mem.revert()
				continue
			}
			return callErr
		}
		call.EraseFromParentAsInstruction()
		for index, obj := range mem.objects {
			r.objects[index] = obj
		}
	}
	r.pkgName = ""

	// Update all global variables in the LLVM module.
	mem := memoryView{r: &r}
	for _, obj := range r.objects {
		if obj.llvmGlobal.IsNil() {
			continue
		}
		if obj.buffer == nil {
			continue
		}
		initializer := obj.buffer.toLLVMValue(obj.llvmGlobal.Type().ElementType(), &mem)
		if checks && initializer.Type() != obj.llvmGlobal.Type().ElementType() {
			panic("initializer type mismatch")
		}
		obj.llvmGlobal.SetInitializer(initializer)
	}

	return nil
}

// getFunction returns the compiled version of the given LLVM function. It
// compiles the function if necessary and caches the result.
func (r *runner) getFunction(llvmFn llvm.Value) *function {
	if fn, ok := r.functionCache[llvmFn]; ok {
		return fn
	}
	fn := r.compileFunction(llvmFn)
	r.functionCache[llvmFn] = fn
	return fn
}
