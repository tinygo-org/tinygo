package llvmutil

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestRemoveGlobalReferences(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("test")
	defer mod.Dispose()

	fnType := llvm.FunctionType(ctx.VoidType(), nil, false)
	kept := llvm.AddFunction(mod, "kept", fnType)
	kept.SetLinkage(llvm.InternalLinkage)
	shared := llvm.AddFunction(mod, "shared", fnType)
	shared.SetLinkage(llvm.InternalLinkage)
	temporary := llvm.AddFunction(mod, "temporary", fnType)
	temporary.SetLinkage(llvm.InternalLinkage)
	AppendToGlobal(mod, "llvm.used", kept, shared, shared, temporary)
	AppendToGlobal(mod, "temporary.roots", shared, temporary)

	RemoveGlobalReferences(mod, "llvm.used", "temporary.roots")
	options := llvm.NewPassBuilderOptions()
	defer options.Dispose()
	if err := mod.RunPasses("globaldce", llvm.TargetMachine{}, options); err != nil {
		t.Fatal(err)
	}

	if mod.NamedFunction("kept").IsNil() {
		t.Error("permanent root was removed")
	}
	if mod.NamedFunction("shared").IsNil() {
		t.Error("permanent root was removed")
	}
	if !mod.NamedFunction("temporary").IsNil() {
		t.Error("temporary root was retained")
	}
}
