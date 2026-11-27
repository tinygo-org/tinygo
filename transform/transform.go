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

import (
	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

// AddStandardAttributes is a helper function to add standard function
// attributes to a function. For example, it adds optsize when requested from
// the -opt= compiler flag.
func AddStandardAttributes(fn llvm.Value, config *compileopts.Config) {
	addFunctionAttributes(fn, standardAttributes(fn.Type().Context(), config))
}

// standardAttributes returns the attributes that AddStandardAttributes adds.
func standardAttributes(ctx llvm.Context, config *compileopts.Config) []llvm.Attribute {
	var attrs []llvm.Attribute
	_, _, sizeLevel := config.OptLevel()
	if sizeLevel >= 1 {
		attrs = append(attrs, ctx.CreateEnumAttribute(llvm.AttributeKindID("optsize"), 0))
	}
	if sizeLevel >= 2 {
		attrs = append(attrs, ctx.CreateEnumAttribute(llvm.AttributeKindID("minsize"), 0))
	}
	if cpu := config.CPU(); cpu != "" {
		attrs = append(attrs, ctx.CreateStringAttribute("target-cpu", cpu))
	}
	if features := config.Features(); features != "" {
		attrs = append(attrs, ctx.CreateStringAttribute("target-features", features))
	}
	return attrs
}

func addFunctionAttributes(fn llvm.Value, attrs []llvm.Attribute) {
	for _, attr := range attrs {
		fn.AddFunctionAttr(attr)
	}
}
