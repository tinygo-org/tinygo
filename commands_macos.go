// +build darwin

package main

// commands used by the compilation process might have different file names on macOS than those used on Linux.
var commands = map[string]string{
	"clang":   "clang-8",
	"ld.lld":  "ld.lld",
	"wasm-ld": "wasm-ld",
}
