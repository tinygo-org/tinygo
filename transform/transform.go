// Package transform contains transformation passes for the TinyGo compiler.
// These transformation passes may be optimization passes or lowering passes.
//
// Optimization passes transform the IR in such a way that they increase the
// performance of the generated code and/or help the LLVM optimizer better do
// its job by simplifying the IR. This usually means that certain
// TinyGo-specific runtime calls are removed or replaced with something simpler
// if that is a valid operation.
//
// Lowering passes are usually required to run. One example is the interface
// lowering pass, which replaces stub runtime calls to get an interface method
// with the method implementation (either a direct call or a thunk).
package transform
