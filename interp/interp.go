// Package interp interprets Go package initializers as much as possible. This
// avoid running them at runtime, improving code size and making other
// optimizations possible.
package interp

// This file provides the overarching Eval object with associated (utility)
// methods.

import (
	"strings"

	"tinygo.org/x/go-llvm"
)

type Eval struct {
	Mod             llvm.Module
	TargetData      llvm.TargetData
	Debug           bool
	builder         llvm.Builder
	dirtyGlobals    map[llvm.Value]struct{}
	sideEffectFuncs map[llvm.Value]*sideEffectResult // cache of side effect scan results
}

// evalPackage encapsulates the Eval type for just a single package. The Eval
// type keeps state across the whole program, the evalPackage type keeps extra
// state for the currently interpreted package.
type evalPackage struct {
	*Eval
	packagePath string
	errors      []error
}

// Run evaluates the function with the given name and then eliminates all
// callers.
func Run(mod llvm.Module, debug bool) error {
	if debug {
		println("\ncompile-time evaluation:")
	}

	name := "runtime.initAll"
	e := &Eval{
		Mod:          mod,
		TargetData:   llvm.NewTargetData(mod.DataLayout()),
		Debug:        debug,
		dirtyGlobals: map[llvm.Value]struct{}{},
	}
	e.builder = mod.Context().NewBuilder()

	initAll := mod.NamedFunction(name)
	bb := initAll.EntryBasicBlock()
	// Create a dummy alloca in the entry block that we can set the insert point
	// to. This is necessary because otherwise we might be removing the
	// instruction (init call) that we are removing after successful
	// interpretation.
	e.builder.SetInsertPointBefore(bb.FirstInstruction())
	dummy := e.builder.CreateAlloca(e.Mod.Context().Int8Type(), "dummy")
	e.builder.SetInsertPointBefore(dummy)
	var initCalls []llvm.Value
	for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		if inst == dummy {
			continue
		}
		if !inst.IsAReturnInst().IsNil() {
			break // ret void
		}
		if inst.IsACallInst().IsNil() || inst.CalledValue().IsAFunction().IsNil() {
			return errorAt(inst, "interp: expected all instructions in "+name+" to be direct calls")
		}
		initCalls = append(initCalls, inst)
	}

	// Do this in a separate step to avoid corrupting the iterator above.
	undefPtr := llvm.Undef(llvm.PointerType(mod.Context().Int8Type(), 0))
	for _, call := range initCalls {
		initName := call.CalledValue().Name()
		if !strings.HasSuffix(initName, ".init") {
			return errorAt(call, "interp: expected all instructions in "+name+" to be *.init() calls")
		}
		pkgName := initName[:len(initName)-5]
		fn := call.CalledValue()
		call.EraseFromParentAsInstruction()
		evalPkg := evalPackage{
			Eval:        e,
			packagePath: pkgName,
		}
		_, err := evalPkg.function(fn, []Value{&LocalValue{e, undefPtr}, &LocalValue{e, undefPtr}}, "")
		if err == errUnreachable {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// function interprets the given function. The params are the function params
// and the indent is the string indentation to use when dumping all interpreted
// instructions.
func (e *evalPackage) function(fn llvm.Value, params []Value, indent string) (Value, error) {
	fr := frame{
		evalPackage: e,
		fn:          fn,
		locals:      make(map[llvm.Value]Value),
	}
	for i, param := range fn.Params() {
		fr.locals[param] = params[i]
	}

	bb := fn.EntryBasicBlock()
	var lastBB llvm.BasicBlock
	for {
		retval, outgoing, err := fr.evalBasicBlock(bb, lastBB, indent)
		if outgoing == nil {
			// returned something (a value or void, or an error)
			return retval, err
		}
		if len(outgoing) > 1 {
			panic("unimplemented: multiple outgoing blocks")
		}
		next := outgoing[0]
		if next.IsABasicBlock().IsNil() {
			panic("did not switch to a basic block")
		}
		lastBB = bb
		bb = next.AsBasicBlock()
	}
}

// getValue determines what kind of LLVM value it gets and returns the
// appropriate Value type.
func (e *Eval) getValue(v llvm.Value) Value {
	return &LocalValue{e, v}
}

// markDirty marks the passed-in LLVM value dirty, recursively. For example,
// when it encounters a constant GEP on a global, it marks the global dirty.
func (e *Eval) markDirty(v llvm.Value) {
	if !v.IsAGlobalVariable().IsNil() {
		if v.IsGlobalConstant() {
			return
		}
		if _, ok := e.dirtyGlobals[v]; !ok {
			e.dirtyGlobals[v] = struct{}{}
			e.sideEffectFuncs = nil // re-calculate all side effects
		}
	} else if v.IsConstant() {
		if v.OperandsCount() >= 2 && !v.Operand(0).IsAGlobalVariable().IsNil() {
			// looks like a constant getelementptr of a global.
			// TODO: find a way to make sure it really is: v.Opcode() returns 0.
			e.markDirty(v.Operand(0))
			return
		}
		return // nothing to mark
	} else if !v.IsAGetElementPtrInst().IsNil() {
		panic("interp: todo: GEP")
	} else {
		// Not constant and not a global or GEP so doesn't have to be marked
		// non-constant.
	}
}
