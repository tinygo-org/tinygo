// +build byollvm

package main

// This file provides a Link() function that uses the bundled lld if possible.

import (
	"errors"
	"os"
	"os/exec"
	"unsafe"
)

/*
#include <stdbool.h>
#include <stdlib.h>
bool tinygo_link_elf(int argc, char **argv);
bool tinygo_link_wasm(int argc, char **argv);
*/
import "C"

// Link invokes a linker with the given name and flags.
//
// This version uses the built-in linker when trying to use lld.
func Link(linker string, flags ...string) error {
	switch linker {
	case "ld.lld":
		flags = append([]string{"tinygo:" + linker}, flags...)
		var cflag *C.char
		buf := C.calloc(C.size_t(len(flags)), C.size_t(unsafe.Sizeof(cflag)))
		cflags := (*[1 << 10]*C.char)(unsafe.Pointer(buf))[:len(flags):len(flags)]
		for i, flag := range flags {
			cflag := C.CString(flag)
			cflags[i] = cflag
			defer C.free(unsafe.Pointer(cflag))
		}
		ok := C.tinygo_link_elf(C.int(len(flags)), (**C.char)(buf))
		if !ok {
			return errors.New("failed to link using built-in ld.lld")
		}
		return nil
	case "wasm-ld":
		flags = append([]string{"tinygo:" + linker}, flags...)
		var cflag *C.char
		buf := C.calloc(C.size_t(len(flags)), C.size_t(unsafe.Sizeof(cflag)))
		defer C.free(buf)
		cflags := (*[1 << 10]*C.char)(unsafe.Pointer(buf))[:len(flags):len(flags)]
		for i, flag := range flags {
			cflag := C.CString(flag)
			cflags[i] = cflag
			defer C.free(unsafe.Pointer(cflag))
		}
		ok := C.tinygo_link_wasm(C.int(len(flags)), (**C.char)(buf))
		if !ok {
			return errors.New("failed to link using built-in wasm-ld")
		}
		return nil
	default:
		// Fall back to external command.
		if cmdNames, ok := commands[linker]; ok {
			return execCommand(cmdNames, flags...)
		}
		cmd := exec.Command(linker, flags...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = sourceDir()
		return cmd.Run()
	}
}
