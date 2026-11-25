//go:build !llvm22 && !llvm23

package compileopts

// LLVMTriple returns the target triple for Clang and for LLVM modules.
// Before LLVM 22, the triple is used as is.
func LLVMTriple(triple string) string {
	return triple
}
