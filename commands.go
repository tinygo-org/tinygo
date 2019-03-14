// +build !darwin

package main

// commands used by the compilation process might have different file names on Linux than those used on macOS.
var commands = map[string]string{
	"ar":      "llvm-ar-7",
	"clang":   "clang-7",
	"ld.lld":  "ld.lld-7",
	"wasm-ld": "wasm-ld-7",
}
