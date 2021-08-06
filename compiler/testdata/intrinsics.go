package main

// Test how intrinsics are lowered: either as regular calls to the math
// functions or as LLVM builtins (such as llvm.sqrt.f64).

import "math"

func mySqrt(x float64) float64 {
	return math.Sqrt(x)
}

func myTrunc(x float64) float64 {
	return math.Trunc(x)
}
