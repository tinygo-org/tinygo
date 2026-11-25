//go:build llvm22 || llvm23

package compileopts

// LLVMTriple returns the target triple for Clang and for LLVM modules.
// LLVM 22 deprecated "wasm32-unknown-wasi" in favor of "wasm32-unknown-wasip1".
func LLVMTriple(triple string) string {
	if triple == "wasm32-unknown-wasi" {
		return "wasm32-unknown-wasip1"
	}
	return triple
}
