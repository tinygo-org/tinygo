package cm

import "unsafe"

// Discriminant is the set of types that can represent the tag or discriminator of a variant.
// Use bool for 2-case variant types, result<T>, or option<T> types, uint8 where there are 256 or
// fewer cases, uint16 for up to 65,536 cases, or uint32 for anything greater.
type Discriminant interface {
	bool | uint8 | uint16 | uint32
}

// Variant represents a loosely-typed Component Model variant.
// Shape and Align must be non-zero sized types. To create a variant with no associated
// types, use an enum.
type Variant[Tag Discriminant, Shape, Align any] struct{ variant[Tag, Shape, Align] }

// NewVariant returns a [Variant] with tag of type Disc, storage and GC shape of type Shape,
// aligned to type Align, with a value of type T.
func NewVariant[Tag Discriminant, Shape, Align any, T any](tag Tag, data T) Variant[Tag, Shape, Align] {
	validateVariant[Tag, Shape, Align, T]()
	var v Variant[Tag, Shape, Align]
	v.tag = tag
	*(*T)(unsafe.Pointer(&v.data)) = data
	return v
}

// New returns a [Variant] with tag of type Disc, storage and GC shape of type Shape,
// aligned to type Align, with a value of type T.
func New[V ~struct{ variant[Tag, Shape, Align] }, Tag Discriminant, Shape, Align any, T any](tag Tag, data T) V {
	validateVariant[Tag, Shape, Align, T]()
	var v variant[Tag, Shape, Align]
	v.tag = tag
	*(*T)(unsafe.Pointer(&v.data)) = data
	return *(*V)(unsafe.Pointer(&v))
}

// Case returns a non-nil *T if the [Variant] case is equal to tag, otherwise it returns nil.
func Case[T any, V ~struct{ variant[Tag, Shape, Align] }, Tag Discriminant, Shape, Align any](v *V, tag Tag) *T {
	validateVariant[Tag, Shape, Align, T]()
	v2 := (*variant[Tag, Shape, Align])(unsafe.Pointer(v))
	if v2.tag == tag {
		return (*T)(unsafe.Pointer(&v2.data))
	}
	return nil
}

// variant is the internal representation of a Component Model variant.
// Shape and Align must be non-zero sized types.
type variant[Tag Discriminant, Shape, Align any] struct {
	tag  Tag
	_    [0]Align
	data Shape // [unsafe.Sizeof(*(*Shape)(unsafe.Pointer(nil)))]byte
}

// Tag returns the tag (discriminant) of variant v.
func (v *variant[Tag, Shape, Align]) Tag() Tag {
	return v.tag
}

// This function is sized so it can be inlined and optimized away.
func validateVariant[Disc Discriminant, Shape, Align any, T any]() {
	var v variant[Disc, Shape, Align]
	var t T

	// Check if size of T is greater than Shape
	if unsafe.Sizeof(t) > unsafe.Sizeof(v.data) {
		panic("variant: size of requested type > data type")
	}

	// Check if Shape is zero-sized, but size of result != 1
	if unsafe.Sizeof(v.data) == 0 && unsafe.Sizeof(v) != 1 {
		panic("variant: size of data type == 0, but variant size != 1")
	}
}
