package interp

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"tinygo.org/x/go-llvm"
)

// failInst is an instruction which fails to execute.
// This is used to implement unrecognized instructions, traps,
// panics, or static undefined behavior (e.g. unreachable).
type failInst struct {
	err error
	dbg llvm.Metadata
}

var _ instruction = (*failInst)(nil)

func (i *failInst) result() typ {
	return nil
}

func (i *failInst) exec(state *execState) error {
	return i.err
}

func (i *failInst) runtime(gen *rtGen) error {
	return errors.New("cannot run control flow at runtime")
}

func (i *failInst) String() string {
	return "fail " + strconv.Quote(i.err.Error()) + dbgSuffix(i.dbg)
}

// brInst is an unconditional branch instruction.
type brInst struct {
	edge brEdge
	dbg  llvm.Metadata
}

var _ instruction = (*brInst)(nil)

func (i *brInst) result() typ {
	return nil
}

func (i *brInst) exec(state *execState) error {
	state.br(i.edge)
	return nil
}

func (i *brInst) runtime(gen *rtGen) error {
	return errors.New("cannot run control flow at runtime")
}

func (i *brInst) String() string {
	return "br " + i.edge.String() + dbgSuffix(i.dbg)
}

// switchInst branches to a location depending on a small integer value.
// This is used to implement `switch` and conditional `br`.
type switchInst struct {
	v     value
	cases map[uint64]brEdge
	dbg   llvm.Metadata
}

var _ instruction = (*switchInst)(nil)

func (i *switchInst) result() typ {
	return nil
}

func (i *switchInst) exec(state *execState) error {
	// Resolve the switch parameter.
	v := i.v.resolve(state.locals())
	_, ok := v.val.(smallInt)
	if !ok {
		// The destination is not known, so revert.
		return errUnknownBranch
	}

	// Look up the switch case.
	if c, ok := i.cases[v.raw]; ok {
		// Branch to the case.
		state.br(c)
	}

	// Fall through to the next instruction (the default case).
	return nil
}

func (i *switchInst) runtime(gen *rtGen) error {
	return errors.New("cannot run control flow at runtime")
}

func (i *switchInst) String() string {
	t := i.v.typ().(iType)
	idxs := make([]uint64, len(i.cases))
	j := 0
	for k := range i.cases {
		idxs[j] = k
		j++
	}
	sort.Slice(idxs, func(i, j int) bool { return idxs[i] < idxs[j] })
	cases := make([]string, len(idxs))
	for j, k := range idxs {
		cases[j] = smallIntValue(t, k).String() + ": " + i.cases[k].String()
	}
	return "switch " + i.v.String() + " {" + strings.Join(cases, ", ") + "}" + dbgSuffix(i.dbg)
}

func (s *execState) br(edge brEdge) {
	stack := s.stack
	phiStart := len(stack)
	locals := s.locals()
	for _, v := range edge.phis {
		stack = append(stack, v.resolve(locals))
	}
	labels := s.curFn.labels
	var pc uint
	var height uint
	if edge.to < uint(len(labels)) {
		l := labels[edge.to]
		pc, height = l.start, l.height
	} else {
		pc = ^uint(0)
		height = uint(len(edge.phis))
	}
	height += s.sp
	copy(stack[height-uint(len(edge.phis)):height], stack[phiStart:phiStart+len(edge.phis)])
	stack = stack[:height]
	s.pc, s.stack = pc, stack
}

type brEdge struct {
	// to is the index of the destination label.
	to uint

	// phis are the PHI values for the destination.
	phis []value
}

func (e brEdge) String() string {
	var to string
	if e.to == ^uint(0) {
		to = "ret"
	} else {
		to = strconv.FormatUint(uint64(e.to), 10)
	}
	phis := make([]string, len(e.phis))
	for i, v := range e.phis {
		phis[i] = v.String()
	}
	return "[to " + to + ", φ(" + strings.Join(phis, ", ") + ")]"
}

type callInst struct {
	called          value
	args            []value
	sig             signature
	dbg             llvm.Metadata
	recursiveRevert bool
}

func (i *callInst) result() typ {
	return i.sig.ret
}

func (i *callInst) exec(state *execState) error {
	locals := state.locals()
	called := i.called.resolve(locals)
	sp := uint(len(state.stack))
	for _, a := range i.args {
		state.stack = append(state.stack, a.resolve(locals))
	}

	// Save some state.
	parentSp := state.sp
	oldMemStart := len(state.oldMem)
	parentVersion := state.version
	lr := state.pc
	revLabel := state.rt.createLabel(state.rt.stackHeight)
	state.rt.startLabel(revLabel)
	escStart := len(state.escapeStack)
	oldFn := state.curFn
	nextObjID := state.nextObjID
	stackMemStart := len(state.allocaStack)

	// Run the function.
	state.sp = sp
	state.version = state.nextVersion
	state.nextVersion++
	state.pc = 0
	err := i.run(state, called)
	switch {
	case err == nil:
		state.sp = parentSp
		state.pc = lr
		state.curFn = oldFn
		state.version = parentVersion
		if uint(len(state.rt.instrs)) == state.rt.labels[revLabel].start {
			state.rt.revert(revLabel)
		}
		endInst := len(state.rt.instrs)
		for _, obj := range state.allocaStack[stackMemStart:] {
			obj.endIdx = uint(endInst)
			obj.data = obj.init
			obj.escaped = false
		}
		state.allocaStack = state.allocaStack[:stackMemStart]
		return nil
	case !i.recursiveRevert && isRuntimeOrRevert(err):
		if debug {
			println("revert", i.String(), err.Error())
		}
		args := append([]value(nil), state.stack[sp:]...)
		state.sp = parentSp
		state.stack = state.stack[:sp]
		{
			restore := state.oldMem[oldMemStart:]
			for i := len(restore) - 1; i >= 0; i-- {
				r := restore[i]
				if debug {
					println("restore", r.obj.String(), "to version", r.version)
				}
				r.obj.data = r.tree
				r.obj.version = r.version
			}
		}
		state.oldMem = state.oldMem[:oldMemStart]
		state.nextVersion = state.version
		state.version = parentVersion
		state.pc = lr
		state.curFn = oldFn
		state.rt.revert(revLabel)
		for _, obj := range state.escapeStack[escStart:] {
			obj.escaped = false
		}
		state.escapeStack = state.escapeStack[:escStart]
		state.nextObjID = nextObjID
		state.allocaStack = state.allocaStack[:stackMemStart]
		return i.execRuntime(state, called, args)
	default:
		return err
	}
}

func (i *callInst) run(state *execState, called value) error {
	p, ok := called.val.(*offPtr)
	obj := p.obj()
	if !ok || called.raw != 0 || !obj.isFunc || obj.isExtern {
		// This is an unknown indirect call.
		return errExternalCall
	}
	err := obj.compile(&state.cp)
	if err != nil {
		return err
	}
	if len(state.stack) > maxStackSize {
		return errMaxStackSize
	}

	// TODO: verify that the signature matches the call signature
	r := obj.r
	state.curFn = r
	for state.pc < uint(len(obj.r.instrs)) {
		inst := r.instrs[state.pc]
		var resPack value
		if debug {
			if res := inst.result(); res != nil {
				resPack, _ = createDecomposedIndices(uint(len(state.locals())), res)
				println("exec", resPack.String(), "=", inst.String())
			} else {
				println("exec", inst.String())
			}
		}
		state.pc++
		err = inst.exec(state)
		if err != nil {
			if !isRuntimeOrRevert(err) {
				err = fmt.Errorf("%w\n\tat %s", err, inst.String())
			}
			return err
		}
		if debug {
			if resPack != (value{}) {
				println("\tresult:", resPack.resolve(state.locals()).String(), "=", inst.String())
			}
		}
	}

	return nil
}

func (i *callInst) execRuntime(state *execState, called value, args []value) error {
	// TODO: merge signatures
	tmp := state.escBuf
	if assert && len(tmp) != 0 {
		panic("escape buffer is not empty")
	}

	// Evaluate call signature.
	sig := i.sig
	if i.sig.ty.IsNil() {
		panic("nil sig type")
	}
	if p, ok := called.val.(*offPtr); ok && p.isFunc {
		obj := p.obj()
		var err error
		sig, err = obj.parseSig(&state.cp)
		var merge signature
		if err == nil {
			merge, err = sig.merge(obj.sig)
		}
		switch {
		case err == nil:
			sig = merge
		case isRuntimeOrRevert(err):
		default:
			return err
		}
	}

	// Calculate side effects.
	type sideEffect uint8
	const (
		observe sideEffect = 1 << iota
		modify
		escape
	)
	var baseEffects sideEffect = observe | modify | escape
	if sig.readOnly {
		baseEffects &^= modify
	}
	if sig.writeOnly {
		baseEffects &^= observe
	}
	if sig.readNone {
		baseEffects &^= observe | modify
	}
	var escEffects sideEffect
	if !sig.argMemOnly {
		escEffects = baseEffects &^ escape
	}
	var off uint
	var sideEffects map[*memObj]sideEffect
	for _, a := range sig.args {
		var v value
		v, off = createDecomposedIndices(off, a.t)
		v = v.resolve(args)
		argEffects := baseEffects
		if a.noCapture {
			argEffects &^= escape
		}
		if a.readOnly {
			argEffects &^= modify
		}
		if a.writeOnly {
			argEffects &^= observe
		}
		if a.readNone {
			argEffects &^= observe | modify
		}
		if argEffects == 0 {
			continue
		}
		complete := v.aliases(tmp)
		if !complete {
			newEscEffects := (argEffects &^ escape) &^ escEffects
			if newEscEffects != 0 {
				for obj, objEffects := range sideEffects {
					if objEffects&newEscEffects == 0 || !obj.escaped {
						continue
					}

					objEffects &^= newEscEffects
					if objEffects == 0 {
						delete(sideEffects, obj)
					} else {
						sideEffects[obj] = objEffects
					}
				}
			}
			escEffects |= newEscEffects
		}
		if len(tmp) == 0 {
			continue
		}

		if sideEffects == nil {
			sideEffects = make(map[*memObj]sideEffect)
		}
		for obj := range tmp {
			objEffects := argEffects
			if obj.escaped {
				objEffects &^= escEffects
				if objEffects == 0 {
					continue
				}
			}
			sideEffects[obj] |= objEffects
		}
		for obj := range tmp {
			delete(tmp, obj)
		}
	}

	// Escape the called function.
	state.escape(called)

	// Apply side-effects.
	dbg := i.dbg
	var worklist, toEscape []*memObj
	for obj := range sideEffects {
		worklist = append(worklist, obj)
	}
	sortObjects(worklist)
	for _, obj := range worklist {
		objEffects := sideEffects[obj]
		var err error
		if objEffects&modify != 0 {
			err = state.invalidate(obj, dbg)
		} else {
			err = state.flush(obj, dbg)
		}
		if err != nil {
			return err
		}
		if objEffects&escape != 0 {
			toEscape = append(toEscape, obj)
		}
	}
	var err error
	if escEffects&modify == 0 && sig.noSync {
		err = state.flushEscaped(dbg)
	} else {
		err = state.invalidateEscaped(dbg)
	}
	if err != nil {
		return err
	}
	for _, obj := range toEscape {
		if obj.escaped {
			continue
		}

		obj.escaped = true
		if debug {
			println("capture", obj.String())
		}
		state.escapeStack = append(state.escapeStack, obj)
	}

	// Create a call to the function.
	v := state.rt.insertInst(&callInst{called, args, sig, dbg, false})
	if sig.ret != nil {
		state.stack = unpack(v, state.stack)
	}

	return nil
}

func (i *callInst) runtime(gen *rtGen) error {
	oldDbg := gen.dbg
	gen.dbg = i.dbg
	defer func() { gen.dbg = oldDbg }()

	if i.sig.ty.IsNil() {
		panic("nil sig type")
	}
	paramTypes := i.sig.ty.ElementType().ParamTypes()
	called := gen.value(i.sig.ty, i.called)
	args := make([]llvm.Value, len(i.sig.args))
	var off uint
	for j, a := range i.sig.args {
		var v value
		v, off = createDecomposedIndices(off, a.t)
		v = v.resolve(i.args)
		args[j] = gen.value(paramTypes[j], v)
	}
	call := gen.builder.CreateCall(called, args, "")
	gen.applyDebug(call)
	// TODO: copy call site attributes
	if i.sig.ret != nil {
		gen.pushUnpacked(i.sig.ret, call)
	}
	return nil
}

func (i *callInst) String() string {
	args := make([]string, len(i.args))
	for i, a := range i.args {
		args[i] = a.String()
	}
	return "call " + i.sig.String() + ", " + strings.Join(append([]string{i.called.String()}, args...), ", ") + dbgSuffix(i.dbg)
}

var errExternalCall = errRevert{errors.New("external call")}
