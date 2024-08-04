package builder

import (
	"os"
	"path/filepath"

	"github.com/tinygo-org/tinygo/goenv"
)

var libWasmBuiltins = Library{
	name: "wasmbuiltins",
	makeHeaders: func(target, includeDir string) error {
		if err := os.Mkdir(includeDir+"/bits", 0o777); err != nil {
			return err
		}
		f, err := os.Create(includeDir + "/bits/alltypes.h")
		if err != nil {
			return err
		}
		if _, err := f.Write([]byte(wasmAllTypes)); err != nil {
			return err
		}
		return f.Close()
	},
	cflags: func(target, headerPath string) []string {
		libcDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/wasi-libc")
		return []string{
			"-Werror",
			"-Wall",
			"-std=gnu11",
			"-nostdlibinc",
			"-isystem", libcDir + "/libc-top-half/musl/arch/wasm32",
			"-isystem", libcDir + "/libc-top-half/musl/arch/generic",
			"-isystem", libcDir + "/libc-top-half/musl/src/internal",
			"-isystem", libcDir + "/libc-top-half/musl/src/include",
			"-isystem", libcDir + "/libc-top-half/musl/include",
			"-isystem", libcDir + "/libc-bottom-half/headers/public",
			"-I" + headerPath,
		}
	},
	sourceDir: func() string { return filepath.Join(goenv.Get("TINYGOROOT"), "lib/wasi-libc") },
	librarySources: func(target string) ([]string, error) {
		return []string{
			// memory builtins needed for llvm.memcpy.*, llvm.memmove.*, and
			// llvm.memset.* LLVM intrinsics.
			"libc-top-half/musl/src/string/memcpy.c",
			"libc-top-half/musl/src/string/memmove.c",
			"libc-top-half/musl/src/string/memset.c",

			// exp, exp2, and log are needed for LLVM math builtin functions
			// like llvm.exp.*.
			"libc-top-half/musl/src/math/__math_divzero.c",
			"libc-top-half/musl/src/math/__math_invalid.c",
			"libc-top-half/musl/src/math/__math_oflow.c",
			"libc-top-half/musl/src/math/__math_uflow.c",
			"libc-top-half/musl/src/math/__math_xflow.c",
			"libc-top-half/musl/src/math/exp.c",
			"libc-top-half/musl/src/math/exp_data.c",
			"libc-top-half/musl/src/math/exp2.c",
			"libc-top-half/musl/src/math/log.c",
			"libc-top-half/musl/src/math/log_data.c",
		}, nil
	},
}

// alltypes.h for wasm-libc, using the types as defined inside Clang.
const wasmAllTypes = `
typedef __SIZE_TYPE__    size_t;
typedef __INT8_TYPE__    int8_t;
typedef __INT16_TYPE__   int16_t;
typedef __INT32_TYPE__   int32_t;
typedef __INT64_TYPE__   int64_t;
typedef __UINT8_TYPE__   uint8_t;
typedef __UINT16_TYPE__  uint16_t;
typedef __UINT32_TYPE__  uint32_t;
typedef __UINT64_TYPE__  uint64_t;
typedef __UINTPTR_TYPE__ uintptr_t;

// This type is used internally in wasi-libc.
typedef double double_t;
`
