package transform

import "tinygo.org/x/go-llvm"

// This file implements small transformations on globals (functions and global
// variables) for specific ABIs/architectures.

// ApplyFunctionSections puts every function in a separate section. This makes
// it possible for the linker to remove dead code. It is the equivalent of
// passing -ffunction-sections to a C compiler.
func ApplyFunctionSections(mod llvm.Module) {
	llvmFn := mod.FirstFunction()
	for !llvmFn.IsNil() {
		if !llvmFn.IsDeclaration() && llvmFn.Section() == "" {
			name := llvmFn.Name()
			llvmFn.SetSection(".text." + name)
		}
		llvmFn = llvm.NextFunction(llvmFn)
	}
}
