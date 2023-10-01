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
	dataPtrType   llvm.Type                // often used type so created in advance
	uintptrType   llvm.Type                // equivalent to uintptr in Go
	maxAlign      int                      // maximum alignment of an object, alignment of runtime.alloc() result
	debug         bool                     // log debug messages
	pkgName       string                   // package name of the currently executing package
	functionCache map[llvm.Value]*function // cache of compiled functions
	objects       []object                 // slice of objects in memory
	globals       map[llvm.Value]int       // map from global to index in objects slice
	start         time.Time
	timeout       time.Duration
	maxDepth      int
	maxInstr      int
	callsExecuted uint64
}

func newRunner(mod llvm.Module, timeout time.Duration, maxDepth int, maxInstr int, debug bool) *runner {
	r := runner{
		mod:           mod,
		targetData:    llvm.NewTargetData(mod.DataLayout()),
		debug:         debug,
		functionCache: make(map[llvm.Value]*function),
		objects:       []object{{}},
		globals:       make(map[llvm.Value]int),
		start:         time.Now(),
		timeout:       timeout,
		maxDepth:      maxDepth,
		maxInstr:      maxInstr,
	}
	r.pointerSize = uint32(r.targetData.PointerSize())
	r.dataPtrType = llvm.PointerType(mod.Context().Int8Type(), 0)
	r.uintptrType = mod.Context().IntType(r.targetData.PointerSize() * 8)
	r.maxAlign = r.targetData.PrefTypeAlignment(r.dataPtrType) // assume pointers are maximally aligned (this is not always the case)
	return &r
}

// Dispose deallocates all alloated LLVM resources.
func (r *runner) dispose() {
	r.targetData.Dispose()
	r.targetData = llvm.TargetData{}
}

// Run evaluates runtime.initAll function as much as possible at compile time.
// Set debug to true if it should print output while running.
func Run(mod llvm.Module, timeout time.Duration, maxDepth int, maxInstr int, debug bool) error {
	r := newRunner(mod, timeout, maxDepth, maxInstr, debug)
	defer r.dispose()

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
		_, mem, callErr := r.run(r.getFunction(fn), nil, nil, 0, "    ")
		call.EraseFromParentAsInstruction()
		if callErr != nil {
			if isRecoverableError(callErr.Err) {
				if r.debug {
					fmt.Fprintln(os.Stderr, "not interpreting", r.pkgName, "because of error:", callErr.Error())
				}
				// Remove instructions that were created as part of interpreting
				// the package.
				mem.revert()
				// Create a call to the package initializer (which was
				// previously deleted).
				i8undef := llvm.Undef(r.dataPtrType)
				r.builder.CreateCall(fn.GlobalValueType(), fn, []llvm.Value{i8undef}, "")
				// Make sure that any globals touched by the package
				// initializer, won't be accessed by later package initializers.
				err := r.markExternalLoad(fn)
				if err != nil {
					return fmt.Errorf("failed to interpret package %s: %w", r.pkgName, err)
				}
				continue
			}
			return callErr
		}
		for index, obj := range mem.objects {
			r.objects[index] = obj
		}
	}
	r.pkgName = ""

	// Update all global variables in the LLVM module.
	mem := memoryView{r: r}
	for i, obj := range r.objects {
		if obj.llvmGlobal.IsNil() {
			continue
		}
		if obj.buffer == nil {
			continue
		}
		if obj.constant {
			continue // constant buffers can't have been modified
		}
		initializer, err := obj.buffer.toLLVMValue(obj.llvmGlobal.GlobalValueType(), &mem)
		if err == errInvalidPtrToIntSize {
			// This can happen when a previous interp run did not have the
			// correct LLVM type for a global and made something up. In that
			// case, some fields could be written out as a series of (null)
			// bytes even though they actually contain a pointer value.
			// As a fallback, use asRawValue to get something of the correct
			// memory layout.
			initializer, err := obj.buffer.asRawValue(r).rawLLVMValue(&mem)
			if err != nil {
				return err
			}
			initializerType := initializer.Type()
			newGlobal := llvm.AddGlobal(mod, initializerType, obj.llvmGlobal.Name()+".tmp")
			newGlobal.SetInitializer(initializer)
			newGlobal.SetLinkage(obj.llvmGlobal.Linkage())
			newGlobal.SetAlignment(obj.llvmGlobal.Alignment())
			// TODO: copy debug info, unnamed_addr, ...
			obj.llvmGlobal.ReplaceAllUsesWith(newGlobal)
			name := obj.llvmGlobal.Name()
			obj.llvmGlobal.EraseFromParentAsGlobal()
			newGlobal.SetName(name)

			// Update interp-internal references.
			delete(r.globals, obj.llvmGlobal)
			obj.llvmGlobal = newGlobal
			r.globals[newGlobal] = i
			r.objects[i] = obj
			continue
		}
		if err != nil {
			return err
		}
		if checks && initializer.Type() != obj.llvmGlobal.GlobalValueType() {
			panic("initializer type mismatch")
		}
		obj.llvmGlobal.SetInitializer(initializer)
	}

	return nil
}

// RunFunc evaluates a single package initializer at compile time.
// Set debug to true if it should print output while running.
func RunFunc(fn llvm.Value, timeout time.Duration, maxDepth int, maxInstr int, debug bool) error {
	// Create and initialize *runner object.
	mod := fn.GlobalParent()
	r := newRunner(mod, timeout, maxDepth, maxInstr, debug)
	defer r.dispose()
	initName := fn.Name()
	if !strings.HasSuffix(initName, ".init") {
		return errorAt(fn, "interp: unexpected function name (expected *.init)")
	}
	r.pkgName = initName[:len(initName)-len(".init")]

	// Create new function with the interp result.
	newFn := llvm.AddFunction(mod, fn.Name()+".tmp", fn.GlobalValueType())
	newFn.SetLinkage(fn.Linkage())
	newFn.SetVisibility(fn.Visibility())
	entry := mod.Context().AddBasicBlock(newFn, "entry")

	// Create a builder, to insert instructions that could not be evaluated at
	// compile time.
	r.builder = mod.Context().NewBuilder()
	defer r.builder.Dispose()
	r.builder.SetInsertPointAtEnd(entry)

	// Copy debug information.
	subprogram := fn.Subprogram()
	if !subprogram.IsNil() {
		newFn.SetSubprogram(subprogram)
		r.builder.SetCurrentDebugLocation(subprogram.SubprogramLine(), 0, subprogram, llvm.Metadata{})
	}

	// Run the initializer, filling the .init.tmp function.
	if r.debug {
		fmt.Fprintln(os.Stderr, "interp:", fn.Name())
	}
	_, pkgMem, callErr := r.run(r.getFunction(fn), nil, nil, 0, "    ")
	if callErr != nil {
		if isRecoverableError(callErr.Err) {
			// Could not finish, but could recover from it.
			if r.debug {
				fmt.Fprintln(os.Stderr, "not interpreting", r.pkgName, "because of error:", callErr.Error())
			}
			newFn.EraseFromParentAsFunction()
			return nil
		}
		return callErr
	}
	for index, obj := range pkgMem.objects {
		r.objects[index] = obj
	}

	// Update globals with values determined while running the initializer above.
	mem := memoryView{r: r}
	for _, obj := range r.objects {
		if obj.llvmGlobal.IsNil() {
			continue
		}
		if obj.buffer == nil {
			continue
		}
		if obj.constant {
			continue // constant, so can't have been modified
		}
		initializer, err := obj.buffer.toLLVMValue(obj.llvmGlobal.GlobalValueType(), &mem)
		if err != nil {
			return err
		}
		if checks && initializer.Type() != obj.llvmGlobal.GlobalValueType() {
			panic("initializer type mismatch")
		}
		obj.llvmGlobal.SetInitializer(initializer)
	}

	// Finalize: remove the old init function and replace it with the new
	// (.init.tmp) function.
	r.builder.CreateRetVoid()
	fnName := fn.Name()
	fn.ReplaceAllUsesWith(newFn)
	fn.EraseFromParentAsFunction()
	newFn.SetName(fnName)

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

// markExternalLoad marks the given llvmValue as being loaded externally. This
// is primarily used to mark package initializers that could not be run at
// compile time. As an example, a package initialize might store to a global
// variable. Another package initializer might read from the same global
// variable. By marking this function as being run at runtime, that load
// instruction will need to be run at runtime instead of at compile time.
func (r *runner) markExternalLoad(llvmValue llvm.Value) error {
	mem := memoryView{r: r}
	err := mem.markExternalLoad(llvmValue)
	if err != nil {
		return err
	}
	for index, obj := range mem.objects {
		if obj.marked > r.objects[index].marked {
			r.objects[index].marked = obj.marked
		}
	}
	return nil
}
