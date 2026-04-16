// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

var inf = float64frombits(0x7FF0000000000000)

// isNaN reports whether f is an IEEE 754 “not-a-number” value.
func isNaN(f float64) (is bool) {
	// IEEE 754 says that only NaNs satisfy f != f.
	return f != f
}

// isFinite reports whether f is neither NaN nor an infinity.
func isFinite(f float64) bool {
	return !isNaN(f - f)
}

// isInf reports whether f is an infinity.
func isInf(f float64) bool {
	return !isNaN(f) && !isFinite(f)
}

// Abs returns the absolute value of x.
//
// Special cases are:
//
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func abs(x float64) float64 {
	const sign = 1 << 63
	return float64frombits(float64bits(x) &^ sign)
}

// copysign returns a value with the magnitude
// of x and the sign of y.
func copysign(x, y float64) float64 {
	const sign = 1 << 63
	return float64frombits(float64bits(x)&^sign | float64bits(y)&sign)
}

// Float64bits returns the IEEE 754 binary representation of f.
func float64bits(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

// Float64frombits returns the floating point number corresponding
// the IEEE 754 binary representation b.
func float64frombits(b uint64) float64 {
	return *(*float64)(unsafe.Pointer(&b))
}

// The fmimimum/fmaximum are missing from most libm implementations.
// Just define them ourselves.

//export fminimum
func fminimum(x, y float64) float64 {
	return minimumFloat64(x, y)
}

//export fminimumf
func fminimumf(x, y float32) float32 {
	return minimumFloat32(x, y)
}

//export fmaximum
func fmaximum(x, y float64) float64 {
	return maximumFloat64(x, y)
}

//export fmaximumf
func fmaximumf(x, y float32) float32 {
	return maximumFloat32(x, y)
}

// Create separate copies of the function that are not exported.
// This is necessary so that LLVM does not recognize them as builtins.
// If tests called the builtins, LLVM would just override them on most platforms.

func minimumFloat32(x, y float32) float32 {
	return minimumFloat[float32, int32](x, y, minPosNaN32, magMask32)
}

func minimumFloat64(x, y float64) float64 {
	return minimumFloat[float64, int64](x, y, minPosNaN64, magMask64)
}

func maximumFloat32(x, y float32) float32 {
	return maximumFloat[float32, int32](x, y, minPosNaN32, magMask32)
}

func maximumFloat64(x, y float64) float64 {
	return maximumFloat[float64, int64](x, y, minPosNaN64, magMask64)
}

// minimumFloat is a generic implementation of the floating-point minimum operation.
// This implementation uses integer operations because this is mainly used for platforms without an FPU.
func minimumFloat[T float, I floatInt](x, y T, minPosNaN, magMask I) T {
	xBits := *(*I)(unsafe.Pointer(&x))
	yBits := *(*I)(unsafe.Pointer(&y))

	// Handle the special case of a positive NaN value.
	switch {
	case xBits >= minPosNaN:
		return x
	case yBits >= minPosNaN:
		return y
	}

	// The exponent-mantissa portion of the float is comparable via unsigned comparison (excluding the NaN case).
	// We can turn a float into a signed-comparable value by reversing the comparison order of negative values.
	// We can reverse the order by inverting the bits.
	// This also ensures that positive zero compares greater than negative zero (as required by the spec).
	// Negative NaN values will compare less than any other value, so they require no special handling to propagate.
	if xBits < 0 {
		xBits ^= magMask
	}
	if yBits < 0 {
		yBits ^= magMask
	}
	if xBits <= yBits {
		return x
	} else {
		return y
	}
}

// maximumFloat is a generic implementation of the floating-point maximum operation.
// This implementation uses integer operations because this is mainly used for platforms without an FPU.
func maximumFloat[T float, I floatInt](x, y T, minPosNaN, magMask I) T {
	xBits := *(*I)(unsafe.Pointer(&x))
	yBits := *(*I)(unsafe.Pointer(&y))

	// The exponent-mantissa portion of the float is comparable via unsigned comparison (excluding the NaN case).
	// We can turn a float into a signed-comparable value by reversing the comparison order of negative values.
	// We can reverse the order by inverting the bits.
	// This also ensures that positive zero compares greater than negative zero (as required by the spec).
	// Positive NaN values will compare greater than any other value, so they require no special handling to propagate.
	if xBits < 0 {
		xBits ^= magMask
	}
	if yBits < 0 {
		yBits ^= magMask
	}
	// Handle the special case of a negative NaN value.
	maxNegNaN := ^minPosNaN
	switch {
	case xBits <= maxNegNaN:
		return x
	case yBits <= maxNegNaN:
		return y
	}
	if xBits >= yBits {
		return x
	} else {
		return y
	}
}

const (
	signPos64     = 63
	exponentPos64 = 52
	minPosNaN64   = ((1 << signPos64) - (1 << exponentPos64)) + 1
	magMask64     = 1<<signPos64 - 1

	signPos32     = 31
	exponentPos32 = 23
	minPosNaN32   = ((1 << signPos32) - (1 << exponentPos32)) + 1
	magMask32     = 1<<signPos32 - 1
)

type float interface {
	float32 | float64
}

type floatInt interface {
	int32 | int64
}
