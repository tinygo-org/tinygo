package interp

import (
	"fmt"
	"strconv"
	"strings"
)

type typ interface {
	fmt.Stringer
	bytes() uint64
	zero() value
}

type nonAggTyp interface {
	typ

	bits() uint64
}

// iType is an integer type.
type iType uint32

const (
	i1  iType = 1
	i8  iType = 8
	i16 iType = 16
	i32 iType = 32
	i64 iType = 64
)

func (t iType) String() string {
	return "i" + strconv.FormatUint(uint64(t), 10)
}

func (t iType) bits() uint64 {
	return uint64(t)
}

func (t iType) bytes() uint64 {
	return (t.bits() + 7) / 8
}

func (t iType) zero() value {
	return smallIntValue(t, 0)
}

// floatType is a single-precision floating point type.
type floatType struct{}

func (t floatType) String() string {
	return "float"
}

func (t floatType) bits() uint64 {
	return 32
}

func (t floatType) bytes() uint64 {
	return t.bits() / 8
}

func (t floatType) zero() value {
	return floatValue(0)
}

// doubleType is a double-precision floating point type.
type doubleType struct{}

func (t doubleType) String() string {
	return "double"
}

func (t doubleType) bits() uint64 {
	return 64
}

func (t doubleType) bytes() uint64 {
	return t.bits() / 8
}

func (t doubleType) zero() value {
	return doubleValue(0)
}

// pointer returns a pointer type.
// The "in" input is the address space of the pointer type.
// The "idxTy" input is the type used to index bytes (e.g. i64 on a 64-bit target).
func pointer(in addrSpace, idxTy iType) ptrType {
	if idxTy > maxPtrIdxWidth {
		panic("pointer index type is too big")
	}
	if in > maxPtrAddrSpace {
		panic("address space index too big")
	}

	return ptrType((64 * uint(in)) | uint(idxTy-1))
}

// ptrType is a pointer type.
type ptrType uint32

const (
	maxPtrIdxWidth  = 64
	maxPtrAddrSpace = addrSpace((^uint32(0)) / 64)
)

func (t ptrType) String() string {
	return "ptr(in " + t.in().String() + ", idx " + t.idxTy().String() + ")"
}

func (t ptrType) idxTy() iType {
	return iType((t % 64) + 1)
}

func (t ptrType) in() addrSpace {
	return addrSpace(t / 64)
}

func (t ptrType) bits() uint64 {
	return t.idxTy().bits()
}

func (t ptrType) bytes() uint64 {
	return t.idxTy().bytes()
}

func (t ptrType) zero() value {
	return cast(t, t.idxTy().zero())
}

// addrSpace is a memory address space.
type addrSpace uint32

const defaultAddrSpace addrSpace = 0

func (s addrSpace) String() string {
	if s == 0 {
		return "default"
	}

	return fmt.Sprintf("addrspace(%d)", uint(s))
}

type aggTyp interface {
	typ

	sub(idx uint32) typ
	elems() uint32
}

// structType is a tuple of values.
type structType struct {
	name   string
	fields []structField
	size   uint64
}

type structField struct {
	ty     typ
	offset uint64
}

func (t *structType) String() string {
	if t.name != "" {
		return "%" + maybeQuoteName(t.name)
	}

	fields := make([]string, len(t.fields))[:0]
	var j uint64
	for _, v := range t.fields {
		if v.offset != j {
			fields = append(fields, strconv.FormatUint(j, 10)+": _ x "+strconv.FormatUint(v.offset-j, 10))
		}
		fields = append(fields, strconv.FormatUint(v.offset, 10)+": "+v.ty.String())
		j = v.offset + v.ty.bytes()
	}
	if size := t.size; j != size {
		fields = append(fields, strconv.FormatUint(j, 10)+": _ x "+strconv.FormatUint(size-j, 10))
	}
	return "{" + strings.Join(fields, ", ") + "}"
}

func maybeQuoteName(name string) string {
	var needQuote bool
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		default:
			switch r {
			case '$', '.', '_':
			default:
				needQuote = true
			}
		}
	}
	if needQuote {
		return strconv.Quote(name)
	}
	return name
}

func (t *structType) sub(idx uint32) typ {
	return t.fields[idx].ty
}

func (t *structType) elems() uint32 {
	return uint32(len(t.fields))
}

func (t *structType) bytes() uint64 {
	return t.size
}

func (t *structType) zero() value {
	fields := make([]value, len(t.fields))
	for i, f := range t.fields {
		fields[i] = f.ty.zero()
	}
	return structValue(t, fields...)
}

// array creates an array type.
func array(of typ, n uint32) arrType {
	return arrType{of, n}
}

// arrType is an array of values.
// Okay these descriptions are getting bad.
type arrType struct {
	of typ
	n  uint32
}

func (t arrType) String() string {
	return "[" + strconv.FormatUint(uint64(t.n), 10) + " x " + t.of.String() + "]"
}

func (t arrType) sub(idx uint32) typ {
	if idx >= t.n {
		panic("invalid array index")
	}
	return t.of
}

func (t arrType) elems() uint32 {
	return t.n
}

func (t arrType) bytes() uint64 {
	return uint64(t.n) * t.of.bytes()
}

func (t arrType) zero() value {
	elemZero := t.of.zero()
	elements := make([]value, t.n)
	for i := range elements {
		elements[i] = elemZero
	}
	return arrayValue(t.of, elements...)
}

type metadataType struct{}

var _ typ = metadataType{}

func (t metadataType) String() string {
	return "metadata"
}
func (t metadataType) bytes() uint64 {
	return 0
}

func (t metadataType) zero() value {
	panic("metadata has no zero value")
}
