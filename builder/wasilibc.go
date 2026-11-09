package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

// generateWasiVersionHeader makes wasi/version.h from version.h.in.
// Upstream makes this file with CMake configure_file.
func generateWasiVersionHeader(includeDir, preview string) error {
	libcDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/wasi-libc")
	versionHeaderIn := filepath.Join(libcDir, "libc-bottom-half/headers/public/wasi/version.h.in")
	data, err := os.ReadFile(versionHeaderIn)
	if err != nil {
		return err
	}
	define := "__wasip" + preview + "__"
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "#cmakedefine ") {
			continue
		}
		name := strings.Fields(line)[1]
		// wasip3 is built with cooperative threads.
		if name == define || (name == "__wasi_cooperative_threads__" && preview == "3") {
			lines[i] = "#define " + name
		} else {
			lines[i] = "/* #undef " + name + " */"
		}
	}
	return os.WriteFile(filepath.Join(includeDir, "wasi", "version.h"), []byte(strings.Join(lines, "\n")), 0o666)
}

// libWasiLibc, libWasiLibcWasip2 and libWasiLibcWasip3 use different library names.
// The build cache uses the name and the triple.
var libWasiLibc = newWasiLibc("1", "wasi-libc")
var libWasiLibcWasip2 = newWasiLibc("2", "wasi-libc-wasip2")
var libWasiLibcWasip3 = newWasiLibc("3", "wasi-libc-wasip3")

func newWasiLibc(preview, name string) Library {
	return Library{
		name: name,
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
			if preview == "3" {
				// Cooperative threads are supported.
				delete(omitHeaders, "pthread.h")
			}
			if preview == "1" {
				// netdb.h is not implemented for wasip1.
				omitHeaders["netdb.h"] = struct{}{}
			}

			// Generic bits headers that aren't supported on WASI or that are
			// replaced by WASI specific ones. See libc-top-half/CMakeLists.txt.
			omitGenericBits := map[string]struct{}{
				"errno.h": {}, "dirent.h": {}, "fcntl.h": {}, "ioctl.h": {},
				"ipc.h": {}, "kd.h": {}, "limits.h": {}, "link.h": {},
				"msg.h": {}, "ptrace.h": {}, "sem.h": {}, "shm.h": {},
				"soundcard.h": {}, "statfs.h": {}, "termios.h": {}, "vt.h": {},
			}

			for _, glob := range [][2]string{
				{"libc-top-half/musl/arch/generic/bits/*.h", "bits"},
				{"libc-bottom-half/headers/public/*.h", ""},
				{"libc-bottom-half/headers/public/wasi/*.h", "wasi"},
				{"libc-top-half/musl/arch/generic/bits/*.h", "bits"},
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
					if strings.Contains(match, "/arch/generic/") {
						if _, ok := omitGenericBits[name]; ok {
							continue
						}
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

			// Upstream makes wasi/version.h with CMake configure_file.
			if err := generateWasiVersionHeader(includeDir, preview); err != nil {
				return err
			}

			return nil
		},
		cflags: func(target, headerPath string) []string {
			libcDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/wasi-libc")
			flags := []string{
				"-Werror",
				"-Wall",
				"-std=gnu11",
				"-nostdlibinc",
				"-mnontrapping-fptoint", "-msign-ext", "-mbulk-memory",
				"-Wno-null-pointer-arithmetic", "-Wno-unused-parameter", "-Wno-sign-compare", "-Wno-unused-variable", "-Wno-unused-function", "-Wno-ignored-attributes", "-Wno-missing-braces", "-Wno-ignored-pragmas", "-Wno-unused-but-set-variable", "-Wno-unknown-warning-option",
				"-Wno-unused-command-line-argument", "-Wno-parentheses", "-Wno-shift-op-parentheses", "-Wno-bitwise-op-parentheses", "-Wno-logical-op-parentheses", "-Wno-string-plus-int", "-Wno-dangling-else", "-Wno-unknown-pragmas",
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
			if preview == "3" {
				// Cooperative threads, see ENABLE_COOP_THREADS in the
				// wasi-libc CMakeLists.txt.
				flags = append(flags, "-mthread-model", "posix", "-pthread", "-ftls-model=local-exec",
					"-I"+libcDir+"/libc-bottom-half/cloudlibc/src",
					"-I"+libcDir+"/libc-bottom-half/headers/private")
			} else {
				flags = append(flags, "-mthread-model", "single")
			}
			if preview != "1" {
				// wasi-libc needs -fdefer-ts. See CMakeLists.txt in wasi-libc.
				flags = append(flags, "-fdefer-ts")
			}
			return flags
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

			// wasip1 and wasip2 have socket sources with the same symbol names.
			// Compile only one set for each preview.
			cloudlibcSockSrcExclude := []string{}
			// wasip3_tls.c needs libcall thread context support, which stores
			// the stack pointer in a component context slot. Only the wasip3
			// target uses it.
			bottomHalfSrcExclude := []string{}
			if preview != "3" {
				bottomHalfSrcExclude = append(bottomHalfSrcExclude, "wasip3_tls.c")
			}
			wasip3Sources := []string{"wasip3.c", "wasip3_block_on.c", "wasip3_stdio.c"}
			wasip2Sources := []string{"wasip2.c", "wasip2_stdio.c"}
			switch preview {
			case "1":
				bottomHalfSrcExclude = append(bottomHalfSrcExclude, wasip2Sources...)
				bottomHalfSrcExclude = append(bottomHalfSrcExclude, wasip3Sources...)
				bottomHalfSrcExclude = append(bottomHalfSrcExclude,
					"accept.c", "bind.c", "connect.c", "descriptor_table.c",
					"file.c", "file_utils.c", "getsockpeername.c", "listen.c",
					"netdb.c", "recv.c", "send.c", "shutdown.c", "socket.c",
					"sockets_utils.c", "sockopt.c", "tcp.c", "udp.c",
				)
			case "2", "3":
				bottomHalfSrcExclude = append(bottomHalfSrcExclude, "accept-wasip1.c")
				if preview == "2" {
					bottomHalfSrcExclude = append(bottomHalfSrcExclude, wasip3Sources...)
				} else {
					bottomHalfSrcExclude = append(bottomHalfSrcExclude, wasip2Sources...)
				}
				cloudlibcSockSrcExclude = append(cloudlibcSockSrcExclude,
					"getsockopt.c", "recv.c", "send.c", "shutdown.c")
			default:
				return nil, fmt.Errorf("wasi-libc: unsupported WASI preview %q", preview)
			}

			// See: LIBC_TOP_HALF_MUSL_SOURCES in the Makefile
			globs := []filePattern{
				// Top half: mostly musl sources.
				{glob: "libc-top-half/sources/*.c"},
				{glob: "libc-top-half/musl/src/conf/*.c"},
				{glob: "libc-top-half/musl/src/internal/*.c", exclude: []string{
					"procfdname.c", "syscall.c", "syscall_ret.c", "vdso.c", "version.c",
					"emulate_wait4.c", // wait4/waitid emulation, not relevant/available on WASI
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
				{glob: "libc-bottom-half/cloudlibc/src/libc/sys/*/*.c", exclude: cloudlibcSockSrcExclude},
				{glob: "libc-bottom-half/sources/math/*.c"},
				{glob: "libc-bottom-half/sources/*.c", exclude: bottomHalfSrcExclude},
			}

			var threadSources []string
			stdioLockSources := []string{"__lockfile.c", "flockfile.c", "funlockfile.c", "ftrylockfile.c"}
			if preview == "3" {
				for _, name := range []string{
					"__lock.c", "__timedwait.c", "__wait.c", "default_attr.c", "lock_ptc.c", "mtx_destroy.c",
					"mtx_init.c", "mtx_lock.c", "mtx_timedlock.c", "mtx_trylock.c", "pthread_attr_destroy.c",
					"pthread_attr_get.c", "pthread_attr_init.c", "pthread_attr_setdetachstate.c",
					"pthread_attr_setguardsize.c", "pthread_attr_setschedparam.c", "pthread_attr_setstack.c",
					"pthread_attr_setstacksize.c", "pthread_barrier_destroy.c", "pthread_barrier_init.c",
					"pthread_barrier_wait.c", "pthread_barrierattr_destroy.c", "pthread_barrierattr_init.c",
					"pthread_barrierattr_setpshared.c", "pthread_cancel.c", "pthread_cleanup_push.c",
					"pthread_cond_broadcast.c", "pthread_cond_destroy.c", "pthread_cond_init.c",
					"pthread_cond_signal.c", "pthread_cond_timedwait.c", "pthread_cond_wait.c",
					"pthread_condattr_destroy.c", "pthread_condattr_init.c", "pthread_condattr_setclock.c",
					"pthread_condattr_setpshared.c", "pthread_equal.c", "pthread_getattr_np.c",
					"pthread_getcpuclockid.c", "pthread_getschedparam.c", "pthread_getspecific.c",
					"pthread_key_create.c", "pthread_mutex_consistent.c", "pthread_mutex_destroy.c",
					"pthread_mutex_getprioceiling.c", "pthread_mutex_init.c", "pthread_mutex_lock.c",
					"pthread_mutex_setprioceiling.c", "pthread_mutex_timedlock.c", "pthread_mutex_trylock.c",
					"pthread_mutex_unlock.c", "pthread_mutexattr_destroy.c", "pthread_mutexattr_init.c",
					"pthread_mutexattr_setprotocol.c", "pthread_mutexattr_setpshared.c",
					"pthread_mutexattr_setrobust.c", "pthread_mutexattr_settype.c", "pthread_once.c",
					"pthread_rwlock_destroy.c", "pthread_rwlock_init.c", "pthread_rwlock_rdlock.c",
					"pthread_rwlock_timedrdlock.c", "pthread_rwlock_timedwrlock.c",
					"pthread_rwlock_tryrdlock.c", "pthread_rwlock_trywrlock.c", "pthread_rwlock_unlock.c",
					"pthread_rwlock_wrlock.c", "pthread_rwlockattr_destroy.c", "pthread_rwlockattr_init.c",
					"pthread_rwlockattr_setpshared.c", "pthread_self.c", "pthread_setcancelstate.c",
					"pthread_setcanceltype.c", "pthread_setconcurrency.c", "pthread_setschedparam.c",
					"pthread_setschedprio.c", "pthread_setspecific.c", "pthread_spin_destroy.c",
					"pthread_spin_init.c", "pthread_spin_lock.c", "pthread_spin_trylock.c",
					"pthread_spin_unlock.c", "pthread_testcancel.c", "sem_destroy.c", "sem_getvalue.c",
					"sem_init.c", "sem_open.c", "sem_post.c", "sem_timedwait.c", "sem_trywait.c",
					"sem_wait.c", "thrd_join.c", "thrd_yield.c",
				} {
					threadSources = append(threadSources, "libc-top-half/musl/src/thread/common/"+name)
				}
				globs = append(globs,
					filePattern{glob: "libc-top-half/musl/src/thread/coop-threads/*.c", exclude: []string{"__lock.c"}},
				)
			} else {
				for i := range globs {
					if globs[i].glob == "libc-top-half/musl/src/stdio/*.c" {
						globs[i].exclude = append(globs[i].exclude, stdioLockSources...)
					}
				}
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
				"libc-top-half/musl/src/unistd/gethostname.c",
				"libc-top-half/musl/src/stat/futimesat.c",
				"libc-top-half/musl/src/legacy/getpagesize.c",
				"libc-top-half/musl/src/thread/common/thrd_sleep.c",
				"libc-top-half/musl/src/misc/realpath.c",
				"libc-top-half/musl/src/network/inet_addr.c",
				"libc-top-half/musl/src/network/inet_legacy.c",
				"libc-top-half/musl/src/network/inet_ntoa.c",
			}
			// See lib/wasi-libc-lockfile-stub.c.
			sources = append(sources, "../wasi-libc-lockfile-stub.c")
			sources = append(sources, threadSources...)
			if preview == "3" {
				sources = append(sources,
					"libc-top-half/musl/src/env/__init_tls.c",
					"libc-top-half/musl/src/thread/coop-threads/__wasi_coop_thread_start.s",
					"libc-bottom-half/sources/__wasm_task_hook.S",
				)
			}
			if preview == "2" || preview == "3" {
				// See lib/wasi-libc-wasip2-stub.c and lib/wasi-libc-wasip3-stub.c.
				sources = append(sources, "../wasi-libc-wasip"+preview+"-stub.c")
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
}
