// Package interp is a partial evaluator of code run at package init time. See
// the README in this package for details.
package interp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

// Enable extra checks, which should be disabled by default.
// This may help track down bugs by adding a few more sanity checks.
const checks = true

// runner contains all state related to one interp run.
type runner struct {
	mod               llvm.Module
	targetData        llvm.TargetData
	builder           llvm.Builder
	pointerSize       uint32                   // cached pointer size from the TargetData
	dataPtrType       llvm.Type                // often used type so created in advance
	uintptrType       llvm.Type                // equivalent to uintptr in Go
	maxAlign          int                      // maximum alignment of an object, alignment of runtime.alloc() result
	byteOrder         binary.ByteOrder         // big-endian or little-endian
	debug             bool                     // log debug messages
	pkgName           string                   // package name of the currently executing package
	functionCache     map[llvm.Value]*function // cache of compiled functions
	objects           []object                 // slice of objects in memory
	globals           map[llvm.Value]int       // map from global to index in objects slice
	start             time.Time
	timeout           time.Duration
	maxLoopIterations int
	callsExecuted     uint64
	xorshift32State   uint32 // state for runtime.fastrand
	xorshift64State   uint64 // state for runtime.fastrand64
}

// localsBufferSize is the initial capacity of the locals passed to run.
const localsBufferSize = 1 << 12

func newRunner(mod llvm.Module, timeout time.Duration, maxLoopIterations int, debug bool) *runner {
	r := runner{
		mod:               mod,
		targetData:        llvm.NewTargetData(mod.DataLayout()),
		byteOrder:         llvmutil.ByteOrder(mod.Target()),
		debug:             debug,
		functionCache:     make(map[llvm.Value]*function),
		objects:           []object{{}},
		globals:           make(map[llvm.Value]int),
		start:             time.Now(),
		timeout:           timeout,
		maxLoopIterations: maxLoopIterations,
		xorshift32State:   1,
		xorshift64State:   1,
	}
	r.pointerSize = uint32(r.targetData.PointerSize())
	r.dataPtrType = llvm.PointerType(mod.Context().Int8Type(), 0)
	r.uintptrType = mod.Context().IntType(r.targetData.PointerSize() * 8)
	r.maxAlign = r.targetData.PrefTypeAlignment(r.dataPtrType) // assume pointers are maximally aligned (this is not always the case)
	return &r
}

// Dispose deallocates all allocated LLVM resources.
func (r *runner) dispose() {
	r.targetData.Dispose()
	r.targetData = llvm.TargetData{}
}

// Run evaluates runtime.initAll function as much as possible at compile time.
// Set debug to true if it should print output while running.
func Run(mod llvm.Module, timeout time.Duration, maxLoopIterations int, debug bool) error {
	r := newRunner(mod, timeout, maxLoopIterations, debug)
	defer r.dispose()
	locals := make([]value, 0, localsBufferSize)

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
	if subprogram := initAll.Subprogram(); !subprogram.IsNil() {
		r.builder.SetCurrentDebugLocation(subprogram.SubprogramLine(), 0, subprogram, llvm.Metadata{})
	}
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
	for inst := range llvmutil.Instructions(bb) {
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

	// Whole-program Run checks hwInit0/hwInit1 after PrepareRuntimeStartup links
	// them to socInit*/boardInit*, while per-package RunFunc checks the hook names directly.
	if err := r.markExternalBootHooks(mod, "runtime.hwInit0", "runtime.hwInit1"); err != nil {
		return err
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
		_, mem, callErr := r.run(r.getFunction(fn), nil, nil, locals, 1)
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
			return callErr.materialize()
		}
		for index, obj := range mem.objects {
			r.objects[index] = obj
		}
		llvmutil.ReadFunctionUsageAttributes(initAll).Merge(mem.attributes).Apply(initAll)
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
func RunFunc(fn llvm.Value, timeout time.Duration, maxLoopIterations int, debug bool) error {
	// Create and initialize *runner object.
	mod := fn.GlobalParent()
	r := newRunner(mod, timeout, maxLoopIterations, debug)
	defer r.dispose()
	locals := make([]value, 0, localsBufferSize)
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

	if err := r.markExternalBootHooks(
		mod,
		"runtime.socInit0",
		"runtime.socInit1",
		"runtime.boardInit0",
		"runtime.boardInit1",
	); err != nil {
		return err
	}

	// Run the initializer, filling the .init.tmp function.
	if r.debug {
		fmt.Fprintln(os.Stderr, "interp:", fn.Name())
	}
	_, pkgMem, callErr := r.run(r.getFunction(fn), nil, nil, locals, 1)
	if callErr != nil {
		if isRecoverableError(callErr.Err) {
			// Could not finish, but could recover from it.
			if r.debug {
				fmt.Fprintln(os.Stderr, "not interpreting", r.pkgName, "because of error:", callErr.Error())
			}
			newFn.EraseFromParentAsFunction()
			return nil
		}
		return callErr.materialize()
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
	pkgMem.attributes.Apply(newFn)
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

func (r *runner) markExternalBootHooks(mod llvm.Module, names ...string) error {
	for _, name := range names {
		fn := mod.NamedFunction(name)
		if !fn.IsNil() && !fn.IsDeclaration() {
			if err := r.markExternalBootHook(fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// markExternalBootHook protects globals written by boot hooks without blocking
// interp from folding globals that the hooks only read.
func (r *runner) markExternalBootHook(fn llvm.Value) error {
	mem := memoryView{r: r}
	visited := map[llvm.Value]bool{}
	if err := mem.markBootHookFunc(fn, visited); err != nil {
		return err
	}
	for index, obj := range mem.objects {
		if obj.marked > r.objects[index].marked {
			r.objects[index].marked = obj.marked
		}
	}
	return nil
}

func (mv *memoryView) markBootHookFunc(fn llvm.Value, visited map[llvm.Value]bool) error {
	if fn.IsNil() || fn.IsDeclaration() || visited[fn] {
		return nil
	}
	visited[fn] = true
	for bb := range llvmutil.BasicBlocks(fn) {
		for inst := range llvmutil.Instructions(bb) {
			if err := mv.markBootHookInst(inst, visited); err != nil {
				return err
			}
		}
	}
	return nil
}

func (mv *memoryView) markBootHookInst(inst llvm.Value, visited map[llvm.Value]bool) error {
	if isBranch(inst) || inst.InstructionOpcode() == llvm.Switch {
		return nil
	}
	switch inst.InstructionOpcode() {
	case llvm.Load:
		if inst.Type().TypeKind() == llvm.PointerTypeKind && !isReadOnlyPtrUseChain(inst, map[llvm.Value]bool{}, true) {
			return mv.markBootHookValue(inst.Operand(0), visited)
		}
		return mv.markReadOnlyPointer(inst.Operand(0), visited)
	case llvm.ICmp:
		return nil
	case llvm.GetElementPtr, llvm.BitCast:
		if isReadOnlyPtrUseChain(inst, map[llvm.Value]bool{}, false) {
			return nil
		}
	case llvm.Call, llvm.Invoke:
		callee := inst.CalledValue()
		if callee.IsAInlineAsm().IsNil() {
			if callee.IsAFunction().IsNil() {
				return errorAt(inst, "interp: indirect call is not allowed in boot hook")
			}
			fnName := callee.Name()
			if fnName == "llvm.assume" || strings.HasPrefix(fnName, "llvm.dbg.") || strings.HasPrefix(fnName, "llvm.lifetime.") {
				return nil
			}
			if strings.HasPrefix(fnName, "interface:") {
				return errorAt(inst, "interp: interface call is not allowed in boot hook")
			}
			if isMemcpyOrMemmove(fnName) {
				if err := mv.markBootHookValue(inst.Operand(0), visited); err != nil {
					return err
				}
				return mv.markReadOnlyPointer(inst.Operand(1), visited)
			}
		}
		for i := range inst.OperandsCount() {
			op := inst.Operand(i)
			if !op.IsABasicBlock().IsNil() {
				continue
			}
			if !callee.IsAFunction().IsNil() && op == callee {
				continue
			}
			if err := mv.markBootHookValue(op, visited); err != nil {
				return err
			}
		}
		if !callee.IsAFunction().IsNil() {
			return mv.markBootHookFunc(callee, visited)
		}
		return mv.markBootHookValue(callee, visited)
	}

	for i := range inst.OperandsCount() {
		if err := mv.markBootHookValue(inst.Operand(i), visited); err != nil {
			return err
		}
	}
	return nil
}

func (mv *memoryView) markReadOnlyPointer(val llvm.Value, visited map[llvm.Value]bool) error {
	for {
		if val.IsUndef() || val.IsNull() {
			return nil
		}
		if !val.IsAInstruction().IsNil() {
			switch val.InstructionOpcode() {
			case llvm.GetElementPtr, llvm.BitCast:
				val = val.Operand(0)
				continue
			}
			return nil
		}
		if !val.IsAConstantExpr().IsNil() {
			switch val.Opcode() {
			case llvm.GetElementPtr, llvm.BitCast, llvm.IntToPtr, llvm.PtrToInt:
				val = val.Operand(0)
				continue
			}
		}
		break
	}
	if val.IsAGlobalVariable().IsNil() {
		return nil
	}
	objectIndex := mv.r.getValue(val).(pointerValue).index()
	obj := mv.get(objectIndex)
	if obj.buffer != nil {
		return mv.markExternalPointers(obj.buffer)
	}
	if initializer := val.Initializer(); !initializer.IsNil() {
		return mv.markBootHookValue(initializer, visited)
	}
	return nil
}

// markBootHookValue keeps reachable functions and constant descriptors in boot-hook mode.
func (mv *memoryView) markBootHookValue(llvmValue llvm.Value, visited map[llvm.Value]bool) error {
	if llvmValue.IsUndef() || llvmValue.IsNull() {
		return nil
	}
	if !llvmValue.IsAInstruction().IsNil() || !llvmValue.IsAArgument().IsNil() || !llvmValue.IsAInlineAsm().IsNil() {
		return nil
	}
	if !llvmValue.IsAGlobalValue().IsNil() {
		if !llvmValue.IsAFunction().IsNil() {
			return mv.markBootHookFunc(llvmValue, visited)
		}
		if !llvmValue.IsAGlobalVariable().IsNil() {
			if !llvmValue.IsGlobalConstant() {
				return mv.markExternalStore(llvmValue)
			}
			if visited[llvmValue] {
				return nil
			}
			visited[llvmValue] = true
			if initializer := llvmValue.Initializer(); !initializer.IsNil() {
				return mv.markBootHookValue(initializer, visited)
			}
			return nil
		}
		return nil
	}
	if !llvmValue.IsAConstantExpr().IsNil() {
		switch llvmValue.Opcode() {
		case llvm.IntToPtr, llvm.PtrToInt, llvm.BitCast, llvm.GetElementPtr:
			return mv.markBootHookValue(llvmValue.Operand(0), visited)
		case llvm.Add, llvm.Sub, llvm.Mul, llvm.UDiv, llvm.SDiv, llvm.URem, llvm.SRem, llvm.Shl, llvm.LShr, llvm.AShr, llvm.And, llvm.Or, llvm.Xor:
			if err := mv.markBootHookValue(llvmValue.Operand(0), visited); err != nil {
				return err
			}
			return mv.markBootHookValue(llvmValue.Operand(1), visited)
		default:
			return fmt.Errorf("interp: unknown constant expression '%s'", instructionNameMap[llvmValue.Opcode()])
		}
	}

	llvmType := llvmValue.Type()
	switch llvmType.TypeKind() {
	case llvm.IntegerTypeKind, llvm.FloatTypeKind, llvm.DoubleTypeKind, llvm.VectorTypeKind:
		return nil
	case llvm.StructTypeKind:
		numElements := llvmType.StructElementTypesCount()
		for i := range numElements {
			if err := mv.markBootHookValue(llvmValue.AggregateElement(i), visited); err != nil {
				return err
			}
		}
		return nil
	case llvm.ArrayTypeKind:
		numElements := llvmType.ArrayLength()
		for i := range numElements {
			if err := mv.markBootHookValue(llvmValue.AggregateElement(i), visited); err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("interp: unknown type kind in markBootHookValue")
	}
}

func isReadOnlyPtrUseChain(val llvm.Value, seen map[llvm.Value]bool, allowVolatileStore bool) bool {
	if seen[val] {
		return true
	}
	seen[val] = true
	for user := range llvmutil.Uses(val) {
		if user.IsAInstruction().IsNil() {
			return false
		}
		switch user.InstructionOpcode() {
		case llvm.Load:
			if user.Operand(0) != val {
				return false
			}
		case llvm.Store:
			// Volatile stores target MMIO registers when reached through a loaded pointer.
			if !allowVolatileStore || !user.IsVolatile() || user.Operand(1) != val || user.Operand(0) == val {
				return false
			}
		case llvm.ICmp:
			// Pointer comparisons (such as nil checks) never mutate or escape the pointee.
		case llvm.GetElementPtr, llvm.BitCast:
			if user.Operand(0) != val || !isReadOnlyPtrUseChain(user, seen, allowVolatileStore) {
				return false
			}
		case llvm.Call:
			callee := user.CalledValue()
			if callee.IsAFunction().IsNil() || !isMemcpyOrMemmove(callee.Name()) {
				return false
			}
			// Only source operand 1 is read-only; aliasing with dst operand 0 writes.
			if user.Operand(0) == val || user.Operand(1) != val {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func isMemcpyOrMemmove(name string) bool {
	return strings.HasPrefix(name, "llvm.memcpy.") || strings.HasPrefix(name, "llvm.memmove.")
}

// These must stay the same as the generators in src/runtime/algorithm.go, so
// that interp gives the same values as the runtime.
func xorshift32(x uint32) uint32 {
	x ^= x << 7
	x ^= x >> 1
	x ^= x << 9
	return x
}

func xorshiftMult64(x uint64) uint64 {
	x ^= x >> 12
	x ^= x << 25
	x ^= x >> 27
	return x * 2685821657736338717
}
