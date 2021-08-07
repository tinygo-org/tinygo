// +build byollvm

package builder

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"unsafe"
)

/*
#cgo CXXFLAGS: -fno-rtti
#cgo LDFLAGS: -lbinaryen
#include <stdbool.h>
#include <stdlib.h>
#include <binaryen-c.h>
bool tinygo_clang_driver(int argc, char **argv);
bool tinygo_link_elf(int argc, char **argv);
bool tinygo_link_wasm(int argc, char **argv);
*/
import "C"

const hasBuiltinTools = true

// RunTool runs the given tool (such as clang).
//
// This version actually runs the tools because TinyGo was compiled while
// linking statically with LLVM (with the byollvm build tag).
func RunTool(tool string, args ...string) error {
	args = append([]string{"tinygo:" + tool}, args...)

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
	case "ld.lld":
		ok = C.tinygo_link_elf(C.int(len(args)), (**C.char)(buf))
	case "wasm-ld":
		ok = C.tinygo_link_wasm(C.int(len(args)), (**C.char)(buf))
	default:
		return errors.New("unknown tool: " + tool)
	}
	if !ok {
		return errors.New("failed to run tool: " + tool)
	}
	return nil
}

// wasmOptLock exists because the binaryen C API uses global state.
// In the future, we can hopefully avoid this by re-execing.
var wasmOptLock sync.Mutex

func WasmOpt(src, dst string, cfg BinaryenConfig) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading source file: %w", err)
	}

	data, err = wasmOptApply(data, cfg)
	if err != nil {
		return fmt.Errorf("running wasm-opt: %w", err)
	}

	err = os.WriteFile(dst, data, 0666)
	if err != nil {
		return fmt.Errorf("writing destination file: %w", err)
	}

	return nil
}

func wasmOptApply(src []byte, cfg BinaryenConfig) ([]byte, error) {
	wasmOptLock.Lock()
	defer wasmOptLock.Unlock()

	// This doesn't actually seem to do anything?
	C.BinaryenSetDebugInfo(false)

	mod := C.BinaryenModuleRead((*C.char)(unsafe.Pointer(&src[0])), C.size_t(len(src)))
	defer C.BinaryenModuleDispose(mod)

	C.BinaryenSetOptimizeLevel(C.int(cfg.OptLevel))
	C.BinaryenSetShrinkLevel(C.int(cfg.ShrinkLevel))
	C.BinaryenModuleOptimize(mod)
	C.BinaryenModuleValidate(mod)

	// TODO: include source map
	res := C.BinaryenModuleAllocateAndWrite(mod, nil)
	defer C.free(unsafe.Pointer(res.binary))
	if res.sourceMap != nil {
		defer C.free(unsafe.Pointer(res.sourceMap))
	}

	return C.GoBytes(res.binary, C.int(res.binaryBytes)), nil
}
