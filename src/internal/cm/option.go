package cm

// Option represents a Component Model [option<T>] type.
//
// [option<T>]: https://component-model.bytecodealliance.org/design/wit.html#options
type Option[T any] struct {
	_ HostLayout
	option[T]
}

// None returns an [Option] representing the none case,
// equivalent to the zero value.
func None[T any]() Option[T] {
	return Option[T]{}
}

// Some returns an [Option] representing the some case.
func Some[T any](v T) Option[T] {
	return Option[T]{
		option: option[T]{
			isSome: 1,
			some:   v,
		},
	}
}

// option represents the internal representation of a Component Model option type.
// The first byte is a byte representing the none or some case,
// followed by storage for the associated type T.
type option[T any] struct {
	_      HostLayout
	isSome uint8
	some   T
}

// None returns true if o represents the none case.
func (o *option[T]) None() bool {
	return o.isSome == 0
}

// Some returns a non-nil *T if o represents the some case,
// or nil if o represents the none case.
func (o *option[T]) Some() *T {
	if o.isSome == 1 {
		return &o.some
	}
	return nil
}

// Value returns T if o represents the some case,
// or the zero value of T if o represents the none case.
// This does not have a pointer receiver, so it can be chained.
func (o option[T]) Value() T {
	if o.isSome == 0 {
		var zero T
		return zero
	}
	return o.some
}
