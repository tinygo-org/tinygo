package builder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/goenv"
)

var MinGW = Library{
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
		return []string{
			"-nostdlibinc",
			"-isystem", mingwDir + "/mingw-w64-headers/crt",
			"-I", mingwDir + "/mingw-w64-headers/defaults/include",
			"-I" + headerPath,
		}
	},
	librarySources: func(target string) ([]string, error) {
		// These files are needed so that printf and the like are supported.
		sources := []string{
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
	// Normally all the api-ms-win-crt-*.def files are all compiled to a single
	// .lib file. But to simplify things, we're going to leave them as separate
	// files.
	for _, name := range []string{
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
	} {
		outpath := filepath.Join(tmpdir, filepath.Base(name)+".lib")
		inpath := filepath.Join(root, "lib/mingw-w64/mingw-w64-crt/lib-common/"+name)
		job := &compileJob{
			description: "create lib file " + inpath,
			result:      outpath,
			run: func(job *compileJob) error {
				defpath := inpath
				var archDef, emulation string
				switch goarch {
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
