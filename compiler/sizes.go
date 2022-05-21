package compiler

import (
	"go/types"
)

// The code in this file has been copied from
// https://golang.org/src/go/types/sizes.go and modified to allow for int and
// pointer sizes to differ.
// The original license can be found here:
//     https://golang.org/LICENSE

type stdSizes struct {
	IntSize  int64
	PtrSize  int64
	MaxAlign int64
}

func (s *stdSizes) Alignof(T types.Type) int64 {
	// For arrays and structs, alignment is defined in terms
	// of alignment of the elements and fields, respectively.
	switch t := T.Underlying().(type) {
	case *types.Array:
		// spec: "For a variable x of array type: unsafe.Alignof(x)
		// is the same as unsafe.Alignof(x[0]), but at least 1."
		return s.Alignof(t.Elem())
	case *types.Struct:
		// spec: "For a variable x of struct type: unsafe.Alignof(x)
		// is the largest of the values unsafe.Alignof(x.f) for each
		// field f of x, but at least 1."
		max := int64(1)
		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)
			if a := s.Alignof(f.Type()); a > max {
				max = a
			}
		}
		return max
	case *types.Slice, *types.Interface:
		// Multiword data structures are effectively structs
		// in which each element has size WordSize.
		return s.PtrSize
	case *types.Basic:
		// Strings are like slices and interfaces.
		if t.Info()&types.IsString != 0 {
			return s.PtrSize
		}
	case *types.Signature:
		// Even though functions in tinygo are 2 pointers, they are not 2 pointer aligned
		return s.PtrSize
	}

	a := s.Sizeof(T) // may be 0
	// spec: "For a variable x of any type: unsafe.Alignof(x) is at least 1."
	if a < 1 {
		return 1
	}
	// complex{64,128} are aligned like [2]float{32,64}.
	if t, ok := T.Underlying().(*types.Basic); ok && t.Info()&types.IsComplex != 0 {
		a /= 2
	}
	if a > s.MaxAlign {
		return s.MaxAlign
	}
	return a
}

func (s *stdSizes) Offsetsof(fields []*types.Var) []int64 {
	offsets := make([]int64, len(fields))
	var o int64
	for i, f := range fields {
		a := s.Alignof(f.Type())
		o = align(o, a)
		offsets[i] = o
		o += s.Sizeof(f.Type())
	}
	return offsets
}

var basicSizes = [...]byte{
	types.Bool:       1,
	types.Int8:       1,
	types.Int16:      2,
	types.Int32:      4,
	types.Int64:      8,
	types.Uint8:      1,
	types.Uint16:     2,
	types.Uint32:     4,
	types.Uint64:     8,
	types.Float32:    4,
	types.Float64:    8,
	types.Complex64:  8,
	types.Complex128: 16,
}

func (s *stdSizes) Sizeof(T types.Type) int64 {
	switch t := T.Underlying().(type) {
	case *types.Basic:
		k := t.Kind()
		if int(k) < len(basicSizes) {
			if s := basicSizes[k]; s > 0 {
				return int64(s)
			}
		}
		if k == types.String {
			return s.PtrSize * 2
		}
		if k == types.Int || k == types.Uint {
			return s.IntSize
		}
		if k == types.Uintptr {
			return s.PtrSize
		}
		if k == types.UnsafePointer {
			return s.PtrSize
		}
		if k == types.Invalid {
			return 0 // only relevant when there is a type error somewhere
		}
		panic("unknown basic type: " + t.String())
	case *types.Array:
		n := t.Len()
		if n <= 0 {
			return 0
		}
		// n > 0
		a := s.Alignof(t.Elem())
		z := s.Sizeof(t.Elem())
		return align(z, a)*(n-1) + z
	case *types.Slice:
		return s.PtrSize * 3
	case *types.Struct:
		n := t.NumFields()
		if n == 0 {
			return 0
		}
		fields := make([]*types.Var, t.NumFields())
		maxAlign := int64(1)
		for i := range fields {
			field := t.Field(i)
			fields[i] = field
			al := s.Alignof(field.Type())
			if al > maxAlign {
				maxAlign = al
			}
		}
		// Pick the size that fits this struct and add some alignment. Some
		// structs have some extra padding at the end which should also be taken
		// care of:
		//     struct { int32 n; byte b }
		offsets := s.Offsetsof(fields)
		return align(offsets[n-1]+s.Sizeof(fields[n-1].Type()), maxAlign)
	case *types.Interface:
		return s.PtrSize * 2
	case *types.Pointer:
		return s.PtrSize
	case *types.Signature:
		// Func values in TinyGo are two words in size.
		return s.PtrSize * 2
	default:
		panic("unknown type: " + t.String())
	}
}

// align returns the smallest y >= x such that y % a == 0.
func align(x, a int64) int64 {
	y := x + a - 1
	return y - y%a
}
