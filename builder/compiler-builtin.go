// +build byollvm

package builder

import (
	"errors"
	"os"
	"os/exec"
	"unsafe"

	"github.com/tinygo-org/tinygo/goenv"
)

/*
#cgo CXXFLAGS: -fno-rtti
#include <stdbool.h>
#include <stdlib.h>
bool tinygo_clang_driver(int argc, char **argv);
*/
import "C"

// runCCompiler invokes a C compiler with the given arguments.
//
// This version invokes the built-in Clang when trying to run the Clang compiler.
func runCCompiler(command string, flags ...string) error {
	switch command {
	case "clang":
		// Compile this with the internal Clang compiler.
		headerPath := getClangHeaderPath(goenv.Get("TINYGOROOT"))
		if headerPath == "" {
			return errors.New("could not locate Clang headers")
		}
		flags = append(flags, "-I"+headerPath)
		flags = append([]string{"tinygo:" + command}, flags...)
		var cflag *C.char
		buf := C.calloc(C.size_t(len(flags)), C.size_t(unsafe.Sizeof(cflag)))
		cflags := (*[1 << 10]*C.char)(unsafe.Pointer(buf))[:len(flags):len(flags)]
		for i, flag := range flags {
			cflag := C.CString(flag)
			cflags[i] = cflag
			defer C.free(unsafe.Pointer(cflag))
		}
		ok := C.tinygo_clang_driver(C.int(len(flags)), (**C.char)(buf))
		if !ok {
			return errors.New("failed to compile using built-in clang")
		}
		return nil
	default:
		// Running some other compiler. Maybe it has been defined in the
		// commands map (unlikely).
		if cmdNames, ok := commands[command]; ok {
			return execCommand(cmdNames, flags...)
		}
		// Alternatively, run the compiler directly.
		cmd := exec.Command(command, flags...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}
