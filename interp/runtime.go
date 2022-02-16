package interp

import "tinygo.org/x/go-llvm"

type rtGen struct {
	ctx llvm.Context

	modGlobals map[*memObj]struct{}

	ptrs   map[ptkey]llvm.Type
	iTypes map[iType]llvm.Type
	sTypes map[*structType]llvm.Type
	aTypes map[atkey]llvm.Type

	//lifetimeStart, lifetimeEnd llvm.Value

	stack []llvm.Value

	vals map[vkey]llvm.Value

	/*
		allocaBuilder llvm.Builder
		ends          []lifetimeEnd
	*/

	builder llvm.Builder
	dbg     llvm.Metadata
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
	panic("TODO")
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
