package builder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

var libMinGW = Library{
	name: "mingw-w64",
	makeHeaders: func(target, includeDir string) error {
		// copy _mingw.h
		srcDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib", "mingw-w64")
		outf, err := os.Create(includeDir + "/_mingw.h")
		if err != nil {
			return err
		}
		defer outf.Close()
		inf, err := os.Open(srcDir + "/mingw-w64-headers/crt/_mingw.h.in")
		if err != nil {
			return err
		}
		_, err = io.Copy(outf, inf)
		return err
	},
	sourceDir: func() string { return filepath.Join(goenv.Get("TINYGOROOT"), "lib/mingw-w64") },
	cflags: func(target, headerPath string) []string {
		mingwDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/mingw-w64")
		flags := []string{
			"-nostdlibinc",
			"-isystem", mingwDir + "/mingw-w64-crt/include",
			"-isystem", mingwDir + "/mingw-w64-headers/crt",
			"-isystem", mingwDir + "/mingw-w64-headers/include",
			"-I", mingwDir + "/mingw-w64-headers/defaults/include",
			"-I" + headerPath,
		}
		if strings.Split(target, "-")[0] == "i386" {
			flags = append(flags,
				"-D__MSVCRT_VERSION__=0x700", // Microsoft Visual C++ .NET 2002
				"-D_WIN32_WINNT=0x0501",      // target Windows XP
				"-D_CRTBLD",
				"-Wno-pragma-pack",
			)
		}
		return flags
	},
	librarySources: func(target string, _ bool) ([]string, error) {
		// These files are needed so that printf and the like are supported.
		var sources []string
		if strings.Split(target, "-")[0] == "i386" {
			// Old 32-bit x86 systems use msvcrt.dll.
			sources = []string{
				"mingw-w64-crt/crt/pseudo-reloc.c",
				"mingw-w64-crt/gdtoa/dmisc.c",
				"mingw-w64-crt/gdtoa/gdtoa.c",
				"mingw-w64-crt/gdtoa/gmisc.c",
				"mingw-w64-crt/gdtoa/misc.c",
				"mingw-w64-crt/math/x86/exp2.S",
				"mingw-w64-crt/math/x86/trunc.S",
				"mingw-w64-crt/misc/___mb_cur_max_func.c",
				"mingw-w64-crt/misc/lc_locale_func.c",
				"mingw-w64-crt/misc/mbrtowc.c",
				"mingw-w64-crt/misc/strnlen.c",
				"mingw-w64-crt/misc/wcrtomb.c",
				"mingw-w64-crt/misc/wcsnlen.c",
				"mingw-w64-crt/stdio/acrt_iob_func.c",
				"mingw-w64-crt/stdio/mingw_lock.c",
				"mingw-w64-crt/stdio/mingw_pformat.c",
				"mingw-w64-crt/stdio/mingw_vfprintf.c",
				"mingw-w64-crt/stdio/mingw_vsnprintf.c",
			}
		} else {
			// Anything somewhat modern (amd64, arm64) uses UCRT.
			sources = []string{
				"mingw-w64-crt/stdio/ucrt_fprintf.c",
				"mingw-w64-crt/stdio/ucrt_fwprintf.c",
				"mingw-w64-crt/stdio/ucrt_printf.c",
				"mingw-w64-crt/stdio/ucrt_snprintf.c",
				"mingw-w64-crt/stdio/ucrt_sprintf.c",
				"mingw-w64-crt/stdio/ucrt_vfprintf.c",
				"mingw-w64-crt/stdio/ucrt_vprintf.c",
				"mingw-w64-crt/stdio/ucrt_vsnprintf.c",
				"mingw-w64-crt/stdio/ucrt_vsprintf.c",
			}
		}
		return sources, nil
	},
}

// makeMinGWExtraLibs returns a slice of jobs to import the correct .dll
// libraries. This is done by converting input .def files to .lib files which
// can then be linked as usual.
//
// TODO: cache the result. At the moment, it costs a few hundred milliseconds to
// compile these files.
func makeMinGWExtraLibs(tmpdir, goarch string) []*compileJob {
	var jobs []*compileJob
	root := goenv.Get("TINYGOROOT")
	var libs []string
	if goarch == "386" {
		libs = []string{
			// x86 uses msvcrt.dll instead of UCRT for compatibility with old
			// Windows versions.
			"advapi32.def.in",
			"kernel32.def.in",
			"msvcrt.def.in",
		}
	} else {
		// Use the modernized UCRT on new systems.
		// Normally all the api-ms-win-crt-*.def files are all compiled to a
		// single .lib file. But to simplify things, we're going to leave them
		// as separate files.
		libs = []string{
			"advapi32.def.in",
			"kernel32.def.in",
			"api-ms-win-crt-conio-l1-1-0.def",
			"api-ms-win-crt-convert-l1-1-0.def.in",
			"api-ms-win-crt-environment-l1-1-0.def",
			"api-ms-win-crt-filesystem-l1-1-0.def",
			"api-ms-win-crt-heap-l1-1-0.def",
			"api-ms-win-crt-locale-l1-1-0.def",
			"api-ms-win-crt-math-l1-1-0.def.in",
			"api-ms-win-crt-multibyte-l1-1-0.def",
			"api-ms-win-crt-private-l1-1-0.def.in",
			"api-ms-win-crt-process-l1-1-0.def",
			"api-ms-win-crt-runtime-l1-1-0.def.in",
			"api-ms-win-crt-stdio-l1-1-0.def",
			"api-ms-win-crt-string-l1-1-0.def",
			"api-ms-win-crt-time-l1-1-0.def",
			"api-ms-win-crt-utility-l1-1-0.def",
		}
	}
	for _, name := range libs {
		outpath := filepath.Join(tmpdir, filepath.Base(name)+".lib")
		inpath := filepath.Join(root, "lib/mingw-w64/mingw-w64-crt/lib-common/"+name)
		job := &compileJob{
			description: "create lib file " + inpath,
			result:      outpath,
			run: func(job *compileJob) error {
				defpath := inpath
				var archDef, emulation string
				switch goarch {
				case "386":
					archDef = "-DDEF_I386"
					emulation = "i386pe"
				case "amd64":
					archDef = "-DDEF_X64"
					emulation = "i386pep"
				case "arm64":
					archDef = "-DDEF_ARM64"
					emulation = "arm64pe"
				default:
					return fmt.Errorf("unsupported architecture for mingw-w64: %s", goarch)
				}
				if strings.HasSuffix(inpath, ".in") {
					// .in files need to be preprocessed by a preprocessor (-E)
					// first.
					defpath = outpath + ".def"
					err := runCCompiler("-E", "-x", "c", "-Wp,-w", "-P", archDef, "-DDATA", "-o", defpath, inpath, "-I"+goenv.Get("TINYGOROOT")+"/lib/mingw-w64/mingw-w64-crt/def-include/")
					if err != nil {
						return err
					}
				}
				return link("ld.lld", "-m", emulation, "-o", outpath, defpath)
			},
		}
		jobs = append(jobs, job)
	}
	return jobs
}
