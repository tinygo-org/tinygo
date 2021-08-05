package main

func AsmSqrt(x float64) float64

func AsmAdd(x, y float64) float64

func AsmFoo(x float64) (int64, float64)

func asmExport(x float64) float64 {
	return 0
}

var asmGlobalExport int32
