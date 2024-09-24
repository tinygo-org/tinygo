package builder

// The well-known conservative Boehm-Demers-Weiser GC.
// This file provides a way to compile this GC for use with TinyGo.

import (
	"path/filepath"

	"github.com/tinygo-org/tinygo/goenv"
)

var BoehmGC = Library{
	name: "bdwgc",
	cflags: func(target, headerPath string) []string {
		libdir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/bdwgc")
		return []string{
			// use a modern environment
			"-DUSE_MMAP",              // mmap is available
			"-DUSE_MUNMAP",            // return memory to the OS using munmap
			"-DGC_BUILTIN_ATOMIC",     // use compiler intrinsics for atomic operations
			"-DNO_EXECUTE_PERMISSION", // don't make the heap executable

			// specific flags for TinyGo
			"-DALL_INTERIOR_POINTERS",  // scan interior pointers (needed for Go)
			"-DIGNORE_DYNAMIC_LOADING", // we don't support dynamic loading at the moment
			"-DNO_GETCONTEXT",          // musl doesn't support getcontext()

			// Special flag to work around the lack of __data_start in ld.lld.
			// TODO: try to fix this in LLVM/lld directly so we don't have to
			// work around it anymore.
			"-DGC_DONT_REGISTER_MAIN_STATIC_DATA",

			// Do not scan the stack. We have our own mechanism to do this.
			"-DSTACK_NOT_SCANNED",

			// Assertions can be enabled while debugging GC issues.
			//"-DGC_ASSERTIONS",

			// Threading is not yet supported, so these are disabled.
			//"-DGC_THREADS",
			//"-DTHREAD_LOCAL_ALLOC",

			"-I" + libdir + "/include",
		}
	},
	sourceDir: func() string {
		return filepath.Join(goenv.Get("TINYGOROOT"), "lib/bdwgc")
	},
	librarySources: func(target string) ([]string, error) {
		return []string{
			"allchblk.c",
			"alloc.c",
			"blacklst.c",
			"dbg_mlc.c",
			"dyn_load.c",
			"finalize.c",
			"headers.c",
			"mach_dep.c",
			"malloc.c",
			"mark.c",
			"mark_rts.c",
			"misc.c",
			"new_hblk.c",
			"obj_map.c",
			"os_dep.c",
			"pthread_stop_world.c",
			"pthread_support.c",
			"reclaim.c",
		}, nil
	},
}
