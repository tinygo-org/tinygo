//go:build llvm22 || llvm23

package compileopts

import "strings"

// ClangTriple returns the target triple to pass to Clang's --target flag.
// LLVM 22 deprecated the "wasm32-unknown-wasi" triple in favor of
// "wasm32-unknown-wasip1", so we substitute it here to avoid warnings.
func ClangTriple(triple string) string {
	if rest, ok := strings.CutSuffix(triple, "wasm32-unknown-wasi"); ok {
		return rest + "wasm32-unknown-wasip1"
	}
	return triple
}
