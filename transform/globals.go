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
		if !llvmFn.IsDeclaration() {
			name := llvmFn.Name()
			llvmFn.SetSection(".text." + name)
		}
		llvmFn = llvm.NextFunction(llvmFn)
	}
}

// NonConstGlobals turns all global constants into global variables. This works
// around a limitation on Harvard architectures (e.g. AVR), where constant and
// non-constant pointers point to a different address space. Normal pointer
// behavior is restored by using the data space only, at the cost of RAM for
// constant global variables.
func NonConstGlobals(mod llvm.Module) {
	global := mod.FirstGlobal()
	for !global.IsNil() {
		global.SetGlobalConstant(false)
		global = llvm.NextGlobal(global)
	}
}
