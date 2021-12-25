// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package bits implements bit counting and manipulation
// functions for the predeclared unsigned integer types.
package bits

import _ "unsafe"

const uintSize = 32 << (^uint(0) >> 63) // 32 or 64

// UintSize is the size of a uint in bits.
const UintSize = uintSize

// --- LeadingZeros ---

// LeadingZeros returns the number of leading zero bits in x; the result is UintSize for x == 0.
func LeadingZeros(x uint) int {
	if UintSize == 32 {
		return LeadingZeros32(uint32(x))
	}
	return LeadingZeros64(uint64(x))
}

// LeadingZeros8 returns the number of leading zero bits in x; the result is 8 for x == 0.
func LeadingZeros8(x uint8) int {
	return int(leadingZeros8(x, false))
}

//export llvm.ctlz.i8
func leadingZeros8(x uint8, isZeroUndef bool) uint8

// LeadingZeros16 returns the number of leading zero bits in x; the result is 16 for x == 0.
func LeadingZeros16(x uint16) int {
	return int(leadingZeros16(x, false))
}

//export llvm.ctlz.i16
func leadingZeros16(x uint16, isZeroUndef bool) uint16

// LeadingZeros32 returns the number of leading zero bits in x; the result is 32 for x == 0.
func LeadingZeros32(x uint32) int {
	return int(leadingZeros32(x, false))
}

//export llvm.ctlz.i32
func leadingZeros32(x uint32, isZeroUndef bool) uint32

// LeadingZeros64 returns the number of leading zero bits in x; the result is 64 for x == 0.
func LeadingZeros64(x uint64) int {
	return int(leadingZeros64(x, false))
}

//export llvm.ctlz.i64
func leadingZeros64(x uint64, isZeroUndef bool) uint64

// --- TrailingZeros ---

// TrailingZeros returns the number of trailing zero bits in x; the result is UintSize for x == 0.
func TrailingZeros(x uint) int {
	if UintSize == 32 {
		return TrailingZeros32(uint32(x))
	}
	return TrailingZeros64(uint64(x))
}

// TrailingZeros8 returns the number of trailing zero bits in x; the result is 8 for x == 0.
func TrailingZeros8(x uint8) int {
	return int(trailingZeros8(x, false))
}

//export llvm.cttz.i8
func trailingZeros8(x uint8, isZeroUndef bool) uint8

// TrailingZeros16 returns the number of trailing zero bits in x; the result is 16 for x == 0.
func TrailingZeros16(x uint16) int {
	return int(trailingZeros16(x, false))
}

//export llvm.cttz.i16
func trailingZeros16(x uint16, isZeroUndef bool) uint16

// TrailingZeros32 returns the number of trailing zero bits in x; the result is 32 for x == 0.
func TrailingZeros32(x uint32) int {
	return int(trailingZeros32(x, false))
}

//export llvm.cttz.i32
func trailingZeros32(x uint32, isZeroUndef bool) uint32

// TrailingZeros64 returns the number of trailing zero bits in x; the result is 64 for x == 0.
func TrailingZeros64(x uint64) int {
	return int(trailingZeros64(x, false))
}

//export llvm.cttz.i64
func trailingZeros64(x uint64, isZeroUndef bool) uint64

// --- OnesCount ---

// OnesCount returns the number of one bits ("population count") in x.
func OnesCount(x uint) int {
	if UintSize == 32 {
		return OnesCount32(uint32(x))
	}
	return OnesCount64(uint64(x))
}

// OnesCount8 returns the number of one bits ("population count") in x.
func OnesCount8(x uint8) int {
	return int(onesCount8(x))
}

//export llvm.ctpop.i8
func onesCount8(x uint8) uint8

// OnesCount16 returns the number of one bits ("population count") in x.
func OnesCount16(x uint16) int {
	return int(onesCount16(x))
}

//export llvm.ctpop.i16
func onesCount16(x uint16) uint16

// OnesCount32 returns the number of one bits ("population count") in x.
func OnesCount32(x uint32) int {
	return int(onesCount32(x))
}

//export llvm.ctpop.i32
func onesCount32(x uint32) uint32

// OnesCount64 returns the number of one bits ("population count") in x.
func OnesCount64(x uint64) int {
	return int(onesCount64(x))
}

//export llvm.ctpop.i64
func onesCount64(x uint64) uint64

// --- RotateLeft ---

// RotateLeft returns the value of x rotated left by (k mod UintSize) bits.
// To rotate x right by k bits, call RotateLeft(x, -k).
//
// This function's execution time does not depend on the inputs.
func RotateLeft(x uint, k int) uint {
	if UintSize == 32 {
		return uint(RotateLeft32(uint32(x), k))
	}
	return uint(RotateLeft64(uint64(x), k))
}

// RotateLeft8 returns the value of x rotated left by (k mod 8) bits.
// To rotate x right by k bits, call RotateLeft8(x, -k).
//
// This function's execution time does not depend on the inputs.
func RotateLeft8(x uint8, k int) uint8 {
	const n = 8
	s := uint(k) & (n - 1)
	return x<<s | x>>(n-s)
}

// RotateLeft16 returns the value of x rotated left by (k mod 16) bits.
// To rotate x right by k bits, call RotateLeft16(x, -k).
//
// This function's execution time does not depend on the inputs.
func RotateLeft16(x uint16, k int) uint16 {
	const n = 16
	s := uint(k) & (n - 1)
	return x<<s | x>>(n-s)
}

// RotateLeft32 returns the value of x rotated left by (k mod 32) bits.
// To rotate x right by k bits, call RotateLeft32(x, -k).
//
// This function's execution time does not depend on the inputs.
func RotateLeft32(x uint32, k int) uint32 {
	const n = 32
	s := uint(k) & (n - 1)
	return x<<s | x>>(n-s)
}

// RotateLeft64 returns the value of x rotated left by (k mod 64) bits.
// To rotate x right by k bits, call RotateLeft64(x, -k).
//
// This function's execution time does not depend on the inputs.
func RotateLeft64(x uint64, k int) uint64 {
	const n = 64
	s := uint(k) & (n - 1)
	return x<<s | x>>(n-s)
}

// --- Reverse ---

// Reverse returns the value of x with its bits in reversed order.
func Reverse(x uint) uint {
	if UintSize == 32 {
		return uint(Reverse32(uint32(x)))
	}
	return uint(Reverse64(uint64(x)))
}

// Reverse8 returns the value of x with its bits in reversed order.
func Reverse8(x uint8) uint8 {
	return reverse8(x)
}

//export llvm.bitreverse.i8
func reverse8(x uint8) uint8

// Reverse16 returns the value of x with its bits in reversed order.
func Reverse16(x uint16) uint16 {
	return reverse16(x)
}

//export llvm.bitreverse.i16
func reverse16(x uint16) uint16

// Reverse32 returns the value of x with its bits in reversed order.
func Reverse32(x uint32) uint32 {
	return reverse32(x)
}

//export llvm.bitreverse.i32
func reverse32(x uint32) uint32

// Reverse64 returns the value of x with its bits in reversed order.
func Reverse64(x uint64) uint64 {
	return reverse64(x)
}

//export llvm.bitreverse.i64
func reverse64(x uint64) uint64

// --- ReverseBytes ---

// ReverseBytes returns the value of x with its bytes in reversed order.
//
// This function's execution time does not depend on the inputs.
func ReverseBytes(x uint) uint {
	if UintSize == 32 {
		return uint(ReverseBytes32(uint32(x)))
	}
	return uint(ReverseBytes64(uint64(x)))
}

// ReverseBytes16 returns the value of x with its bytes in reversed order.
//
// This function's execution time does not depend on the inputs.
func ReverseBytes16(x uint16) uint16 {
	return reverseBytes16(x)
}

//export llvm.bswap.i16
func reverseBytes16(x uint16) uint16

// ReverseBytes32 returns the value of x with its bytes in reversed order.
//
// This function's execution time does not depend on the inputs.
func ReverseBytes32(x uint32) uint32 {
	return reverseBytes32(x)
}

//export llvm.bswap.i32
func reverseBytes32(x uint32) uint32

// ReverseBytes64 returns the value of x with its bytes in reversed order.
//
// This function's execution time does not depend on the inputs.
func ReverseBytes64(x uint64) uint64 {
	return reverseBytes64(x)
}

//export llvm.bswap.i64
func reverseBytes64(x uint64) uint64

// --- Len ---

// Len returns the minimum number of bits required to represent x; the result is 0 for x == 0.
func Len(x uint) int {
	if UintSize == 32 {
		return Len32(uint32(x))
	}
	return Len64(uint64(x))
}

// Len8 returns the minimum number of bits required to represent x; the result is 0 for x == 0.
func Len8(x uint8) int {
	return 8 - LeadingZeros8(x)
}

// Len16 returns the minimum number of bits required to represent x; the result is 0 for x == 0.
func Len16(x uint16) (n int) {
	return 16 - LeadingZeros16(x)
}

// Len32 returns the minimum number of bits required to represent x; the result is 0 for x == 0.
func Len32(x uint32) (n int) {
	return 32 - LeadingZeros32(x)
}

// Len64 returns the minimum number of bits required to represent x; the result is 0 for x == 0.
func Len64(x uint64) (n int) {
	return 64 - LeadingZeros64(x)
}

// --- Add with carry ---
// Currently, LLVM only provides intrinsics which support carry out.
// There is no intrinsic which supports carry-in.
// Additionally, those intrinsics return i1, which Go can only represent as bool.

// Add returns the sum with carry of x, y and carry: sum = x + y + carry.
// The carry input must be 0 or 1; otherwise the behavior is undefined.
// The carryOut output is guaranteed to be 0 or 1.
//
// This function's execution time does not depend on the inputs.
func Add(x, y, carry uint) (sum, carryOut uint) {
	if UintSize == 32 {
		s32, c32 := Add32(uint32(x), uint32(y), uint32(carry))
		return uint(s32), uint(c32)
	}
	s64, c64 := Add64(uint64(x), uint64(y), uint64(carry))
	return uint(s64), uint(c64)
}

// Add32 returns the sum with carry of x, y and carry: sum = x + y + carry.
// The carry input must be 0 or 1; otherwise the behavior is undefined.
// The carryOut output is guaranteed to be 0 or 1.
//
// This function's execution time does not depend on the inputs.
func Add32(x, y, carry uint32) (sum, carryOut uint32) {
	sum64 := uint64(x) + uint64(y) + uint64(carry)
	sum = uint32(sum64)
	carryOut = uint32(sum64 >> 32)
	return
}

// Add64 returns the sum with carry of x, y and carry: sum = x + y + carry.
// The carry input must be 0 or 1; otherwise the behavior is undefined.
// The carryOut output is guaranteed to be 0 or 1.
//
// This function's execution time does not depend on the inputs.
func Add64(x, y, carry uint64) (sum, carryOut uint64) {
	sum = x + y + carry
	// The sum will overflow if both top bits are set (x & y) or if one of them
	// is (x | y), and a carry from the lower place happened. If such a carry
	// happens, the top bit will be 1 + 0 + 1 = 0 (&^ sum).
	carryOut = ((x & y) | ((x | y) &^ sum)) >> 63
	return
}

// --- Subtract with borrow ---
// This has the same issue as "Add with carry".

// Sub returns the difference of x, y and borrow: diff = x - y - borrow.
// The borrow input must be 0 or 1; otherwise the behavior is undefined.
// The borrowOut output is guaranteed to be 0 or 1.
//
// This function's execution time does not depend on the inputs.
func Sub(x, y, borrow uint) (diff, borrowOut uint) {
	if UintSize == 32 {
		d32, b32 := Sub32(uint32(x), uint32(y), uint32(borrow))
		return uint(d32), uint(b32)
	}
	d64, b64 := Sub64(uint64(x), uint64(y), uint64(borrow))
	return uint(d64), uint(b64)
}

// Sub32 returns the difference of x, y and borrow, diff = x - y - borrow.
// The borrow input must be 0 or 1; otherwise the behavior is undefined.
// The borrowOut output is guaranteed to be 0 or 1.
//
// This function's execution time does not depend on the inputs.
func Sub32(x, y, borrow uint32) (diff, borrowOut uint32) {
	diff = x - y - borrow
	// The difference will underflow if the top bit of x is not set and the top
	// bit of y is set (^x & y) or if they are the same (^(x ^ y)) and a borrow
	// from the lower place happens. If that borrow happens, the result will be
	// 1 - 1 - 1 = 0 - 0 - 1 = 1 (& diff).
	borrowOut = ((^x & y) | (^(x ^ y) & diff)) >> 31
	return
}

// Sub64 returns the difference of x, y and borrow: diff = x - y - borrow.
// The borrow input must be 0 or 1; otherwise the behavior is undefined.
// The borrowOut output is guaranteed to be 0 or 1.
//
// This function's execution time does not depend on the inputs.
func Sub64(x, y, borrow uint64) (diff, borrowOut uint64) {
	diff = x - y - borrow
	// See Sub32 for the bit logic.
	borrowOut = ((^x & y) | (^(x ^ y) & diff)) >> 63
	return
}

// --- Full-width multiply ---

// Mul returns the full-width product of x and y: (hi, lo) = x * y
// with the product bits' upper half returned in hi and the lower
// half returned in lo.
//
// This function's execution time does not depend on the inputs.
func Mul(x, y uint) (hi, lo uint) {
	if UintSize == 32 {
		h, l := Mul32(uint32(x), uint32(y))
		return uint(h), uint(l)
	}
	h, l := Mul64(uint64(x), uint64(y))
	return uint(h), uint(l)
}

// Mul32 returns the 64-bit product of x and y: (hi, lo) = x * y
// with the product bits' upper half returned in hi and the lower
// half returned in lo.
//
// This function's execution time does not depend on the inputs.
func Mul32(x, y uint32) (hi, lo uint32) {
	tmp := uint64(x) * uint64(y)
	hi, lo = uint32(tmp>>32), uint32(tmp)
	return
}

// Mul64 returns the 128-bit product of x and y: (hi, lo) = x * y
// with the product bits' upper half returned in hi and the lower
// half returned in lo.
//
// This function's execution time does not depend on the inputs.
func Mul64(x, y uint64) (hi, lo uint64) {
	// This is implemented by custom logic in the compiler, as it requires a 128-bit integer type to be temporarily created.
	return mul64(x, y)
}

//export tinygo.math.mul64
func mul64(x, y uint64) (hi, lo uint64)

// --- Full-width divide ---

// Div returns the quotient and remainder of (hi, lo) divided by y:
// quo = (hi, lo)/y, rem = (hi, lo)%y with the dividend bits' upper
// half in parameter hi and the lower half in parameter lo.
// Div panics for y == 0 (division by zero) or y <= hi (quotient overflow).
func Div(hi, lo, y uint) (quo, rem uint) {
	if UintSize == 32 {
		q, r := Div32(uint32(hi), uint32(lo), uint32(y))
		return uint(q), uint(r)
	}
	q, r := Div64(uint64(hi), uint64(lo), uint64(y))
	return uint(q), uint(r)
}

// Div32 returns the quotient and remainder of (hi, lo) divided by y:
// quo = (hi, lo)/y, rem = (hi, lo)%y with the dividend bits' upper
// half in parameter hi and the lower half in parameter lo.
// Div32 panics for y == 0 (division by zero) or y <= hi (quotient overflow).
func Div32(hi, lo, y uint32) (quo, rem uint32) {
	if y != 0 && y <= hi {
		runtimePanic("integer overflow")
	}
	z := uint64(hi)<<32 | uint64(lo)
	quo, rem = uint32(z/uint64(y)), uint32(z%uint64(y))
	return
}

//go:linkname runtimePanic runtime.runtimePanic
func runtimePanic(str string)

// Div64 returns the quotient and remainder of (hi, lo) divided by y:
// quo = (hi, lo)/y, rem = (hi, lo)%y with the dividend bits' upper
// half in parameter hi and the lower half in parameter lo.
// Div64 panics for y == 0 (division by zero) or y <= hi (quotient overflow).
func Div64(hi, lo, y uint64) (quo, rem uint64) {
	if y == 0 {
		divideByZeroPanic()
	}
	if y <= hi {
		runtimePanic("integer overflow")
	}

	return div64(hi, lo, y)
}

//go:linkname divideByZeroPanic runtime.divideByZeroPanic
func divideByZeroPanic()

// div64 divides a 128-bit integer by a 64-bit integer.
// It computes the remainder and the lower 64 bits of the quotient.
// If y is 0, the result is undefined.
// It is implemented inside the compiler.

//export tinygo.math.div64
func div64(hi, lo, y uint64) (quo, rem uint64)

// Rem returns the remainder of (hi, lo) divided by y. Rem panics for
// y == 0 (division by zero) but, unlike Div, it doesn't panic on a
// quotient overflow.
func Rem(hi, lo, y uint) uint {
	if UintSize == 32 {
		return uint(Rem32(uint32(hi), uint32(lo), uint32(y)))
	}
	return uint(Rem64(uint64(hi), uint64(lo), uint64(y)))
}

// Rem32 returns the remainder of (hi, lo) divided by y. Rem32 panics
// for y == 0 (division by zero) but, unlike Div32, it doesn't panic
// on a quotient overflow.
func Rem32(hi, lo, y uint32) uint32 {
	return uint32((uint64(hi)<<32 | uint64(lo)) % uint64(y))
}

// Rem64 returns the remainder of (hi, lo) divided by y. Rem64 panics
// for y == 0 (division by zero) but, unlike Div64, it doesn't panic
// on a quotient overflow.
func Rem64(hi, lo, y uint64) uint64 {
	if y == 0 {
		divideByZeroPanic()
	}

	_, rem := div64(hi, lo, y)
	return rem
}
