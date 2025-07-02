package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

var libWasiLibc = Library{
	name: "wasi-libc",
	makeHeaders: func(target, includeDir string) error {
		bits := filepath.Join(includeDir, "bits")
		err := os.Mkdir(bits, 0777)
		if err != nil {
			return err
		}

		muslDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib", "wasi-libc/libc-top-half/musl")
		err = buildMuslAllTypes("wasm32", muslDir, bits)
		if err != nil {
			return err
		}

		// See MUSL_OMIT_HEADERS in the Makefile.
		omitHeaders := map[string]struct{}{
			"syslog.h":   {},
			"wait.h":     {},
			"ucontext.h": {},
			"paths.h":    {},
			"utmp.h":     {},
			"utmpx.h":    {},
			"lastlog.h":  {},
			"elf.h":      {},
			"link.h":     {},
			"pwd.h":      {},
			"shadow.h":   {},
			"grp.h":      {},
			"mntent.h":   {},
			"netdb.h":    {},
			"resolv.h":   {},
			"pty.h":      {},
			"dlfcn.h":    {},
			"setjmp.h":   {},
			"ulimit.h":   {},
			"wordexp.h":  {},
			"spawn.h":    {},
			"termios.h":  {},
			"libintl.h":  {},
			"aio.h":      {},

			"stdarg.h": {},
			"stddef.h": {},

			"pthread.h": {},
		}

		for _, glob := range [][2]string{
			{"libc-bottom-half/headers/public/*.h", ""},
			{"libc-bottom-half/headers/public/wasi/*.h", "wasi"},
			{"libc-top-half/musl/arch/wasm32/bits/*.h", "bits"},
			{"libc-top-half/musl/include/*.h", ""},
			{"libc-top-half/musl/include/netinet/*.h", "netinet"},
			{"libc-top-half/musl/include/sys/*.h", "sys"},
		} {
			matches, _ := filepath.Glob(filepath.Join(goenv.Get("TINYGOROOT"), "lib/wasi-libc", glob[0]))
			outDir := filepath.Join(includeDir, glob[1])
			os.MkdirAll(outDir, 0o777)
			for _, match := range matches {
				name := filepath.Base(match)
				if _, ok := omitHeaders[name]; ok {
					continue
				}
				data, err := os.ReadFile(match)
				if err != nil {
					return err
				}
				err = os.WriteFile(filepath.Join(outDir, name), data, 0o666)
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
	cflags: func(target, headerPath string) []string {
		libcDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/wasi-libc")
		return []string{
			"-Werror",
			"-Wall",
			"-std=gnu11",
			"-nostdlibinc",
			"-mnontrapping-fptoint", "-msign-ext", "-mbulk-memory",
			"-Wno-null-pointer-arithmetic", "-Wno-unused-parameter", "-Wno-sign-compare", "-Wno-unused-variable", "-Wno-unused-function", "-Wno-ignored-attributes", "-Wno-missing-braces", "-Wno-ignored-pragmas", "-Wno-unused-but-set-variable", "-Wno-unknown-warning-option",
			"-Wno-parentheses", "-Wno-shift-op-parentheses", "-Wno-bitwise-op-parentheses", "-Wno-logical-op-parentheses", "-Wno-string-plus-int", "-Wno-dangling-else", "-Wno-unknown-pragmas",
			"-DNDEBUG",
			"-D__wasilibc_printscan_no_long_double",
			"-D__wasilibc_printscan_full_support_option=\"long double support is disabled\"",
			"-DBULK_MEMORY_THRESHOLD=32", // default threshold in wasi-libc
			"-isystem", headerPath,
			"-I" + libcDir + "/libc-top-half/musl/src/include",
			"-I" + libcDir + "/libc-top-half/musl/src/internal",
			"-I" + libcDir + "/libc-top-half/musl/arch/wasm32",
			"-I" + libcDir + "/libc-top-half/musl/arch/generic",
			"-I" + libcDir + "/libc-top-half/headers/private",
		}
	},
	cflagsForFile: func(path string) []string {
		if strings.HasPrefix(path, "libc-bottom-half"+string(os.PathSeparator)) {
			libcDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/wasi-libc")
			return []string{
				"-I" + libcDir + "/libc-bottom-half/headers/private",
				"-I" + libcDir + "/libc-bottom-half/cloudlibc/src/include",
				"-I" + libcDir + "/libc-bottom-half/cloudlibc/src",
			}
		}
		return nil
	},
	sourceDir: func() string { return filepath.Join(goenv.Get("TINYGOROOT"), "lib/wasi-libc") },
	librarySources: func(target string, libcNeedsMalloc bool) ([]string, error) {
		type filePattern struct {
			glob    string
			exclude []string
		}

		// See: LIBC_TOP_HALF_MUSL_SOURCES in the Makefile
		globs := []filePattern{
			// Top half: mostly musl sources.
			{glob: "libc-top-half/sources/*.c"},
			{glob: "libc-top-half/musl/src/conf/*.c"},
			{glob: "libc-top-half/musl/src/internal/*.c", exclude: []string{
				"procfdname.c", "syscall.c", "syscall_ret.c", "vdso.c", "version.c",
			}},
			{glob: "libc-top-half/musl/src/locale/*.c", exclude: []string{
				"dcngettext.c", "textdomain.c", "bind_textdomain_codeset.c"}},
			{glob: "libc-top-half/musl/src/math/*.c", exclude: []string{
				"__signbit.c", "__signbitf.c", "__signbitl.c",
				"__fpclassify.c", "__fpclassifyf.c", "__fpclassifyl.c",
				"ceilf.c", "ceil.c",
				"floorf.c", "floor.c",
				"truncf.c", "trunc.c",
				"rintf.c", "rint.c",
				"nearbyintf.c", "nearbyint.c",
				"sqrtf.c", "sqrt.c",
				"fabsf.c", "fabs.c",
				"copysignf.c", "copysign.c",
				"fminf.c", "fmaxf.c",
				"fmin.c", "fmax.c,",
			}},
			{glob: "libc-top-half/musl/src/multibyte/*.c"},
			{glob: "libc-top-half/musl/src/stdio/*.c", exclude: []string{
				"vfwscanf.c", "vfwprintf.c", // long double is unsupported
				"__lockfile.c", "flockfile.c", "funlockfile.c", "ftrylockfile.c",
				"rename.c",
				"tmpnam.c", "tmpfile.c", "tempnam.c",
				"popen.c", "pclose.c",
				"remove.c",
				"gets.c"}},
			{glob: "libc-top-half/musl/src/stdlib/*.c"},
			{glob: "libc-top-half/musl/src/string/*.c", exclude: []string{
				"strsignal.c"}},

			// Bottom half: connect top half to WASI equivalents.
			{glob: "libc-bottom-half/cloudlibc/src/libc/*/*.c"},
			{glob: "libc-bottom-half/cloudlibc/src/libc/sys/*/*.c"},
			{glob: "libc-bottom-half/sources/*.c"},
		}

		// We're using the Boehm GC, so we need a heap implementation in the libc.
		if libcNeedsMalloc {
			globs = append(globs, filePattern{glob: "dlmalloc/src/dlmalloc.c"})
		}

		// See: LIBC_TOP_HALF_MUSL_SOURCES in the Makefile
		sources := []string{
			"libc-top-half/musl/src/misc/a64l.c",
			"libc-top-half/musl/src/misc/basename.c",
			"libc-top-half/musl/src/misc/dirname.c",
			"libc-top-half/musl/src/misc/ffs.c",
			"libc-top-half/musl/src/misc/ffsl.c",
			"libc-top-half/musl/src/misc/ffsll.c",
			"libc-top-half/musl/src/misc/fmtmsg.c",
			"libc-top-half/musl/src/misc/getdomainname.c",
			"libc-top-half/musl/src/misc/gethostid.c",
			"libc-top-half/musl/src/misc/getopt.c",
			"libc-top-half/musl/src/misc/getopt_long.c",
			"libc-top-half/musl/src/misc/getsubopt.c",
			"libc-top-half/musl/src/misc/uname.c",
			"libc-top-half/musl/src/misc/nftw.c",
			"libc-top-half/musl/src/errno/strerror.c",
			"libc-top-half/musl/src/network/htonl.c",
			"libc-top-half/musl/src/network/htons.c",
			"libc-top-half/musl/src/network/ntohl.c",
			"libc-top-half/musl/src/network/ntohs.c",
			"libc-top-half/musl/src/network/inet_ntop.c",
			"libc-top-half/musl/src/network/inet_pton.c",
			"libc-top-half/musl/src/network/inet_aton.c",
			"libc-top-half/musl/src/network/in6addr_any.c",
			"libc-top-half/musl/src/network/in6addr_loopback.c",
			"libc-top-half/musl/src/fenv/fenv.c",
			"libc-top-half/musl/src/fenv/fesetround.c",
			"libc-top-half/musl/src/fenv/feupdateenv.c",
			"libc-top-half/musl/src/fenv/fesetexceptflag.c",
			"libc-top-half/musl/src/fenv/fegetexceptflag.c",
			"libc-top-half/musl/src/fenv/feholdexcept.c",
			"libc-top-half/musl/src/exit/exit.c",
			"libc-top-half/musl/src/exit/atexit.c",
			"libc-top-half/musl/src/exit/assert.c",
			"libc-top-half/musl/src/exit/quick_exit.c",
			"libc-top-half/musl/src/exit/at_quick_exit.c",
			"libc-top-half/musl/src/time/strftime.c",
			"libc-top-half/musl/src/time/asctime.c",
			"libc-top-half/musl/src/time/asctime_r.c",
			"libc-top-half/musl/src/time/ctime.c",
			"libc-top-half/musl/src/time/ctime_r.c",
			"libc-top-half/musl/src/time/wcsftime.c",
			"libc-top-half/musl/src/time/strptime.c",
			"libc-top-half/musl/src/time/difftime.c",
			"libc-top-half/musl/src/time/timegm.c",
			"libc-top-half/musl/src/time/ftime.c",
			"libc-top-half/musl/src/time/gmtime.c",
			"libc-top-half/musl/src/time/gmtime_r.c",
			"libc-top-half/musl/src/time/timespec_get.c",
			"libc-top-half/musl/src/time/getdate.c",
			"libc-top-half/musl/src/time/localtime.c",
			"libc-top-half/musl/src/time/localtime_r.c",
			"libc-top-half/musl/src/time/mktime.c",
			"libc-top-half/musl/src/time/__tm_to_secs.c",
			"libc-top-half/musl/src/time/__month_to_secs.c",
			"libc-top-half/musl/src/time/__secs_to_tm.c",
			"libc-top-half/musl/src/time/__year_to_secs.c",
			"libc-top-half/musl/src/time/__tz.c",
			"libc-top-half/musl/src/fcntl/creat.c",
			"libc-top-half/musl/src/dirent/alphasort.c",
			"libc-top-half/musl/src/dirent/versionsort.c",
			"libc-top-half/musl/src/env/__stack_chk_fail.c",
			"libc-top-half/musl/src/env/clearenv.c",
			"libc-top-half/musl/src/env/getenv.c",
			"libc-top-half/musl/src/env/putenv.c",
			"libc-top-half/musl/src/env/setenv.c",
			"libc-top-half/musl/src/env/unsetenv.c",
			"libc-top-half/musl/src/unistd/posix_close.c",
			"libc-top-half/musl/src/stat/futimesat.c",
			"libc-top-half/musl/src/legacy/getpagesize.c",
			"libc-top-half/musl/src/thread/thrd_sleep.c",
		}

		basepath := goenv.Get("TINYGOROOT") + "/lib/wasi-libc/"
		for _, pattern := range globs {
			matches, err := filepath.Glob(basepath + pattern.glob)
			if err != nil {
				// From the documentation:
				// > Glob ignores file system errors such as I/O errors reading
				// > directories. The only possible returned error is
				// > ErrBadPattern, when pattern is malformed.
				// So the only possible error is when the (statically defined)
				// pattern is wrong. In other words, a programming bug.
				return nil, fmt.Errorf("wasi-libc: could not glob source dirs: %w", err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("wasi-libc: did not find any files for pattern %#v", pattern)
			}
			excludeSet := map[string]struct{}{}
			for _, exclude := range pattern.exclude {
				excludeSet[exclude] = struct{}{}
			}
			for _, match := range matches {
				if _, ok := excludeSet[filepath.Base(match)]; ok {
					continue
				}
				relpath, err := filepath.Rel(basepath, match)
				if err != nil {
					// Not sure if this is even possible.
					return nil, err
				}
				sources = append(sources, relpath)
			}
		}
		return sources, nil
	},
}
