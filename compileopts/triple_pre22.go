//go:build !llvm22

package compileopts

// ClangTriple returns the target triple to pass to Clang's --target flag.
// For pre-LLVM 22, the triple is used as-is.
func ClangTriple(triple string) string {
	return triple
}
