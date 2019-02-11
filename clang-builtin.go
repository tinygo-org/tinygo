// +build byollvm

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"unsafe"
)

/*
#cgo CXXFLAGS: -fno-rtti
#include <stdbool.h>
#include <stdlib.h>
bool tinygo_libclang_driver(int argc, char **argv, char *resourcesDir);
*/
import "C"

// CCompiler invokes a C compiler with the given arguments.
//
// This version uses the builtin libclang when trying to run clang.
func CCompiler(dir, command, path string, flags ...string) error {
	switch command {
	case "clang", commands["clang"]:
		// Compile this with the internal Clang compiler.
		if !filepath.IsAbs(path) {
			path = filepath.Join(dir, path)
		}
		flags = append([]string{"clang", path}, flags...)
		var cflag *C.char
		buf := C.calloc(C.size_t(len(flags)), C.size_t(unsafe.Sizeof(cflag)))
		cflags := (*[1 << 10]*C.char)(unsafe.Pointer(buf))[:len(flags):len(flags)]
		for i, flag := range flags {
			cflag := C.CString(flag)
			cflags[i] = cflag
			defer C.free(unsafe.Pointer(cflag))
		}
		resourcesDir := C.CString(filepath.Join(dir, "clang"))
		defer C.free(unsafe.Pointer(resourcesDir))
		ok := C.tinygo_libclang_driver(C.int(len(flags)), (**C.char)(buf), resourcesDir)
		if !ok {
			return errors.New("failed to compile using built-in clang")
		}
		return nil
	default:
		// Fall back to external command.
		cmd := exec.Command(command, append(flags, path)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = dir
		return cmd.Run()
	}
}
