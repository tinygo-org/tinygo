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
func (e *evalPackage) function(fn llvm.Value, params []Value, indent string) (Value, *Error) {
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
	if v.Type().TypeKind() != llvm.PointerTypeKind {
		return
	}
	v = unwrap(v)
	if v.IsAGlobalVariable().IsNil() {
		if !v.IsAUndefValue().IsNil() || !v.IsAConstantPointerNull().IsNil() {
			// Nothing to mark here: these definitely don't point to globals.
			return
		}
		panic("interp: trying to mark a non-global as dirty")
	}
	e.dirtyGlobals[v] = struct{}{}
}

// isDirty returns whether the value is tracked as part of the set of dirty
// globals or not. If it's not dirty, it means loads/stores can be done at
// compile time.
func (e *Eval) isDirty(v llvm.Value) bool {
	v = unwrap(v)
	if _, ok := e.dirtyGlobals[v]; ok {
		return true
	}
	return false
}
