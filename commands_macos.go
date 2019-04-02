// +build darwin

package main

// commands used by the compilation process might have different file names on macOS than those used on Linux.
var commands = map[string]string{
	"ar":      "llvm-ar",
	"clang":   "clang-8",
	"ld.lld":  "ld.lld-8",
	"wasm-ld": "wasm-ld",
}
