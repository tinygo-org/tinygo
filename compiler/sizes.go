package compiler

import (
	"go/types"
)

// The code in this file has been copied from
// https://golang.org/src/go/types/sizes.go and modified to allow for int and
// pointer sizes to differ.
// The original license can be found here:
//     https://golang.org/LICENSE

type StdSizes struct {
	IntSize  int64
	PtrSize  int64
	MaxAlign int64
}

func (s *StdSizes) Alignof(T types.Type) int64 {
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

func (s *StdSizes) Offsetsof(fields []*types.Var) []int64 {
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

func (s *StdSizes) Sizeof(T types.Type) int64 {
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
		for i := range fields {
			fields[i] = t.Field(i)
		}
		offsets := s.Offsetsof(fields)
		return offsets[n-1] + s.Sizeof(fields[n-1].Type())
	case *types.Interface:
		return s.PtrSize * 2
	case *types.Pointer:
		return s.PtrSize
	case *types.Signature:
		params := t.Params()
		// We assume that function reference has the same size as the data pointer.
		// This is not necessarily true for targets that use separate address space
		// for functions (e.g. WebAssembly tables), but so far seems to be true
		// for all supported targets.
		useExportAbi := params.Len() > 0 && params.At(0).Name() == "_exportABI"
		if useExportAbi {
			return s.PtrSize
		} else {
			return s.PtrSize * 2
		}
	default:
		panic("unknown type: " + t.String())
	}
}

// align returns the smallest y >= x such that y % a == 0.
func align(x, a int64) int64 {
	y := x + a - 1
	return y - y%a
}
