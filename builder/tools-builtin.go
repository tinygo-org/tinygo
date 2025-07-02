//go:build byollvm

package builder

import (
	"errors"
	"unsafe"
)

/*
#cgo CXXFLAGS: -fno-rtti
#include <stdbool.h>
#include <stdlib.h>
bool tinygo_clang_driver(int argc, char **argv);
bool tinygo_link(int argc, char **argv);
*/
import "C"

const hasBuiltinTools = true

// RunTool runs the given tool (such as clang).
//
// This version actually runs the tools because TinyGo was compiled while
// linking statically with LLVM (with the byollvm build tag).
func RunTool(tool string, args ...string) error {
	args = append([]string{tool}, args...)

	var cflag *C.char
	buf := C.calloc(C.size_t(len(args)), C.size_t(unsafe.Sizeof(cflag)))
	defer C.free(buf)
	cflags := (*[1 << 10]*C.char)(unsafe.Pointer(buf))[:len(args):len(args)]
	for i, flag := range args {
		cflag := C.CString(flag)
		cflags[i] = cflag
		defer C.free(unsafe.Pointer(cflag))
	}

	var ok C.bool
	switch tool {
	case "clang":
		ok = C.tinygo_clang_driver(C.int(len(args)), (**C.char)(buf))
	case "ld.lld", "wasm-ld":
		ok = C.tinygo_link(C.int(len(args)), (**C.char)(buf))
	default:
		return errors.New("unknown tool: " + tool)
	}
	if !ok {
		return errors.New("failed to run tool: " + tool)
	}
	return nil
}
