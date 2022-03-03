package interp

import (
	"errors"
	"fmt"
	"math/bits"
	"strconv"
	"strings"

	"tinygo.org/x/go-llvm"
)

func (c *constParser) mapGlobals(mod llvm.Module, forceNoEscape map[llvm.Value]struct{}) (uint64, error) {
	objs := make(map[llvm.Value]*memObj)
	var nextID uint64

	// List all globals.
	for g := mod.FirstGlobal(); !g.IsNil(); g = llvm.NextGlobal(g) {
		ptrTy := g.Type()
		ty, err := c.typ(ptrTy)
		if err != nil {
			return 0, err
		}

		isExtern := g.IsDeclaration()
		var size uint64
		var ety typ
		if !isExtern {
			ety, err = c.typ(ptrTy.ElementType())
			if err == nil {
				size = ety.bytes()
			}
		}
		linkage := g.Linkage()
		align := g.Alignment()
		if align == 0 {
			// Assume ABI type alignment?
			align = c.td.ABITypeAlignment(g.Type().ElementType())
		}
		if assert {
			if align <= 0 {
				println(g.Name(), align)
				panic("negative alignment")
			}
			if bits.OnesCount(uint(align)) != 1 {
				panic("align is not a power of 2")
			}
		}
		_, forceNoEscape := forceNoEscape[g]
		objs[g] = &memObj{
			ptrTy:      ty.(ptrType),
			name:       g.Name(),
			id:         nextID,
			isExtern:   isExtern,
			isConst:    g.IsGlobalConstant(),
			unique:     false, // TODO: check for unnamed_addr.
			escaped:    !(forceNoEscape || linkage == llvm.InternalLinkage || linkage == llvm.PrivateLinkage),
			size:       size,
			alignScale: uint(bits.TrailingZeros(uint(align))),
			ty:         ety,
			llval:      g,
			err:        err,
		}
		nextID++
	}

	// List all functions.
	for f := mod.FirstFunction(); !f.IsNil(); f = llvm.NextFunction(f) {
		ptrTy := f.Type()
		ty, err := c.typ(ptrTy)
		if err != nil {
			return 0, err
		}
		linkage := f.Linkage()
		_, forceNoEscape := forceNoEscape[f]
		objs[f] = &memObj{
			ptrTy:    ty.(ptrType),
			name:     f.Name(),
			isFunc:   true,
			isExtern: f.IsDeclaration(),
			unique:   false, // TODO: check for unnamed_addr.
			escaped:  !(forceNoEscape || linkage == llvm.InternalLinkage || linkage == llvm.PrivateLinkage),
			llval:    f,
		}
	}

	// Do a global escape analysis.
	if debug {
		println("global escape analysis:")
	}
	var worklist []llvm.Value
	visited := make(map[llvm.Value]struct{})
	for g := range objs {
		worklist = append(worklist, g)
		visited[g] = struct{}{}
	}
	used := make(map[llvm.Value][]llvm.Value)
	dedup := make(map[llvm.Value]struct{})
	for len(worklist) > 0 {
		v := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		for use := v.FirstUse(); !use.IsNil(); use = use.NextUse() {
			user := use.User()
			if _, ok := forceNoEscape[user]; ok {
				continue
			}
			if !user.IsAInstruction().IsNil() {
				user = user.InstructionParent().Parent()
			}
			if _, ok := dedup[user]; ok {
				continue
			}
			dedup[user] = struct{}{}
			used[user] = append(used[user], v)
			if _, ok := visited[user]; !ok {
				worklist = append(worklist, user)
				visited[user] = struct{}{}
			}
		}
		for user := range dedup {
			delete(dedup, user)
		}
	}
	for v := range visited {
		delete(visited, v)
	}
	for g, obj := range objs {
		if obj.escaped {
			if debug {
				println("\tescaped by linkage:", obj.String())
			}
			worklist = append(worklist, g)
			visited[g] = struct{}{}
		}
	}
	for len(worklist) > 0 {
		v := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		if obj, ok := objs[v]; ok {
			if !obj.escaped {
				if debug {
					println("\tescape by reference from escaped global:", obj.String())
				}
				obj.escaped = true
			}
		}
		for _, u := range used[v] {
			if _, ok := visited[u]; !ok {
				worklist = append(worklist, u)
				visited[u] = struct{}{}
			}
		}
	}
	for g, obj := range objs {
		if obj.escaped {
			continue
		}
		var esc []*memObj
		if assert && len(worklist) != 0 {
			panic("worklist not empty")
		}
		worklist = append(worklist, g)
		for i := 0; i < len(worklist); i++ {
			v := worklist[i]
			for _, u := range used[v] {
				if _, ok := visited[u]; ok {
					continue
				}

				visited[u] = struct{}{}
				if obj, ok := objs[u]; ok {
					esc = append(esc, obj)
				} else {
					worklist = append(worklist, u)
				}
			}
		}
		sortObjects(esc)
		if debug {
			for _, e := range esc {
				println("\tescape", e.String(), "if", obj.String(), "escapes")
			}
		}
		for _, v := range worklist {
			delete(visited, v)
		}
		for _, obj := range esc {
			delete(visited, obj.llval)
		}
		worklist = worklist[:0]
		obj.esc = esc
	}

	c.globals = objs
	return nextID, nil
}

type memObj struct {
	// ptrTy is the pointer type of this object.
	ptrTy ptrType

	// name is the internal name of the memory object.
	name string

	// id is an internal ID for the object.
	// It is used to provide consistent ordering for flushing and invalidation.
	id uint64

	// isFunc indicates whether this is a function
	isFunc bool

	// isExtern indicates whether this is defined externally.
	isExtern bool

	// isConst indicates whether this is a constant.
	isConst bool

	// unique is set if the object is not equal to any other object.
	unique bool

	// escaped is set if the object is escaped.
	escaped bool

	// stack is set if this is a stack object.
	stack bool

	// llval is the corresponding IR value (if applicable).
	llval llvm.Value

	// endIdx is the runtime instruction after which this object is dead.
	endIdx uint

	// size is the size of the backing storage for the object.
	// If this is not known, it is set to 0.
	size uint64

	// alignScale is the minimum alignment of the object (as an exponent of 2).
	alignScale uint

	// ty is the type of the object (if known)
	ty typ

	// dbg is the debug metadata associated with the object's creation.
	dbg llvm.Metadata

	// esc is a set of escape graph edges.
	// If this object is escaped, these objects must also be escaped.
	esc []*memObj

	// version is the version of the current contents of the node.
	// If the object has not yet been initialized/compiled, this is memVersionZero.
	version uint64

	// init was the initializer before interp ran (if this is an internal global).
	// This should remain nil if the initializer was not fully parsed.
	init memTreeNode

	// data is the current contents of the node.
	data memTreeNode

	// sig is the signature for this function.
	sig signature

	// r is the runnable for this function once it has been compiled.
	r *runnable

	// err is the error produced when compiling (if this is a function).
	err error
}

func (obj *memObj) parseSig(c *constParser) (signature, error) {
	if !obj.sig.ty.IsNil() {
		return obj.sig, nil
	}

	sig, err := c.parseFuncSignature(obj.llval)
	if err != nil {
		obj.err = err
		return signature{}, err
	}
	if debug {
		println("parsed sig for", obj.String(), "as", sig.String())
	}
	obj.sig = sig
	return sig, nil
}

func (obj *memObj) compile(c *constParser) error {
	if obj.err != nil {
		return obj.err
	}
	if obj.version != 0 {
		return nil
	}
	sig, err := obj.parseSig(c)
	if err != nil {
		return err
	}
	if sig.noInline {
		if debug {
			println("compile blocked by noinline")
		}
		obj.sig = sig
		obj.err = errRuntime
		return errRuntime
	}
	if debug {
		println("compile", sig.String(), obj.String())
	}
	r, err := compile(c, sig, obj.llval)
	if err != nil {
		obj.err = err
		return err
	}
	obj.r = r
	obj.version = 1
	if debug {
		println("\t" + strings.ReplaceAll(r.String(), "\n", "\n\t"))
	}
	return nil
}

func (obj *memObj) parseInit(c *constParser) error {
	if obj.err != nil {
		return obj.err
	}
	if obj.version != 0 {
		return nil
	}
	if obj.isExtern {
		return errRuntime
	}
	if debug {
		println("parse init", obj.ty.String(), obj.String())
	}
	init, err := tryParseMem(c, obj.llval.Initializer())
	switch err {
	case nil:
	case errRuntime:
		init = makeUnknownMem(obj.version, obj.size)
	default:
		obj.err = err
		return err
	}
	obj.init = init
	obj.data = init
	obj.version = 1
	if debug {
		println("\t" + strings.ReplaceAll(init.String(), "\n", "\n\t"))
	}
	return nil
}

func tryParseMem(c *constParser, init llvm.Value) (memTreeNode, error) {
	v, err := c.parseConst(init)
	if err != nil {
		return nil, err
	}
	return makeInitializedMem(v)
}

func (obj *memObj) ptr(off uint64) value {
	return value{(*offPtr)(obj), off & ((1 << obj.ptrTy.bits()) - 1)}
}

func (obj *memObj) addr(off uint64) value {
	return value{(*offAddr)(obj), off & ((1 << obj.ptrTy.bits()) - 1)}
}

func (obj *memObj) gep(off value) value {
	if assert {
		if offTy, pIdxTy := off.typ(), obj.ptrTy.idxTy(); offTy != pIdxTy {
			panic(typeError{pIdxTy, offTy})
		}
	}
	switch off.val.(type) {
	case smallInt:
		return obj.ptr(off.raw)
	case undef:
		return undefValue(obj.ptrTy)
	default:
		return value{uglyGEP{obj, off.val}, off.raw}
	}
}

func (obj *memObj) String() string {
	return "@" + maybeQuoteName(obj.name)
}

type allocaInst struct {
	// ty is the type of the allocation element.
	ty typ

	// n is the number of elements.
	n value

	// ptrTy is the pointer type of the allocation.
	ptrTy ptrType

	alignScale uint

	// dbg is the source instruction's debug metadata.
	dbg llvm.Metadata
}

func (i *allocaInst) result() typ {
	return i.ptrTy
}

func (i *allocaInst) exec(state *execState) error {
	n := i.n.resolve(state.locals())
	_, ok := n.val.(smallInt)
	if !ok {
		return todo("alloca of unknown size")
	}
	elemSize := i.ty.bytes()
	overflow, size := bits.Mul64(n.raw, elemSize)
	switch {
	case overflow != 0 || size&^((1<<i.ptrTy.idxTy().bits())-1) != 0:
		return errUB

	case size > 1<<32 || n.raw >= 1<<32:
		return errRevert{errors.New("alloca is too big")}
	}
	var ty typ
	if n.raw == 1 {
		ty = i.ty
	} else {
		ty = array(i.ty, uint32(n.raw))
	}
	mem := makeUndefMem(size)
	id := state.nextObjID
	state.nextObjID++
	obj := &memObj{
		ptrTy:      i.ptrTy,
		name:       "interp.alloca." + strconv.FormatUint(id, 10),
		id:         id,
		unique:     size != 0,
		size:       size,
		alignScale: i.alignScale,
		ty:         ty,
		dbg:        i.dbg,
		version:    state.version,
		init:       mem,
		data:       mem,
		stack:      true,
	}
	state.stack = append(state.stack, obj.ptr(0))
	state.allocaStack = append(state.allocaStack, obj)
	return nil
}

func (i *allocaInst) runtime(gen *rtGen) error {
	panic("cannot alloca at runtime")
}

func (i *allocaInst) String() string {
	return i.ptrTy.String() + " alloca " + i.n.String() + " of " + i.ty.String() + " align " + strconv.FormatUint(1<<i.alignScale, 10)
}

type loadInst struct {
	// from is source pointer.
	from value

	// ty is the loaded type.
	ty typ

	alignScale uint

	volatile bool

	order llvm.AtomicOrdering

	// rawTy is the raw LLVM pointer type of the load.
	rawTy llvm.Type

	// dbg is the source instruction's debug metadata.
	dbg llvm.Metadata
}

func (i *loadInst) result() typ {
	return i.ty
}

func (i *loadInst) exec(state *execState) error {
	// TODO: combine loads (insert load index into tree when doing a runtime load)
	// TODO: range invalidation (instead of invalidating the whole object)
	from := i.from.resolve(state.locals())
	size := i.ty.bytes()
	v, err := i.tryLoad(state, from, size)
	switch err {
	case nil:
	case errRuntime:
		m := state.escBuf
		complete := from.aliases(m)
		var aliases []*memObj
		for obj := range m {
			aliases = append(aliases, obj)
		}
		for obj := range m {
			delete(m, obj)
		}
		sortObjects(aliases)
		for _, obj := range aliases {
			err := state.flush(obj, i.dbg)
			if err != nil {
				return err
			}
		}
		needSync := i.order != llvm.AtomicOrderingNotAtomic
		switch {
		case needSync:
			err := state.invalidateEscaped(i.dbg)
			if err != nil {
				return err
			}

		case !complete:
			err := state.flushEscaped(i.dbg)
			if err != nil {
				return err
			}
		}
		v = state.rt.insertInst(&loadInst{from, i.ty, i.alignScale, i.volatile, i.order, i.rawTy, i.dbg})
	default:
		return err
	}
	state.stack = unpack(v, state.stack)
	return nil
}

func (i *loadInst) tryLoad(state *execState, from value, size uint64) (value, error) {
	if i.volatile {
		return value{}, errRuntime
	}
	p, ok := from.val.(*offPtr)
	if !ok {
		return value{}, errRuntime
	}
	obj := p.obj()
	if obj.isExtern || obj.isFunc || from.raw+size > obj.size || (i.order != llvm.AtomicOrderingNotAtomic && obj.escaped) {
		return value{}, errRuntime
	}
	if ptrAlignScale := uint(bits.TrailingZeros64(from.raw | (1 << obj.alignScale))); ptrAlignScale < i.alignScale {
		return value{}, errAlign{"load", i.alignScale, ptrAlignScale}
	}
	err := obj.parseInit(&state.cp)
	if err != nil {
		return value{}, err
	}
	return obj.data.load(i.ty, from.raw)
}

func (i *loadInst) runtime(gen *rtGen) error {
	oldDbg := gen.dbg
	gen.dbg = i.dbg
	defer func() { gen.dbg = oldDbg }()

	from := gen.value(i.rawTy, i.from)
	v := gen.builder.CreateLoad(from, "")
	if i.volatile {
		v.SetVolatile(true)
	}
	if i.order != llvm.AtomicOrderingNotAtomic {
		v.SetOrdering(i.order)
	}
	v.SetAlignment(1 << i.alignScale)
	gen.applyDebug(v)
	gen.pushUnpacked(i.ty, v)
	return nil
}

func (i *loadInst) String() string {
	var volatile string
	if i.volatile {
		volatile = " volatile"
	}
	return i.ty.String() + volatile + " " + orderString(i.order) + " load " + i.from.String() + " align " + strconv.FormatUint(1<<i.alignScale, 10) + dbgSuffix(i.dbg)
}

type storeInst struct {
	// to is the destination pointer.
	to value

	// v is the stored value.
	v value

	alignScale uint

	volatile bool

	order llvm.AtomicOrdering

	// rawTy is the raw LLVM type of the store.
	// TODO: this should not be necessary/possible except for volatile
	rawTy llvm.Type

	// dbg is the source instruction's debug metadata.
	dbg llvm.Metadata

	// init is used to flag initializing runtime stores.
	init bool
}

func (i *storeInst) result() typ {
	return nil
}

func (i *storeInst) exec(state *execState) error {
	locals := state.locals()
	to, v := i.to.resolve(locals), i.v.resolve(locals)
	size := v.typ().bytes()
	switch err := i.tryStore(state, to, v, size); err {
	case nil:
		return nil
	case errRuntime:
	default:
		return err
	}
	m := state.escBuf
	complete := to.aliases(m)
	var aliases []*memObj
	for obj := range m {
		aliases = append(aliases, obj)
	}
	for obj := range m {
		delete(m, obj)
	}
	sortObjects(aliases)
	for _, obj := range aliases {
		err := state.invalidate(obj, i.dbg)
		if err != nil {
			return err
		}
	}
	if !complete || i.order != llvm.AtomicOrderingNotAtomic {
		err := state.invalidateEscaped(i.dbg)
		if err != nil {
			return err
		}
	}
	state.rt.insertInst(&storeInst{to, v, i.alignScale, i.volatile, i.order, i.rawTy, i.dbg, false})
	return nil
}

func (i *storeInst) tryStore(state *execState, to, v value, size uint64) error {
	if i.volatile {
		return errRuntime
	}
	p, ok := to.val.(*offPtr)
	if !ok {
		return errRuntime
	}
	obj := p.obj()
	if obj.isExtern || obj.isFunc || to.raw+size > obj.size || (i.order != llvm.AtomicOrderingNotAtomic && obj.escaped) {
		return errRuntime
	}
	if obj.isConst {
		return errUB
	}
	if ptrAlignScale := uint(bits.TrailingZeros64(to.raw | (1 << obj.alignScale))); ptrAlignScale < i.alignScale {
		return errAlign{"store", i.alignScale, ptrAlignScale}
	}
	err := obj.parseInit(&state.cp)
	if err != nil {
		return err
	}
	node, err := obj.data.store(v, to.raw, state.version)
	if err != nil {
		return err
	}
	if obj.version < state.version {
		state.oldMem = append(state.oldMem, memSave{
			obj:     obj,
			tree:    obj.data,
			version: obj.version,
		})
	}
	obj.data = node
	obj.version = state.version
	return nil
}

func (i *storeInst) runtime(gen *rtGen) error {
	if i.init {
		to, v := i.to, i.v
		p := to.val.(*offPtr)
		size := v.typ().bytes()
		obj := p.obj()
		ok, err := obj.init.canInit(to.raw, size)
		if err != nil {
			return err
		}
		if ok && i.v.constant() {
			if debug {
				println("init store ok", v.String(), "to", to.String())
			}
			node, err := obj.init.store(v, to.raw, memVersionStart)
			if err != nil {
				return err
			}
			obj.init = node
			gen.modGlobals[obj] = struct{}{}
			return nil
		}
		if debug {
			println("cannot init store", v.String(), "to", to.String())
		}
		node, err := obj.init.blockInit(to.raw, size, memVersionStart)
		if err != nil {
			return err
		}
		obj.init = node
	}

	oldDbg := gen.dbg
	gen.dbg = i.dbg
	defer func() { gen.dbg = oldDbg }()

	ty := i.rawTy
	if i.rawTy.IsNil() {
		ty = gen.typ(i.v.typ())
	}
	to := gen.value(gen.ptr(ty, i.to.typ().(ptrType).in()), i.to)
	v := gen.value(ty, i.v)
	store := gen.builder.CreateStore(v, to)
	if i.volatile {
		store.SetVolatile(true)
	}
	if i.order != llvm.AtomicOrderingNotAtomic {
		store.SetOrdering(i.order)
	}
	store.SetAlignment(1 << i.alignScale)
	gen.applyDebug(store)
	return nil
}

func (i *storeInst) String() string {
	var volatile string
	if i.volatile {
		volatile = "volatile "
	}
	var init string
	if i.init {
		init = "init "
	}
	return "store " + init + volatile + orderString(i.order) + " " + i.v.String() + " to " + i.to.String() + " align " + strconv.FormatUint(1<<i.alignScale, 10) + dbgSuffix(i.dbg)
}

func orderString(order llvm.AtomicOrdering) string {
	switch order {
	case llvm.AtomicOrderingNotAtomic:
		return "nosync"
	case llvm.AtomicOrderingUnordered:
		return "unordered"
	case llvm.AtomicOrderingMonotonic:
		return "monotonic"
	case llvm.AtomicOrderingAcquire:
		return "acquire"
	case llvm.AtomicOrderingRelease:
		return "release"
	case llvm.AtomicOrderingAcquireRelease:
		return "acq_rel"
	case llvm.AtomicOrderingSequentiallyConsistent:
		return "seq_cst"
	default:
		return "unknown_order"
	}
}

type errAlign struct {
	accessTy         string
	accessAlignScale uint
	ptrAlignScale    uint
}

func (err errAlign) Error() string {
	return fmt.Sprintf("insufficient alignment for %s: operation requires %d-byte alignment but got %d-byte alignment", err.accessTy, 1<<err.accessAlignScale, 1<<err.ptrAlignScale)
}
