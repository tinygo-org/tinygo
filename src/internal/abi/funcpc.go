package abi

// These two signatures are present to satisfy the expectation of some programs
// (in particular internal/syscall/unix on MacOS). They do not currently have an
// implementation, in part because TinyGo doesn't use ABI0 or ABIInternal (it
// uses a C-like calling convention).
// Calls to FuncPCABI0 however are treated specially by the compiler when
// compiling for MacOS.

func FuncPCABI0(f interface{}) uintptr

func FuncPCABIInternal(f interface{}) uintptr
