// +build !darwin

package main

// commands used by the compilation process might have different file names on Linux than those used on macOS.
var commands = map[string]string{
	"ar":      "llvm-ar-8",
	"clang":   "clang-8",
	"ld.lld":  "ld.lld-8",
	"wasm-ld": "wasm-ld-8",
}
