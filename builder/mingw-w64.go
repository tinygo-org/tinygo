package builder

import (
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
	sourceDir: func() string { return "" }, // unused
	cflags: func(target, headerPath string) []string {
		// No flags necessary because there are no files to compile.
		return nil
	},
	librarySources: func(target string) []string {
		// We only use the UCRT DLL file. No source files necessary.
		return nil
	},
}

// makeMinGWExtraLibs returns a slice of jobs to import the correct .dll
// libraries. This is done by converting input .def files to .lib files which
// can then be linked as usual.
//
// TODO: cache the result. At the moment, it costs a few hundred milliseconds to
// compile these files.
func makeMinGWExtraLibs(tmpdir string) []*compileJob {
	var jobs []*compileJob
	root := goenv.Get("TINYGOROOT")
	// Normally all the api-ms-win-crt-*.def files are all compiled to a single
	// .lib file. But to simplify things, we're going to leave them as separate
	// files.
	for _, name := range []string{
		"kernel32.def.in",
		"api-ms-win-crt-conio-l1-1-0.def",
		"api-ms-win-crt-convert-l1-1-0.def",
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
				if strings.HasSuffix(inpath, ".in") {
					// .in files need to be preprocessed by a preprocessor (-E)
					// first.
					defpath = outpath + ".def"
					err := runCCompiler("-E", "-x", "c", "-Wp,-w", "-P", "-DDEF_X64", "-o", defpath, inpath, "-I"+goenv.Get("TINYGOROOT")+"/lib/mingw-w64/mingw-w64-crt/def-include/")
					if err != nil {
						return err
					}
				}
				return link("ld.lld", "-m", "i386pep", "-o", outpath, defpath)
			},
		}
		jobs = append(jobs, job)
	}
	return jobs
}
