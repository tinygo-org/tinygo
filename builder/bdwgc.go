package builder

// The well-known conservative Boehm-Demers-Weiser GC.
// This file provides a way to compile this GC for use with TinyGo.

import (
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

var BoehmGC = Library{
	name: "bdwgc",
	cflags: func(target, headerPath string) []string {
		libdir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/bdwgc")
		flags := []string{
			// use a modern environment
			"-DUSE_MMAP",              // mmap is available
			"-DUSE_MUNMAP",            // return memory to the OS using munmap
			"-DGC_BUILTIN_ATOMIC",     // use compiler intrinsics for atomic operations
			"-DNO_EXECUTE_PERMISSION", // don't make the heap executable

			// specific flags for TinyGo
			"-DALL_INTERIOR_POINTERS",  // scan interior pointers (needed for Go)
			"-DIGNORE_DYNAMIC_LOADING", // we don't support dynamic loading at the moment
			"-DNO_GETCONTEXT",          // musl doesn't support getcontext()
			"-DGC_DISABLE_INCREMENTAL", // don't mess with SIGSEGV and such

			// Use a minimal environment.
			"-DNO_MSGBOX_ON_ERROR", // don't call MessageBoxA on Windows
			"-DDONT_USE_ATEXIT",
			"-DNO_GETENV",          // smaller binary, more predictable configuration
			"-DNO_CLOCK",           // don't use system clock
			"-DNO_DEBUGGING",       // reduce code size
			"-DGC_NO_FINALIZATION", // finalization is not used at the moment

			// Special flag to work around the lack of __data_start in ld.lld.
			// TODO: try to fix this in LLVM/lld directly so we don't have to
			// work around it anymore.
			"-DGC_DONT_REGISTER_MAIN_STATIC_DATA",

			// Do not scan the stack. We have our own mechanism to do this.
			"-DSTACK_NOT_SCANNED",
			"-DNO_PROC_STAT",  // we scan the stack manually (don't read /proc/self/stat on Linux)
			"-DSTACKBOTTOM=0", // dummy value, we scan the stack manually

			// Assertions can be enabled while debugging GC issues.
			//"-DGC_ASSERTIONS",

			// We use our own way of dealing with threads (that is a bit hacky).
			// See src/runtime/gc_boehm.go.
			//"-DGC_THREADS",
			//"-DTHREAD_LOCAL_ALLOC",

			"-I" + libdir + "/include",
		}
		return flags
	},
	needsLibc: true,
	sourceDir: func() string {
		return filepath.Join(goenv.Get("TINYGOROOT"), "lib/bdwgc")
	},
	librarySources: func(target string, _ bool) ([]string, error) {
		sources := []string{
			"allchblk.c",
			"alloc.c",
			"blacklst.c",
			"dbg_mlc.c",
			"dyn_load.c",
			"headers.c",
			"mach_dep.c",
			"malloc.c",
			"mark.c",
			"mark_rts.c",
			"misc.c",
			"new_hblk.c",
			"os_dep.c",
			"reclaim.c",
		}
		if strings.Split(target, "-")[2] == "windows" {
			// Due to how the linker on Windows works (that doesn't allow
			// undefined functions), we need to include these extra files.
			sources = append(sources,
				"mallocx.c",
				"ptr_chck.c",
			)
		}
		return sources, nil
	},
}
