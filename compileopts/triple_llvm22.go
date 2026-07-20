//go:build llvm22

package compileopts

import "strings"

// ClangTriple returns the target triple to pass to Clang's --target flag.
// LLVM 22 deprecated the "wasm32-unknown-wasi" triple in favor of
// "wasm32-unknown-wasip1", so we substitute it here to avoid warnings.
func ClangTriple(triple string) string {
	return strings.Replace(triple, "wasm32-unknown-wasi", "wasm32-unknown-wasip1", 1)
}
