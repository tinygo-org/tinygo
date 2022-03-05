package interp

import (
	"fmt"
	"strconv"
	"strings"

	"tinygo.org/x/go-llvm"
)

// value is a concrete or symbolic IR value.
// It is always comparable, and equal values can be considered equivalent.
// Equivalence does not imply equality.
type value struct {
	// val holds additional data and controls the interpretation of the value.
	val val

	// raw is a raw component of the value.
	// The usage of this depends on the implementation of val.
	raw uint64
}

func (v value) constant() bool {
	return v.val.constant(v.raw)
}

func (v value) typ() typ {
	return v.val.typ(v.raw)
}

func (v value) String() string {
	return v.typ().String() + " " + v.val.str(v.raw)
}

func (v value) resolve(stack []value) value {
	return v.val.resolve(stack, v.raw)
}

func (v value) aliases(objs map[*memObj]struct{}) bool {
	return v.val.aliases(objs, v.raw)
}

// val is a component of the value which indicates how to process it.
type val interface {
	// constant returns whether this is a constant (by LLVM's definition).
	constant(raw uint64) bool

	// typ returns the type of the value, given the raw component.
	typ(raw uint64) typ

	// str returns a textual representation of the value, given the raw component.
	str(raw uint64) string

	// resolve the value within a context.
	// This replaces references with corresponding values from the stack.
	resolve(stack []value, raw uint64) value

	// toLLVM converts a value to an LLVM value.
	toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value

	// aliases adds all aliased objects to the set.
	// It returns true if this is a complete list.
	// If it returns false, the value may also alias escaped objects.
	aliases(objs map[*memObj]struct{}, raw uint64) bool
}

func boolValue(v bool) value {
	if v {
		return smallIntValue(i1, 1)
	} else {
		return smallIntValue(i1, 0)
	}
}

// smallIntValue returns a value of a provided small integer type.
func smallIntValue(ty iType, raw uint64) value {
	if ty > 64 {
		panic("too big")
	}

	return value{smallInt(ty), raw & ((1 << ty) - 1)}
}

// smallInt is a small integer value.
type smallInt uint8

func (v smallInt) constant(raw uint64) bool {
	return true
}

func (v smallInt) typ(raw uint64) typ {
	return iType(v)
}

func (v smallInt) str(raw uint64) string {
	if v == 1 {
		switch raw {
		case 1:
			return "true"
		case 0:
			return "false"
		}
	}

	shift := 64 - int(v)
	return strconv.FormatInt(int64(raw<<shift)>>shift, 10)
}

func (v smallInt) resolve(stack []value, raw uint64) value {
	return value{v, raw}
}

func (v smallInt) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	return llvm.ConstInt(t, raw, false)
}

func (v smallInt) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	return true
}

/*
// bigIntValue returns a value of a provided big integer type.
func bigIntValue(ty iType, raw *big.Int) value {
	if ty <= 64 {
		panic("too small")
	}

	return value{(*bigInt)(raw), uint64(ty)}
}

// bigInt is a big integer value.
type bigInt big.Int

func (v *bigInt) typ(raw uint64) typ {
	return iType(raw)
}

func (v *bigInt) str(raw uint64) string {
	return (*big.Int)(v).String()
}

func (v *bigInt) resolve(stack []value, raw uint64) value {
	return value{v, raw}
}
*/

// offPtr is a pointer to an object with a constant offset.
// It is safe to assume that this is non-nil and does not alias any other object.
type offPtr memObj

func (v *offPtr) constant(raw uint64) bool {
	return !v.obj().stack
}

func (v *offPtr) obj() *memObj {
	return (*memObj)(v)
}

func (v *offPtr) typ(raw uint64) typ {
	return v.ptrTy
}

func (v *offPtr) str(raw uint64) string {
	str := v.obj().String()
	if raw != 0 {
		str += " + 0x" + strconv.FormatUint(raw, 16)
	}
	return str
}

func (v *offPtr) resolve(stack []value, raw uint64) value {
	return value{v, raw}
}

func (v *offPtr) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	if t.IsNil() {
		panic("t is nil")
	}
	if raw == 0 {
		ptr := gen.objPtr(v.obj())
		if ptr.Type() != t {
			if v.constant(raw) {
				ptr = llvm.ConstBitCast(ptr, t)
			} else {
				ptr = gen.builder.CreateBitCast(ptr, t, "")
				gen.applyDebug(ptr)
			}
		}
		return ptr
	}

	off := smallIntValue(v.ptrTy.idxTy(), raw)
	return uglyGEP{v.obj(), off.val}.toLLVM(t, gen, off.raw)
}

func (v *offPtr) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	objs[v.obj()] = struct{}{}
	return true
}

// offAddr is an equal-size ptrtoint of an offPtr.
type offAddr offPtr

func (v *offAddr) constant(raw uint64) bool {
	return (*offPtr)(v).constant(raw)
}

func (v *offAddr) obj() *memObj {
	return (*memObj)(v)
}

func (v *offAddr) typ(raw uint64) typ {
	return v.obj().ptrTy.idxTy()
}

func (v *offAddr) str(raw uint64) string {
	return (*offPtr)(v).str(raw)
}

func (v *offAddr) resolve(stack []value, raw uint64) value {
	return value{v, raw}
}

func (v *offAddr) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	ptr := gen.objPtr(v.obj())
	var addr llvm.Value
	isConst := v.constant(raw)
	if isConst {
		addr = llvm.ConstPtrToInt(ptr, t)
	} else {
		addr = gen.builder.CreatePtrToInt(ptr, t, "")
		gen.applyDebug(addr)
	}
	if raw == 0 {
		return addr
	}
	off := gen.value(t, smallIntValue(v.ptrTy.idxTy(), raw))
	if isConst {
		addr = llvm.ConstAdd(addr, off)
	} else {
		addr = gen.builder.CreateAdd(addr, off, "")
		gen.applyDebug(addr)
	}
	return addr
}

func (v *offAddr) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	return (*offPtr)(v).aliases(objs, raw)
}

// cast converts v to the provided type while preserving the bit representation.
// The input and output types must be non-aggregate.
// If the input is larger than the output, it is truncated as an integer.
// If the output is larger than the input, it is zero-extended as an integer.
// The input/output values are ptrtoint/inttoptr/bitcast/addrspacecasted appropriately.
func cast(to nonAggTyp, v value) value {
	if assert {
		if _, ok := v.typ().(nonAggTyp); !ok {
			panic("source type cannot be casted")
		}
	}

	fromTy := v.typ().(nonAggTyp)
	if fromTy == to {
		// No type change is required.
		return v
	}

	switch val := v.val.(type) {
	case castVal:
		// An extend(trunc(x)) is not x.
		// All other conversions can be merged.
		toBits := to.bits()
		tmpBits := val.to.bits()
		fromBits := val.val.typ(v.raw).(nonAggTyp).bits()
		if !(fromBits > tmpBits && toBits > tmpBits) {
			// Merge the conversions.
			return cast(to, value{val.val, v.raw})
		}

	case smallInt:
		switch t := to.(type) {
		case iType:
			return smallIntValue(t, v.raw)

			// TODO: bitcast int to float
		}

	case *offPtr:
		// Normalize as an addr.
		return cast(to, value{(*offAddr)(val), v.raw})

	case *offAddr:
		obj := val.obj()
		if to == obj.ptrTy {
			return value{(*offPtr)(val), v.raw}
		}
		if toBits := to.bits(); toBits <= uint64(obj.alignScale) {
			// This contains only alignment bits.
			return cast(to, smallIntValue(iType(toBits), v.raw))
		}

	case undef:
		toBits, fromBits := to.bits(), fromTy.bits()
		if toBits <= fromBits {
			// The result is completely undefined.
			return undefValue(to)
		}

	case *bitCat:
		toBits := to.bits()
		fromBits := fromTy.bits()
		switch {
		case toBits < fromBits:
			// Slice the concatenation.
			var dst []value
			var n uint64
			for _, part := range *val {
				if n == toBits {
					break
				}
				size := part.typ().(nonAggTyp).bits()
				if toBits-n < size {
					part = cast(iType(toBits-n), part)
					size = toBits - n
				}
				dst = append(dst, part)
				n += size
			}
			return cast(to, cat(dst))

		case toBits > fromBits:
			// Zero-extend the concatenation.
			src := *val
			return cast(to, cat(append(src[:len(src):len(src)], smallIntValue(iType(toBits-fromBits), 0))))
		}

	case bitSlice:
		toBits := to.bits()
		if val.width > toBits {
			// Trim the slice.
			return cast(to, slice(value{val.from, v.raw}, val.off, toBits))
		}
	}

	// Generate a symbolic cast.
	return value{castVal{v.val, to}, v.raw}
}

type castVal struct {
	val val
	to  nonAggTyp
}

func (v castVal) constant(raw uint64) bool {
	return value{v.val, raw}.constant()
}

func (v castVal) typ(raw uint64) typ {
	return v.to
}

func (v castVal) str(raw uint64) string {
	return "cast(" + value{v.val, raw}.String() + ")"
}

func (v castVal) resolve(stack []value, raw uint64) value {
	return cast(v.to, v.val.resolve(stack, raw))
}

func (v castVal) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	from := value{v.val, raw}
	fromTy := from.typ()
	switch fromTy := fromTy.(type) {
	case iType:
		switch toTy := v.to.(type) {
		case iType:
			isConstant := v.constant(raw)
			fromVal := gen.value(gen.iType(fromTy), from)
			switch {
			case toTy < fromTy:
				var v llvm.Value
				if isConstant {
					v = llvm.ConstTrunc(fromVal, t)
				} else {
					v = gen.builder.CreateTrunc(fromVal, t, "")
					gen.applyDebug(v)
				}
				return v

			case toTy > fromTy:
				var v llvm.Value
				if isConstant {
					v = llvm.ConstZExt(fromVal, t)
				} else {
					v = gen.builder.CreateZExt(fromVal, t, "")
					gen.applyDebug(v)
				}
				return v
			}

		case ptrType:
			isConstant := v.constant(raw)
			fromVal := gen.value(gen.iType(fromTy), from)
			var v llvm.Value
			if isConstant {
				v = llvm.ConstIntToPtr(fromVal, t)
			} else {
				v = gen.builder.CreateIntToPtr(fromVal, t, "")
				gen.applyDebug(v)
			}
			return v
		}

	case ptrType:
		switch toTy := v.to.(type) {
		case iType:
			isConstant := v.constant(raw)
			fromVal := gen.value(gen.ptr(gen.iType(i8), fromTy.in()), from)
			var v llvm.Value
			if isConstant {
				v = llvm.ConstPtrToInt(fromVal, t)
			} else {
				v = gen.builder.CreatePtrToInt(fromVal, t, "")
				gen.applyDebug(v)
			}
			return v

		case ptrType:
			return castVal{castVal{v.val, toTy.idxTy()}, toTy}.toLLVM(t, gen, raw)
		}
	}

	panic("bad cast")
}

func (v castVal) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	return v.val.aliases(objs, raw)
}

// uglyGEP is a pointer to an object with a dynamic offset.
// This *cannot* be safely constructed from pointer math.
// It must only be constructed from a GEP instruction.
// It is safe to assume that this is does not alias any other object.
// It is not safe to assume that this is non-nil (we lack the "inbounds" flag).
type uglyGEP struct {
	obj *memObj
	val val
}

func (v uglyGEP) constant(raw uint64) bool {
	return v.obj.ptr(0).constant() && value{v.val, raw}.constant()
}

func (v uglyGEP) typ(raw uint64) typ {
	return v.obj.ptrTy
}

func (v uglyGEP) str(raw uint64) string {
	return "uglygep(" + v.obj.String() + ", " + v.val.str(raw) + ")"
}

func (v uglyGEP) resolve(stack []value, raw uint64) value {
	return v.obj.gep(value{v.val, raw}.resolve(stack))
}

func (v uglyGEP) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	objPtr := gen.objPtr(v.obj)
	bytePtr := gen.ptr(gen.iType(i8), v.obj.ptrTy.in())
	isConst := v.constant(raw)
	if objPtrTy := objPtr.Type(); objPtrTy != bytePtr {
		if isConst {
			objPtr = llvm.ConstBitCast(objPtr, bytePtr)
		} else {
			objPtr = gen.builder.CreateBitCast(objPtr, t, "")
			gen.applyDebug(objPtr)
		}
	}
	off := gen.value(gen.iType(v.obj.ptrTy.idxTy()), value{v.val, raw})
	var ptr llvm.Value
	if isConst {
		ptr = llvm.ConstGEP(objPtr, []llvm.Value{off})
	} else {
		ptr = gen.builder.CreateGEP(objPtr, []llvm.Value{off}, "")
		gen.applyDebug(ptr)
	}
	if t != bytePtr {
		if isConst {
			ptr = llvm.ConstBitCast(ptr, t)
		} else {
			ptr = gen.builder.CreateBitCast(ptr, t, "")
			gen.applyDebug(ptr)
		}
	}
	return ptr
}

func (v uglyGEP) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	objs[v.obj] = struct{}{}
	return true
}

func structValue(t *structType, fields ...value) value {
	if len(fields) > 0 {
		// Convert to an undef value if all elements are undef.
		allUndef := true
		for _, v := range fields {
			if _, ok := v.val.(undef); !ok {
				allUndef = false
				break
			}
		}
		if allUndef {
			return undefValue(t)
		}

		// TODO: convert a struct of extracted values back to the source
	}

	isConst := true
	for _, v := range fields {
		if !v.constant() {
			isConst = false
			break
		}
	}
	var raw uint64
	if isConst {
		raw = 1
	}
	return value{&structVal{t: t, fields: fields}, raw}
}

// structVal is a composite structure value.
type structVal struct {
	t      *structType
	fields []value
}

func (v *structVal) constant(raw uint64) bool {
	return raw != 0
}

func (v *structVal) typ(raw uint64) typ {
	return v.t
}

func (v *structVal) str(raw uint64) string {
	fields := make([]string, len(v.t.fields))
	for i, f := range v.t.fields {
		fields[i] = strconv.FormatUint(f.offset, 10) + ": " + v.fields[i].String()
	}
	return "{" + strings.Join(fields, ", ") + "}"
}

func (v *structVal) resolve(stack []value, raw uint64) value {
	if v.constant(raw) {
		return value{v, raw}
	}

	fields := make([]value, len(v.fields))
	for i, v := range v.fields {
		fields[i] = v.resolve(stack)
	}
	return structValue(v.t, fields...)
}

func (v *structVal) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	isConst := v.constant(raw)
	fields := make([]llvm.Value, len(v.fields))
	elemTypes := t.StructElementTypes()
	for i, v := range v.fields {
		fields[i] = gen.value(elemTypes[i], v)
	}
	if isConst {
		return llvm.ConstNamedStruct(t, fields)
	}
	agg := llvm.Undef(t)
	for i, v := range fields {
		agg = gen.builder.CreateInsertValue(agg, v, i, "")
		gen.applyDebug(agg)
	}
	return agg
}

func (v *structVal) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	complete := true
	for _, f := range v.fields {
		complete = f.aliases(objs) && complete
	}
	return complete
}

func arrayValue(of typ, elems ...value) value {
	if uint64(len(elems)) > uint64(1<<32) {
		panic("array too big")
	}

	if len(elems) > 0 {
		// Convert to an undef value if all elements are undef.
		allUndef := true
		for _, v := range elems {
			if _, ok := v.val.(undef); !ok {
				allUndef = false
				break
			}
		}
		if allUndef {
			return undefValue(array(of, uint32(len(elems))))
		}

		// TODO: convert an array of extracted values back to the source
	}

	isConst := true
	for _, v := range elems {
		if !v.constant() {
			isConst = false
			break
		}
	}
	var raw uint64
	if isConst {
		raw = 1
	}
	return value{&arrVal{of: of, elems: elems}, raw}
}

// arrVal is an array of values.
type arrVal struct {
	of    typ
	elems []value
}

func (v *arrVal) constant(raw uint64) bool {
	return raw != 0
}

func (v *arrVal) typ(raw uint64) typ {
	return array(v.of, uint32(len(v.elems)))
}

func (v *arrVal) str(raw uint64) string {
	elems := make([]string, len(v.elems))
	for i, v := range v.elems {
		elems[i] = v.String()
	}
	return "[" + strings.Join(elems, ", ") + "]"
}

func (v *arrVal) resolve(stack []value, raw uint64) value {
	if v.constant(raw) {
		return value{v, raw}
	}

	elems := make([]value, len(v.elems))
	for i, v := range v.elems {
		elems[i] = v.resolve(stack)
	}
	return arrayValue(v.of, elems...)
}

func (v *arrVal) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	isConst := v.constant(raw)
	eTy := t.ElementType()
	elems := make([]llvm.Value, len(v.elems))
	for i, v := range v.elems {
		elems[i] = gen.value(eTy, v)
	}
	if isConst {
		return llvm.ConstArray(eTy, elems)
	}
	agg := llvm.Undef(t)
	for i, v := range elems {
		agg = gen.builder.CreateInsertValue(agg, v, i, "")
		gen.applyDebug(agg)
	}
	return agg
}

func (v *arrVal) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	complete := true
	for _, e := range v.elems {
		complete = e.aliases(objs) && complete
	}
	return complete
}

func poisonValue(ty typ) value {
	return undefValue(ty)
}

// undefValue creates an undefined constant of the specified type.
func undefValue(ty typ) value {
	if ty.bytes() == 0 {
		return ty.zero()
	}
	return value{undef{ty}, 0}
}

// undef is an undefined constant.
type undef struct {
	ty typ
}

func (v undef) constant(raw uint64) bool {
	return true
}

func (v undef) typ(raw uint64) typ {
	return v.ty
}

func (v undef) str(raw uint64) string {
	return "undef"
}

func (v undef) resolve(stack []value, raw uint64) value {
	return value{v, 0}
}

func (v undef) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	return llvm.Undef(t)
}

func (v undef) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	return true
}

func insertValue(into, val value, indices ...uint32) value {
	if len(indices) == 0 {
		panic("insertValue requires at least one index")
	}

	idx := indices[0]
	if len(indices) > 1 {
		val = insertValue(extractValue(into, idx), val, indices[1:]...)
	}
	switch t := into.typ().(type) {
	case arrType:
		elems := make([]value, t.n)
		for i := range elems {
			var v value
			if uint32(i) == idx {
				v = val
			} else {
				v = extractValue(into, uint32(i))
			}
			elems[i] = v
		}
		return arrayValue(t.of, elems...)

	case *structType:
		fields := make([]value, len(t.fields))
		for i := range fields {
			var v value
			if uint32(i) == idx {
				v = val
			} else {
				v = extractValue(into, uint32(i))
			}
			fields[i] = v
		}
		return structValue(t, fields...)

	default:
		panic("invalid value type")
	}
}

// extractValue returns the value at the specified indices of the aggregate.
func extractValue(val value, indices ...uint32) value {
	if len(indices) == 0 {
		panic("extractValue requires at least one index")
	}

	for _, idx := range indices {
		switch v := val.val.(type) {
		case undef:
			val = undefValue(v.ty.(aggTyp).sub(idx))
		case *arrVal:
			val = v.elems[idx]
		case *structVal:
			val = v.fields[idx]
		default:
			val = value{extractedValue{from: val}, uint64(idx)}
		}
	}

	return val
}

// extractedValue extracts a value from an unknown aggregate.
// This simplifies a lot of internal logic.
type extractedValue struct {
	from value
}

func (v extractedValue) constant(raw uint64) bool {
	return v.from.constant()
}

func (v extractedValue) typ(raw uint64) typ {
	return v.from.typ().(aggTyp).sub(uint32(raw))
}

func (v extractedValue) str(raw uint64) string {
	return "extract " + strconv.FormatUint(raw, 10) + " from " + v.from.String()
}

func (v extractedValue) resolve(stack []value, raw uint64) value {
	return extractValue(v.from.resolve(stack), uint32(raw))
}

func (v extractedValue) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	isConstant := v.constant(raw)
	var from llvm.Value
	switch fromTy := v.from.typ().(type) {
	case arrType:
		from = gen.value(gen.arr(t, fromTy.n), v.from)
	case *structType:
		from = gen.value(gen.sTypes[fromTy], v.from)
	default:
		panic("bad type")
	}
	var part llvm.Value
	if isConstant {
		part = llvm.ConstExtractValue(from, []uint32{uint32(raw)})
	} else {
		part = gen.builder.CreateExtractValue(from, int(raw), "")
		gen.applyDebug(part)
	}
	return part
}

func (v extractedValue) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	return v.from.aliases(objs)
}

// slice a non-aggregate value to get a range of bits as an integer.
// This is effectively (uint(in)>>off)&((1<<width)-1).
// The result is an integer "width" bits wide.
func slice(in value, off, width uint64) value {
	if assert {
		if width == 0 {
			panic("cannot slice to zero width")
		}
		if off > in.typ().(nonAggTyp).bits() || off+width > in.typ().(nonAggTyp).bits() {
			panic("slice index out of bounds")
		}
		if width > 64 {
			// Integers wider than 64 bits are not yet supported.
			panic("result is too big")
		}
	}

	if off == 0 {
		// Process this as a truncation.
		return cast(iType(width), in)
	}

	switch v := in.val.(type) {
	case smallInt:
		// Slice the integer directly by shifting and truncating.
		return smallIntValue(iType(width), in.raw>>off)

	case undef:
		// Slice the undef.
		return undefValue(iType(width))

	case castVal:
		from := value{v.val, in.raw}
		fromBits := from.typ().(nonAggTyp).bits()
		toBits := v.to.bits()
		switch {
		case off+width <= fromBits && off+width <= toBits:
			// This operates on an unmodified part of the cast.
			return slice(from, off, width)

		case off >= fromBits:
			// Extract the extended zero-padding.
			return smallIntValue(iType(width), 0)

		default:
			// Move the slice inside the cast.
			return cast(iType(width), slice(from, off, fromBits-off))
		}

	case bitSlice:
		// Combine the nested slices.
		return slice(value{v.from, in.raw}, off+v.off, width)

	case *offPtr:
		// Normalize as an addr.
		return slice(value{(*offAddr)(v), in.raw}, off, width)

	case *offAddr:
		if width+off <= uint64(v.alignScale) {
			// This contains only alignment bits.
			return smallIntValue(iType(width), in.raw>>off)
		}

	case *bitCat:
		// Filter and slice the concatenated values.
		var dst []value
		var coff uint64
		for _, v := range *v {
			if coff >= off+width {
				break
			}
			size := v.typ().(nonAggTyp).bits()
			if coff+size <= off {
				coff += size
				continue
			}
			start := coff
			if start < off {
				start = off
			}
			end := coff + size
			if end > off+width {
				end = off + width
			}
			if start != coff || end != coff+size {
				v = slice(v, start-coff, end-start)
			}
			dst = append(dst, v)
			coff = end
		}
		return cat(dst)
	}

	return value{bitSlice{in.val, off, width}, in.raw}
}

type bitSlice struct {
	from       val
	off, width uint64
}

func (v bitSlice) constant(raw uint64) bool {
	return value{v.from, raw}.constant()
}

func (v bitSlice) typ(raw uint64) typ {
	return iType(v.width)
}

func (v bitSlice) str(raw uint64) string {
	return fmt.Sprintf("bitslice(%s)[%d:%d]", value{v.from, raw}.String(), v.off, v.off+v.width)
}

func (v bitSlice) resolve(stack []value, raw uint64) value {
	return slice(value{v.from, raw}.resolve(stack), v.off, v.width)
}

func (v bitSlice) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	isConstant := v.constant(raw)
	fromTy := iType(v.from.typ(raw).(nonAggTyp).bits())
	fromTyLLVM := gen.iType(fromTy)
	fromVal := gen.value(fromTyLLVM, cast(fromTy, value{v.from, raw}))
	shift := gen.value(fromTyLLVM, smallIntValue(fromTy, v.off))
	var part llvm.Value
	if isConstant {
		part = llvm.ConstLShr(fromVal, shift)
		part = llvm.ConstTrunc(part, t)
	} else {
		part = gen.builder.CreateLShr(fromVal, shift, "")
		gen.applyDebug(part)
		part = gen.builder.CreateTrunc(part, t, "")
		gen.applyDebug(part)
	}
	return part
}

func (v bitSlice) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	return v.from.aliases(objs, raw)
}

// cat performs bitwise concatenation of non-aggregate values.
// The result is an integer value as wide as the combined inputs.
// This is effectively a shift-or chain.
func cat(in []value) value {
	var totalSize uint64
	for _, v := range in {
		totalSize += v.typ().(nonAggTyp).bits()
	}
	if assert {
		if totalSize > 64 {
			// Integers bigger than 64 bits are not yet supported.
			panic("too big")
		}
		if len(in) == 0 {
			panic("no cat inputs")
		}
	}

	// Combine values.
	var dst []value
	for _, v := range in {
		dst = catAppend(dst, v)
	}
	switch len(dst) {
	case 1:
		// This has been reduced to a single element.
		return cast(iType(totalSize), dst[0])

	case 2:
		if _, ok := dst[1].val.(smallInt); ok && dst[1].raw == 0 {
			// Turn a zero-padded value into a zero-extend.
			return cast(iType(totalSize), dst[0])
		}
	}

	res := bitCat(dst)
	return value{&res, totalSize}
}

func catAppend(dst []value, v value) []value {
	if len(dst) > 0 {
		i := len(dst) - 1
		prev := dst[i]
		switch val := v.val.(type) {
		case smallInt:
			switch prevVal := prev.val.(type) {
			case smallInt:
				// Concatenate the integers.
				lowBits := iType(prevVal).bits()
				dst[i] = smallIntValue(iType(lowBits)+iType(val), prev.raw|(v.raw<<lowBits))
				return dst
			}

		case undef:
			switch prevVal := prev.val.(type) {
			case undef:
				// Merge the undef values.
				dst[i] = undefValue(prevVal.ty.(iType) + val.ty.(iType))
				return dst
			}

		case bitSlice:
			switch prevVal := prev.val.(type) {
			case bitSlice:
				if prevVal.from == val.from && prev.raw == v.raw && prevVal.off+prevVal.width == val.off {
					// Merge the slices into a bigger slice.
					dst[i] = slice(value{val.from, v.raw}, prevVal.off, prevVal.width+val.width)
					return dst
				}

			case castVal:
				if prevVal.val == val.from && prev.raw == v.raw && prevVal.to.bits() == val.off {
					// Merge the slice with the cast.
					dst[i] = slice(value{val.from, v.raw}, 0, val.off+val.width)
					return dst
				}
			}
		}
	}
	switch val := v.val.(type) {
	case *bitCat:
		for _, v := range *val {
			dst = catAppend(dst, v)
		}
		return dst

	case castVal:
		from := value{val.val, v.raw}
		fromBits := from.typ().(nonAggTyp).bits()
		toBits := val.to.bits()
		switch {
		case toBits == fromBits:
			// Strip the cast.
			return catAppend(dst, from)

		case toBits > fromBits:
			// Split the zero extension into the raw value plus a zero integer.
			return catAppend(catAppend(dst, from), smallIntValue(iType(toBits-fromBits), 0))
		}

	case *offPtr:
		// Normalize as an addr.
		return catAppend(dst, value{(*offAddr)(val), v.raw})
	}

	return append(dst, v)
}

type bitCat []value

func (v *bitCat) constant(raw uint64) bool {
	for _, v := range *v {
		if !v.constant() {
			return false
		}
	}

	return true
}

func (v *bitCat) typ(raw uint64) typ {
	return iType(raw)
}

func (v *bitCat) str(raw uint64) string {
	parts := make([]string, len(*v))
	for i, v := range *v {
		parts[i] = v.String()
	}
	return "cat(" + strings.Join(parts, ", ") + ")"
}

func (v *bitCat) resolve(stack []value, raw uint64) value {
	parts := make([]value, len(*v))
	for i, v := range *v {
		parts[i] = v.resolve(stack)
	}
	return cat(parts)
}

func (v *bitCat) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	isConstant := v.constant(raw)
	toTy := iType(raw)
	res := gen.value(t, toTy.zero())
	var off uint64
	for _, v := range *v {
		size := v.typ().(nonAggTyp).bits()
		part := gen.value(t, cast(toTy, v))
		shift := gen.value(t, smallIntValue(toTy, off))
		if isConstant {
			part = llvm.ConstShl(part, shift)
			res = llvm.ConstOr(res, part)
		} else {
			part = gen.builder.CreateShl(part, shift, "")
			gen.applyDebug(part)
			res = gen.builder.CreateOr(res, part, "")
			gen.applyDebug(res)
		}
		off += size
	}
	return res
}

func (v *bitCat) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	complete := true
	for _, part := range *v {
		complete = part.aliases(objs) && complete
	}
	return complete
}

func runtime(ty typ, idx uint) value {
	return value{runtimeValue{ty}, uint64(idx)}
}

// runtimeValue is a reference to a value which has not yet been computed.
type runtimeValue struct {
	ty typ
}

func (v runtimeValue) constant(raw uint64) bool {
	return false
}

func (v runtimeValue) typ(raw uint64) typ {
	return v.ty
}

func (v runtimeValue) str(raw uint64) string {
	return "%" + strconv.FormatUint(raw, 10)
}

func (v runtimeValue) resolve(stack []value, raw uint64) value {
	return stack[raw]
}

func (v runtimeValue) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	val := gen.stack[raw]
	if val.Type() != t {
		val = gen.builder.CreateBitCast(val, t, "")
		gen.applyDebug(val)
	}
	return val
}

func (v runtimeValue) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	return false
}

type metadataValue llvm.Value

var _ val = metadataValue{}

func (v metadataValue) typ(raw uint64) typ {
	return metadataType{}
}

func (v metadataValue) str(raw uint64) string {
	return "????"
}

func (v metadataValue) aliases(objs map[*memObj]struct{}, raw uint64) bool {
	return true
}

func (v metadataValue) constant(raw uint64) bool {
	return true
}

func (v metadataValue) resolve(stack []value, raw uint64) value {
	return value{v, raw}
}

func (v metadataValue) toLLVM(t llvm.Type, gen *rtGen, raw uint64) llvm.Value {
	return llvm.Value(v)
}
