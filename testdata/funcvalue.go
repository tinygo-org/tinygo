package main

import "unsafe"

type FuncNonExport func(x int) int
type FuncExport func(_exportABI struct{}, x int) int

func privateDouble(x int) int {
	return x * 2
}

//go:export PublicSquare
func PublicSquare(_exportABI struct{}, x int) int {
	return x * x
}

type NonExportFuncInternalRepr struct {
	closureContext     unsafe.Pointer
	functionEntryPoint unsafe.Pointer
}

type ExportFuncInternalRepr struct {
	functionEntryPoint unsafe.Pointer
}

func main() {
	// x is mutated to stop LLVM from just calculating everything at the compile time
	x := 1
	println(privateDouble(x))
	x++
	println(PublicSquare(struct{}{}, x))
	var f FuncNonExport = privateDouble
	x++
	println(f(x))
	var f2 FuncExport = PublicSquare
	x++
	println(f2(struct{}{}, x))

	// Demo: how to obtain actual function pointers
	var fptr1 unsafe.Pointer
	var fptr2 unsafe.Pointer
	fptr1 = (*NonExportFuncInternalRepr)(unsafe.Pointer(&f)).functionEntryPoint // Warning, the raw function params are actually (x int, closureContext unsafePtr)
	fptr2 = (*ExportFuncInternalRepr)(unsafe.Pointer(&f)).functionEntryPoint    // The raw function params is simply (x int)
	if fptr1 == fptr2 {
		println("This should not happen")
	}

	if unsafe.Sizeof(f) != unsafe.Sizeof(fptr1)*2 {
		panic("Non-exported function should be represented as 2 pointers")
	}
	if unsafe.Sizeof(f2) != unsafe.Sizeof(fptr1) {
		panic("Non-exported function should be represented as 1 pointer")
	}
}
