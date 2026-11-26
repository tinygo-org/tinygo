package cm

import "unsafe"

const (
	// ResultOK represents the OK case of a result.
	ResultOK = 0

	// ResultErr represents the error case of a result.
	ResultErr = 1
)

// BoolResult represents a result with no OK or error type.
// False represents the OK case and true represents the error case.
type BoolResult uint8

// Result represents a result sized to hold the Shape type.
// The size of the Shape type must be greater than or equal to the size of OK and Err types.
// For results with two zero-length types, use [BoolResult].
type Result[Shape, OK, Err any] struct {
	_ HostLayout
	result[Shape, OK, Err]
}

// AnyResult is a type constraint for generic functions that accept any [Result] type.
type AnyResult[Shape, OK, Err any] interface {
	~struct {
		_ HostLayout
		result[Shape, OK, Err]
	}
}

// result represents the internal representation of a Component Model result type.
type result[Shape, OK, Err any] struct {
	_     HostLayout
	isErr uint8
	_     [0]OK
	_     [0]Err
	data  Shape // [unsafe.Sizeof(*(*Shape)(unsafe.Pointer(nil)))]byte
}

// OK returns an OK result with shape Shape and type OK and Err.
// Pass Result[OK, OK, Err] or Result[Err, OK, Err] as the first type argument.
func OK[R AnyResult[Shape, OK, Err], Shape, OK, Err any](ok OK) R {
	var r Result[Shape, OK, Err]
	r.validate()
	r.isErr = ResultOK
	*((*OK)(unsafe.Pointer(&r.data))) = ok
	return R(r)
}

// Err returns an error result with shape Shape and type OK and Err.
// Pass Result[OK, OK, Err] or Result[Err, OK, Err] as the first type argument.
func Err[R AnyResult[Shape, OK, Err], Shape, OK, Err any](err Err) R {
	var r Result[Shape, OK, Err]
	r.validate()
	r.isErr = ResultErr
	*((*Err)(unsafe.Pointer(&r.data))) = err
	return R(r)
}

// SetOK sets r to an OK result.
func (r *result[Shape, OK, Err]) SetOK(ok OK) {
	r.validate()
	r.isErr = ResultOK
	*((*OK)(unsafe.Pointer(&r.data))) = ok
}

// SetErr sets r to an error result.
func (r *result[Shape, OK, Err]) SetErr(err Err) {
	r.validate()
	r.isErr = ResultErr
	*((*Err)(unsafe.Pointer(&r.data))) = err
}

// IsOK returns true if r represents the OK case.
func (r *result[Shape, OK, Err]) IsOK() bool {
	r.validate()
	return r.isErr == ResultOK
}

// IsErr returns true if r represents the error case.
func (r *result[Shape, OK, Err]) IsErr() bool {
	r.validate()
	return r.isErr == ResultErr
}

// OK returns a non-nil *OK pointer if r represents the OK case.
// If r represents an error, then it returns nil.
func (r *result[Shape, OK, Err]) OK() *OK {
	r.validate()
	if r.isErr == ResultErr {
		return nil
	}
	return (*OK)(unsafe.Pointer(&r.data))
}

// Err returns a non-nil *Err pointer if r represents the error case.
// If r represents the OK case, then it returns nil.
func (r *result[Shape, OK, Err]) Err() *Err {
	r.validate()
	if r.isErr == ResultOK {
		return nil
	}
	return (*Err)(unsafe.Pointer(&r.data))
}

// Result returns (OK, zero value of Err, false) if r represents the OK case,
// or (zero value of OK, Err, true) if r represents the error case.
// This does not have a pointer receiver, so it can be chained.
func (r result[Shape, OK, Err]) Result() (ok OK, err Err, isErr bool) {
	if r.isErr == ResultErr {
		return ok, *(*Err)(unsafe.Pointer(&r.data)), true
	}
	return *(*OK)(unsafe.Pointer(&r.data)), err, false
}

// This function is sized so it can be inlined and optimized away.
func (r *result[Shape, OK, Err]) validate() {
	var shape Shape
	var ok OK
	var err Err

	// Check if size of Shape is greater than both OK and Err
	if unsafe.Sizeof(shape) > unsafe.Sizeof(ok) && unsafe.Sizeof(shape) > unsafe.Sizeof(err) {
		panic("result: size of data type > OK and Err types")
	}

	// Check if size of OK is greater than Shape
	if unsafe.Sizeof(ok) > unsafe.Sizeof(shape) {
		panic("result: size of OK type > data type")
	}

	// Check if size of Err is greater than Shape
	if unsafe.Sizeof(err) > unsafe.Sizeof(shape) {
		panic("result: size of Err type > data type")
	}

	// Check if Shape is zero-sized, but size of result != 1
	if unsafe.Sizeof(shape) == 0 && unsafe.Sizeof(*r) != 1 {
		panic("result: size of data type == 0, but result size != 1")
	}
}
