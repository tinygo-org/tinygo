package interp

import (
	"errors"

	"tinygo.org/x/go-llvm"
)

const (
	minStackSize = 1 << 10
	maxStackSize = 1 << 24
)

const debug = false

func Run(mod llvm.Module, fn llvm.Value) error {
	if fn.IsNil() {
		return errors.New("function to run is nil")
	}

	if debug {
		println("run fn:", fn.Name())
		println("source:")
		mod.Dump()
	}

	// Set up execution state.
	params := fn.Params()
	state := execState{
		cp: constParser{
			td:     llvm.NewTargetData(mod.DataLayout()),
			tCache: make(map[llvm.Type]typ),
			vCache: make(map[llvm.Value]value),
			fCache: make(map[llvm.Type]fnTyInfo),
		},
		stack:       make([]value, minStackSize)[:0],
		version:     memVersionStart,
		nextVersion: memVersionStart + 1,
		rt: builder{
			stackHeight: uint(len(params)),
		},
	}
	err := state.cp.mapGlobals(mod)
	if err != nil {
		if isRuntimeOrRevert(err) {
			return nil
		}
		return err
	}
	for _, g := range state.cp.globals {
		if g.escaped {
			state.escapeStack = append(state.escapeStack, g)
		}
	}
	sortObjects(state.escapeStack)
	err = state.invalidateEscaped(llvm.Metadata{})
	if err != nil {
		if isRuntimeOrRevert(err) {
			return nil
		}
		return err
	}

	// Compile the function.
	fnObj := state.cp.globals[fn]
	err = fnObj.compile(&state.cp)
	if err != nil {
		if isRuntimeOrRevert(err) {
			return nil
		}
		return err
	}
	if fnObj.sig.ret != nil {
		return errors.New("unexprected return from init function")
	}

	// Run the function.
	state.curFn = fnObj.r
	state.escBuf = make(map[*memObj]struct{})
	var args []value
	for i, a := range fnObj.sig.args {
		args = unpack(runtime(a.t, uint(i)), args)
		state.stack = append(state.stack, runtime(a.t, uint(i)))
	}
	if len(fnObj.sig.args) != len(params) {
		panic("arg len not equal to param len")
	}
	err = (&callInst{fnObj.ptr(0), args, fnObj.sig, llvm.Metadata{}, true}).exec(&state)
	if err != nil {
		if isRuntimeOrRevert(err) {
			return nil
		}
		return err
	}
	state.escape(fnObj.ptr(0))
	err = state.flushEscaped(llvm.Metadata{})
	if err != nil {
		if isRuntimeOrRevert(err) {
			return nil
		}
		return err
	}
	if debug {
		println("runtime:")
		println("\targs:", len(params))
		height := uint(len(args))
		for _, inst := range state.rt.instrs {
			resStr := ""
			if res := inst.result(); res != nil {
				var v value
				v, height = createDecomposedIndices(height, res)
				resStr = v.String() + " = "
			}
			println("\t" + resStr + inst.String())
		}
	}

	// Emit runtime instructions.
	ctx := mod.Context()
	gen := rtGen{
		ctx:        ctx,
		modGlobals: make(map[*memObj]struct{}),
		ptrs:       make(map[ptkey]llvm.Type),
		iTypes:     make(map[iType]llvm.Type),
		sTypes:     make(map[*structType]llvm.Type),
		aTypes:     make(map[atkey]llvm.Type),
		stack:      append(make([]llvm.Value, minStackSize)[:0], fn.Params()...),
		vals:       make(map[vkey]llvm.Value, minStackSize),
		builder:    ctx.NewBuilder(),
	}
	oldBlocks := fn.BasicBlocks()
	if subprogram := fn.Subprogram(); !subprogram.IsNil() {
		gen.builder.SetCurrentDebugLocation(subprogram.SubprogramLine(), 0, subprogram, llvm.Metadata{})
	}
	b := ctx.AddBasicBlock(fn, "interpreted")
	var ok bool
	defer func() {
		if ok {
			for _, b := range oldBlocks {
				b.EraseFromParent()
			}
		} else {
			b.EraseFromParent()
		}
	}()
	gen.builder.SetInsertPointAtEnd(oldBlocks[0])
	gen.builder.SetInsertPointAtEnd(b)
	for raw, t := range state.cp.tCache {
		if t, ok := t.(*structType); ok {
			gen.sTypes[t] = raw
		}
	}
	for _, inst := range state.rt.instrs {
		err := inst.runtime(&gen)
		if err != nil {
			return err
		}
	}
	gen.builder.CreateRetVoid()
	gvals := make(map[llvm.Value]llvm.Value, len(gen.modGlobals))
	for g := range gen.modGlobals {
		v, err := g.init.load(g.ty, 0)
		if err != nil {
			return err
		}
		gvals[g.llval] = gen.value(g.llval.Type().ElementType(), v)
	}
	for g, init := range gvals {
		g.SetInitializer(init)
	}
	ok = true
	return nil
}
