package interp

import (
	"errors"
	"math"
	"math/bits"

	"tinygo.org/x/go-llvm"
)

type parser interface {
	value(llvm.Value) (value, error)
	typ(llvm.Type) (typ, error)
}

type constParser struct {
	ctx           llvm.Context
	td            llvm.TargetData
	tCache        map[llvm.Type]typ
	vCache        map[llvm.Value]value
	globals       map[llvm.Value]*memObj
	globalsByName map[string]*memObj
	fCache        map[llvm.Type]fnTyInfo
	alignCache    map[typ]uint
	layouts       map[value]typ
	uintptr       iType
}

func (p *constParser) value(v llvm.Value) (value, error) {
	if val, ok := p.vCache[v]; ok {
		return val, nil
	}

	val, err := p.parseConst(v)
	if err != nil {
		return value{}, err
	}

	p.vCache[v] = val
	return val, nil
}

func (p *constParser) parseConst(v llvm.Value) (value, error) {
	switch {
	case !v.IsAConstantInt().IsNil():
		typ, err := p.typ(v.Type())
		if err != nil {
			return value{}, err
		}
		iTyp, ok := typ.(iType)
		if !ok {
			return value{}, errors.New("integer constant is not an integer")
		}
		if iTyp <= i64 {
			return smallIntValue(iTyp, v.ZExtValue()), nil
		}

	case !v.IsAConstantFP().IsNil():
		typ, err := p.typ(v.Type())
		if err != nil {
			return value{}, err
		}
		result, inexact := v.DoubleValue()
		if inexact || math.IsNaN(result) {
			return value{}, todo("NaN")
		}
		switch typ {
		case floatType{}:
			return floatValue(float32(result)), nil
		case doubleType{}:
			return doubleValue(result), nil
		}

	case !v.IsAGlobalValue().IsNil():
		if obj, ok := p.globals[v]; ok {
			return obj.ptr(0), nil
		}

		// TODO: handle global aliases.
		// Right now the whole pass breaks if they are used.

	case v.IsNull():
		typ, err := p.typ(v.Type())
		if err != nil {
			return value{}, err
		}
		return typ.zero(), nil

	case v.IsUndef():
		typ, err := p.typ(v.Type())
		if err != nil {
			return value{}, err
		}
		return undefValue(typ), nil

	case !v.IsAConstantExpr().IsNil():
		op := v.Opcode()
		switch op {
		case llvm.ExtractValue:
			return parseExtractValue(v, p)
		case llvm.InsertValue:
			return parseInsertValue(v, p)
		case llvm.Trunc, llvm.ZExt, llvm.BitCast, llvm.PtrToInt, llvm.IntToPtr:
			from, err := p.value(v.Operand(0))
			if err != nil {
				return value{}, err
			}
			to, err := p.typ(v.Type())
			if err != nil {
				return value{}, err
			}
			switch to := to.(type) {
			case nonAggTyp:
				return cast(to, from), nil
			default:
				return value{}, todo("cast type " + to.String())
			}
		}
		expr, err := parseExpr(op, v, p)
		if err != nil {
			return value{}, err
		}
		res, err := expr.eval()
		if err != nil {
			return value{}, err
		}
		return res, nil

	case !v.IsAConstantArray().IsNil():
		ty, err := p.typ(v.Type())
		if err != nil {
			return value{}, err
		}
		arrTy := ty.(arrType)
		arr := make([]value, arrTy.n)
		for i := range arr {
			v, err := p.value(v.Operand(i))
			if err != nil {
				return value{}, err
			}
			arr[i] = v
		}
		return arrayValue(arrTy.of, arr...), nil

	case !v.IsAConstant().IsNil() && v.Type().TypeKind() == llvm.ArrayTypeKind:
		// Yes. This is entirely different from !v.IsAConstantArray().IsNil().
		ty, err := p.typ(v.Type())
		if err != nil {
			return value{}, err
		}
		arrTy := ty.(arrType)
		arr := make([]value, arrTy.n)
		for i := range arr {
			v, err := p.value(llvm.ConstExtractValue(v, []uint32{uint32(i)}))
			if err != nil {
				return value{}, err
			}
			arr[i] = v
		}
		return arrayValue(arrTy.of, arr...), nil

	case !v.IsAConstantStruct().IsNil():
		ty, err := p.typ(v.Type())
		if err != nil {
			return value{}, err
		}
		sTy := ty.(*structType)
		fields := make([]value, len(sTy.fields))
		for i := range fields {
			v, err := p.value(v.Operand(i))
			if err != nil {
				return value{}, err
			}
			fields[i] = v
		}
		return structValue(sTy, fields...), nil

	case v.Type().TypeKind() == llvm.MetadataTypeKind:
		return value{metadataValue(v), 0}, nil
	}

	if debug {
		println("unable to parse constant:")
		print("\t")
		v.Dump()
		println()
	}
	return value{}, todo("parse this constant")
}

func (p *constParser) typ(t llvm.Type) (typ, error) {
	if typ, ok := p.tCache[t]; ok {
		return typ, nil
	}

	typ, err := p.parseTyp(t)
	if err != nil {
		return nil, err
	}

	p.tCache[t] = typ
	return typ, nil
}

func (p *constParser) align(t typ) uint {
	if a, ok := p.alignCache[t]; ok {
		return a
	}

	a := p.calcAlign(t)
	p.alignCache[t] = a

	return a
}

func (p *constParser) calcAlign(t typ) uint {
	switch t := t.(type) {
	case iType:
		return uint(bits.TrailingZeros(uint(p.td.ABITypeAlignment(p.ctx.IntType(int(t))))))
	case arrType:
		return p.align(t.of)
	case ptrType:
		return p.align(t.idxTy())
	case *structType:
		var max uint
		for _, f := range t.fields {
			a := p.align(f.ty)
			if a > max {
				max = a
			}
		}
		return max
	default:
		panic("unknown type " + t.String())
	}
}

func parseExtractValue(expr llvm.Value, p parser) (value, error) {
	agg, err := p.value(expr.Operand(0))
	if err != nil {
		return value{}, err
	}
	return extractValue(agg, expr.Indices()...), nil
}

func parseInsertValue(expr llvm.Value, p parser) (value, error) {
	agg, err := p.value(expr.Operand(0))
	if err != nil {
		return value{}, err
	}
	v, err := p.value(expr.Operand(1))
	if err != nil {
		return value{}, err
	}
	return insertValue(agg, v, expr.Indices()...), nil
}

func (p *constParser) parseTyp(t llvm.Type) (typ, error) {
	switch t.TypeKind() {
	case llvm.IntegerTypeKind:
		width := t.IntTypeWidth()
		if width <= 64 {
			return iType(width), nil
		}

	case llvm.FloatTypeKind:
		return floatType{}, nil

	case llvm.DoubleTypeKind:
		return doubleType{}, nil

	case llvm.PointerTypeKind:
		space := t.PointerAddressSpace()
		idxBits := p.td.TypeSizeInBits(t)
		if space < int(maxPtrAddrSpace) && idxBits <= maxPtrIdxWidth {
			return pointer(addrSpace(space), iType(idxBits)), nil
		}

	case llvm.ArrayTypeKind:
		elemTyp, err := p.typ(t.ElementType())
		if err != nil {
			return nil, err
		}
		n := t.ArrayLength()
		if uint64(n) < uint64(1<<32) {
			return array(elemTyp, uint32(n)), nil
		}

	case llvm.StructTypeKind:
		elemTypes := t.StructElementTypes()
		fields := make([]structField, len(elemTypes))
		for i, e := range elemTypes {
			typ, err := p.typ(e)
			if err != nil {
				return nil, err
			}
			off := p.td.ElementOffset(t, i)
			fields[i] = structField{typ, off}
		}
		return &structType{t.StructName(), fields, p.td.TypeAllocSize(t)}, nil

	case llvm.VoidTypeKind:
		return nil, nil

	case llvm.MetadataTypeKind:
		return metadataType{}, nil
	}

	return nil, todo("parse type " + t.String())
}
