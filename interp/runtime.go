package interp

import "tinygo.org/x/go-llvm"

type rtGen struct {
	ctx llvm.Context
	mod llvm.Module

	cur uint

	modGlobals map[*memObj]struct{}

	ptrs   map[ptkey]llvm.Type
	iTypes map[iType]llvm.Type
	sTypes map[*structType]llvm.Type
	aTypes map[atkey]llvm.Type

	lifetimeStart, lifetimeEnd llvm.Value

	stack []llvm.Value

	vals map[vkey]llvm.Value

	ends        []lifetimeEnd
	toMoveStart []llvm.Value

	block   llvm.BasicBlock
	builder llvm.Builder
	dbg     llvm.Metadata
}

func (g *rtGen) run(insts ...instruction) error {
	for i, inst := range insts {
		g.cur = uint(i)
		if debug {
			println("gen rt inst:", inst.String())
		}
		err := inst.runtime(g)
		if err != nil {
			return err
		}
		for len(g.ends) > 0 && g.ends[0].at == uint(i) {
			end := g.ends[0]
			if debug {
				println("end lifetime of", end.obj.String())
			}
			g.builder.CreateCall(g.getLifetimeEndFunc(), []llvm.Value{
				g.value(g.iType(i64), smallIntValue(i64, end.obj.size)),
				g.value(g.ptr(g.iType(8), defaultAddrSpace), end.obj.ptr(0)),
			}, "")
			g.ends[0] = g.ends[len(g.ends)-1]
			g.ends = g.ends[:len(g.ends)-1]
			i := 0
			for {
				min := i
				if l := 2*i + 1; l < len(g.ends) && g.ends[l].less(g.ends[min]) {
					min = l
				}
				if r := 2*i + 2; r < len(g.ends) && g.ends[r].less(g.ends[min]) {
					min = r
				}
				if min == i {
					break
				}
				g.ends[i], g.ends[min] = g.ends[min], g.ends[i]
				i = min
			}
		}
	}
	g.builder.CreateRetVoid()
	if len(g.toMoveStart) != 0 {
		for _, i := range g.toMoveStart {
			i.RemoveFromParentAsInstruction()
		}
		g.builder.SetInsertPointBefore(g.block.FirstInstruction())
		for _, i := range g.toMoveStart {
			g.builder.Insert(i)
		}
	}
	return nil
}

// getLifetimeStartFunc returns the llvm.lifetime.start intrinsic and creates it
// first if it doesn't exist yet.
func (g *rtGen) getLifetimeStartFunc() llvm.Value {
	fn := g.lifetimeStart
	if fn.IsNil() {
		fn = g.mod.NamedFunction("llvm.lifetime.start.p0i8")
		if fn.IsNil() {
			fnType := llvm.FunctionType(g.ctx.VoidType(), []llvm.Type{g.iType(i64), g.ptr(g.iType(i8), defaultAddrSpace)}, false)
			fn = llvm.AddFunction(g.mod, "llvm.lifetime.start.p0i8", fnType)
		}
		g.lifetimeStart = fn
	}
	return fn
}

// getLifetimeEndFunc returns the llvm.lifetime.end intrinsic and creates it
// first if it doesn't exist yet.
func (g *rtGen) getLifetimeEndFunc() llvm.Value {
	fn := g.lifetimeEnd
	if fn.IsNil() {
		fn = g.mod.NamedFunction("llvm.lifetime.end.p0i8")
		if fn.IsNil() {
			fnType := llvm.FunctionType(g.ctx.VoidType(), []llvm.Type{g.iType(i64), g.ptr(g.iType(i8), defaultAddrSpace)}, false)
			fn = llvm.AddFunction(g.mod, "llvm.lifetime.end.p0i8", fnType)
		}
		g.lifetimeEnd = fn
	}
	return fn
}

func (g *rtGen) typ(typ typ) llvm.Type {
	switch typ := typ.(type) {
	case iType:
		return g.iType(typ)
	case arrType:
		return g.arr(g.typ(typ.of), typ.n)
	case *structType:
		return g.sTypes[typ]
	case ptrType:
		return g.ptr(g.iType(i8), typ.in())
	default:
		panic("bad type")
	}
}

func (g *rtGen) arr(of llvm.Type, n uint32) llvm.Type {
	if t, ok := g.aTypes[atkey{of, n}]; ok {
		return t
	}

	t := llvm.ArrayType(of, int(n))
	g.aTypes[atkey{of, n}] = t
	return t
}

func (g *rtGen) ptr(to llvm.Type, in addrSpace) llvm.Type {
	if t, ok := g.ptrs[ptkey{to, in}]; ok {
		return t
	}

	t := llvm.PointerType(to, int(in))
	g.ptrs[ptkey{to, in}] = t
	return t
}

func (g *rtGen) iType(i iType) llvm.Type {
	if t, ok := g.iTypes[i]; ok {
		return t
	}

	t := g.ctx.IntType(int(i))
	g.iTypes[i] = t
	return t
}

func (g *rtGen) pushUnpacked(ty typ, v llvm.Value) {
	switch ty := ty.(type) {
	case nonAggTyp:
		g.stack = append(g.stack, v)
	case *structType:
		for i, f := range ty.fields {
			v := g.builder.CreateExtractValue(v, i, "")
			g.applyDebug(v)
			g.pushUnpacked(f.ty, v)
		}
	case arrType:
		for i := uint32(0); i < ty.n; i++ {
			v := g.builder.CreateExtractValue(v, int(i), "")
			g.applyDebug(v)
			g.pushUnpacked(ty.of, v)
		}
	default:
		panic("bad type")
	}
}

func (g *rtGen) value(t llvm.Type, v value) llvm.Value {
	if val, ok := g.vals[vkey{v, t}]; ok {
		return val
	}

	if debug {
		println("convert", v.String(), "to llvm", t.String())
	}
	val := v.val.toLLVM(t, g, v.raw)
	g.vals[vkey{v, t}] = val
	return val
}

func (g *rtGen) objPtr(obj *memObj) llvm.Value {
	if !obj.llval.IsNil() {
		return obj.llval
	}
	if obj.stack {
		ty := g.typ(obj.ty)
		alloca := g.builder.CreateAlloca(ty, obj.name)
		g.toMoveStart = append(g.toMoveStart, alloca)
		obj.llval = alloca
		i8Ptr := g.ptr(g.iType(8), defaultAddrSpace)
		cast := g.builder.CreateBitCast(alloca, i8Ptr, "")
		g.vals[vkey{obj.ptr(0), i8Ptr}] = cast
		sizeVal := g.value(g.iType(64), smallIntValue(i64, obj.size))
		g.builder.CreateCall(g.getLifetimeStartFunc(), []llvm.Value{sizeVal, cast}, "")
		if g.cur > obj.endIdx {
			g.builder.CreateCall(g.getLifetimeEndFunc(), []llvm.Value{sizeVal, cast}, "")
		} else {
			i := len(g.ends)
			g.ends = append(g.ends, lifetimeEnd{obj, obj.endIdx})
			for g.ends[i].less(g.ends[(i-1)/2]) {
				g.ends[i], g.ends[(i-1)/2] = g.ends[(i-1)/2], g.ends[i]
			}
		}
		return alloca
	}

	ty := g.typ(obj.ty)
	init, err := obj.init.load(obj.ty, 0)
	if err != nil {
		panic(err)
	}
	old := g.mod.NamedGlobal(obj.name)
	global := llvm.AddGlobal(g.mod, ty, obj.name)
	if !old.IsNil() {
		panic("collision")
	}
	global.SetLinkage(llvm.PrivateLinkage)
	global.SetInitializer(g.value(ty, init))
	if obj.alignScale != 0 {
		global.SetAlignment(1 << obj.alignScale)
	}
	obj.llval = global
	return global
}

func (g *rtGen) applyDebug(to llvm.Value) {
	// Disabled due to bugs.
	/*
		if !g.dbg.IsNil() {
			to.InstructionSetDebugLoc(g.dbg)
		}
	*/
}

type ptkey struct {
	to llvm.Type
	in addrSpace
}

type atkey struct {
	elem llvm.Type
	n    uint32
}

type vkey struct {
	v value
	t llvm.Type
}

type lifetimeEnd struct {
	obj *memObj
	at  uint
}

func (l lifetimeEnd) less(other lifetimeEnd) bool {
	switch {
	case l.at < other.at:
		return true
	case l.at > other.at:
		return false
	}
	return l.obj.id < other.obj.id
}
