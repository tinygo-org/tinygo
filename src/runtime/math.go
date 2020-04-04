package runtime

// This file redirects math stubs to their fallback implementation.
// TODO: use optimized versions if possible.

import (
	_ "unsafe"
)

//go:linkname math_Asin math.Asin
func math_Asin(x float64) float64 { return math_asin(x) }

//go:linkname math_asin math.asin
func math_asin(x float64) float64

//go:linkname math_Asinh math.Asinh
func math_Asinh(x float64) float64 { return math_asinh(x) }

//go:linkname math_asinh math.asinh
func math_asinh(x float64) float64

//go:linkname math_Acos math.Acos
func math_Acos(x float64) float64 { return math_acos(x) }

//go:linkname math_acos math.acos
func math_acos(x float64) float64

//go:linkname math_Acosh math.Acosh
func math_Acosh(x float64) float64 { return math_acosh(x) }

//go:linkname math_acosh math.acosh
func math_acosh(x float64) float64

//go:linkname math_Atan math.Atan
func math_Atan(x float64) float64 { return math_atan(x) }

//go:linkname math_atan math.atan
func math_atan(x float64) float64

//go:linkname math_Atanh math.Atanh
func math_Atanh(x float64) float64 { return math_atanh(x) }

//go:linkname math_atanh math.atanh
func math_atanh(x float64) float64

//go:linkname math_Atan2 math.Atan2
func math_Atan2(y, x float64) float64 { return math_atan2(y, x) }

//go:linkname math_atan2 math.atan2
func math_atan2(y, x float64) float64

//go:linkname math_Cbrt math.Cbrt
func math_Cbrt(x float64) float64 { return math_cbrt(x) }

//go:linkname math_cbrt math.cbrt
func math_cbrt(x float64) float64

//go:linkname math_Ceil math.Ceil
func math_Ceil(x float64) float64 {
	if GOARCH == "arm64" || GOARCH == "wasm" {
		return llvm_ceil(x)
	}
	return math_ceil(x)
}

//export llvm.ceil.f64
func llvm_ceil(x float64) float64

//go:linkname math_ceil math.ceil
func math_ceil(x float64) float64

//go:linkname math_Cos math.Cos
func math_Cos(x float64) float64 { return math_cos(x) }

//go:linkname math_cos math.cos
func math_cos(x float64) float64

//go:linkname math_Cosh math.Cosh
func math_Cosh(x float64) float64 { return math_cosh(x) }

//go:linkname math_cosh math.cosh
func math_cosh(x float64) float64

//go:linkname math_Erf math.Erf
func math_Erf(x float64) float64 { return math_erf(x) }

//go:linkname math_erf math.erf
func math_erf(x float64) float64

//go:linkname math_Erfc math.Erfc
func math_Erfc(x float64) float64 { return math_erfc(x) }

//go:linkname math_erfc math.erfc
func math_erfc(x float64) float64

//go:linkname math_Exp math.Exp
func math_Exp(x float64) float64 { return math_exp(x) }

//go:linkname math_exp math.exp
func math_exp(x float64) float64

//go:linkname math_Expm1 math.Expm1
func math_Expm1(x float64) float64 { return math_expm1(x) }

//go:linkname math_expm1 math.expm1
func math_expm1(x float64) float64

//go:linkname math_Exp2 math.Exp2
func math_Exp2(x float64) float64 { return math_exp2(x) }

//go:linkname math_exp2 math.exp2
func math_exp2(x float64) float64

//go:linkname math_Floor math.Floor
func math_Floor(x float64) float64 {
	if GOARCH == "arm64" || GOARCH == "wasm" {
		return llvm_floor(x)
	}
	return math_floor(x)
}

//export llvm.floor.f64
func llvm_floor(x float64) float64

//go:linkname math_floor math.floor
func math_floor(x float64) float64

//go:linkname math_Frexp math.Frexp
func math_Frexp(x float64) (float64, int) { return math_frexp(x) }

//go:linkname math_frexp math.frexp
func math_frexp(x float64) (float64, int)

//go:linkname math_Hypot math.Hypot
func math_Hypot(p, q float64) float64 { return math_hypot(p, q) }

//go:linkname math_hypot math.hypot
func math_hypot(p, q float64) float64

//go:linkname math_Ldexp math.Ldexp
func math_Ldexp(frac float64, exp int) float64 { return math_ldexp(frac, exp) }

//go:linkname math_ldexp math.ldexp
func math_ldexp(frac float64, exp int) float64

//go:linkname math_Log math.Log
func math_Log(x float64) float64 { return math_log(x) }

//go:linkname math_log math.log
func math_log(x float64) float64

//go:linkname math_Log1p math.Log1p
func math_Log1p(x float64) float64 { return math_log1p(x) }

//go:linkname math_log1p math.log1p
func math_log1p(x float64) float64

//go:linkname math_Log10 math.Log10
func math_Log10(x float64) float64 { return math_log10(x) }

//go:linkname math_log10 math.log10
func math_log10(x float64) float64

//go:linkname math_Log2 math.Log2
func math_Log2(x float64) float64 { return math_log2(x) }

//go:linkname math_log2 math.log2
func math_log2(x float64) float64

//go:linkname math_Max math.Max
func math_Max(x, y float64) float64 {
	if GOARCH == "arm64" || GOARCH == "wasm" {
		return llvm_maximum(x, y)
	}
	return math_max(x, y)
}

//export llvm.maximum.f64
func llvm_maximum(x, y float64) float64

//go:linkname math_max math.max
func math_max(x, y float64) float64

//go:linkname math_Min math.Min
func math_Min(x, y float64) float64 {
	if GOARCH == "arm64" || GOARCH == "wasm" {
		return llvm_minimum(x, y)
	}
	return math_min(x, y)
}

//export llvm.minimum.f64
func llvm_minimum(x, y float64) float64

//go:linkname math_min math.min
func math_min(x, y float64) float64

//go:linkname math_Mod math.Mod
func math_Mod(x, y float64) float64 { return math_mod(x, y) }

//go:linkname math_mod math.mod
func math_mod(x, y float64) float64

//go:linkname math_Modf math.Modf
func math_Modf(x float64) (float64, float64) { return math_modf(x) }

//go:linkname math_modf math.modf
func math_modf(x float64) (float64, float64)

//go:linkname math_Pow math.Pow
func math_Pow(x, y float64) float64 { return math_pow(x, y) }

//go:linkname math_pow math.pow
func math_pow(x, y float64) float64

//go:linkname math_Remainder math.Remainder
func math_Remainder(x, y float64) float64 { return math_remainder(x, y) }

//go:linkname math_remainder math.remainder
func math_remainder(x, y float64) float64

//go:linkname math_Sin math.Sin
func math_Sin(x float64) float64 { return math_sin(x) }

//go:linkname math_sin math.sin
func math_sin(x float64) float64

//go:linkname math_Sinh math.Sinh
func math_Sinh(x float64) float64 { return math_sinh(x) }

//go:linkname math_sinh math.sinh
func math_sinh(x float64) float64

//go:linkname math_Sqrt math.Sqrt
func math_Sqrt(x float64) float64 {
	if GOARCH == "x86" || GOARCH == "amd64" || GOARCH == "wasm" {
		return llvm_sqrt(x)
	}
	return math_sqrt(x)
}

//export llvm.sqrt.f64
func llvm_sqrt(x float64) float64

//go:linkname math_sqrt math.sqrt
func math_sqrt(x float64) float64

//go:linkname math_Tan math.Tan
func math_Tan(x float64) float64 { return math_tan(x) }

//go:linkname math_tan math.tan
func math_tan(x float64) float64

//go:linkname math_Tanh math.Tanh
func math_Tanh(x float64) float64 { return math_tanh(x) }

//go:linkname math_tanh math.tanh
func math_tanh(x float64) float64

//go:linkname math_Trunc math.Trunc
func math_Trunc(x float64) float64 {
	if GOARCH == "arm64" || GOARCH == "wasm" {
		return llvm_trunc(x)
	}
	return math_trunc(x)
}

//export llvm.trunc.f64
func llvm_trunc(x float64) float64

//go:linkname math_trunc math.trunc
func math_trunc(x float64) float64
