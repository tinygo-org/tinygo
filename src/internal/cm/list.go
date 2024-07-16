package cm

import "unsafe"

// List represents a Component Model list.
// The binary representation of list<T> is similar to a Go slice minus the cap field.
type List[T any] struct{ list[T] }

// NewList returns a List[T] from data and len.
func NewList[T any](data *T, len uint) List[T] {
	return List[T]{
		list[T]{
			data: data,
			len:  len,
		},
	}
}

// ToList returns a List[T] equivalent to the Go slice s.
// The underlying slice data is not copied, and the resulting List points at the
// same array storage as the slice.
func ToList[S ~[]T, T any](s S) List[T] {
	return NewList[T](unsafe.SliceData([]T(s)), uint(len(s)))
}

// list represents the internal representation of a Component Model list.
// It is intended to be embedded in a [List], so embedding types maintain
// the methods defined on this type.
type list[T any] struct {
	data *T
	len  uint
}

// Slice returns a Go slice representing the List.
func (l list[T]) Slice() []T {
	return unsafe.Slice(l.data, l.len)
}

// Data returns the data pointer for the list.
func (l list[T]) Data() *T {
	return l.data
}

// Len returns the length of the list.
// TODO: should this return an int instead of a uint?
func (l list[T]) Len() uint {
	return l.len
}
