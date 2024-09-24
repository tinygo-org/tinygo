package builder

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

var libMusl = Library{
	name: "musl",
	makeHeaders: func(target, includeDir string) error {
		bits := filepath.Join(includeDir, "bits")
		err := os.Mkdir(bits, 0777)
		if err != nil {
			return err
		}

		arch := compileopts.MuslArchitecture(target)
		muslDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib", "musl")

		// Create the file alltypes.h.
		f, err := os.Create(filepath.Join(bits, "alltypes.h"))
		if err != nil {
			return err
		}
		infiles := []string{
			filepath.Join(muslDir, "arch", arch, "bits", "alltypes.h.in"),
			filepath.Join(muslDir, "include", "alltypes.h.in"),
		}
		for _, infile := range infiles {
			data, err := os.ReadFile(infile)
			if err != nil {
				return err
			}
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "TYPEDEF ") {
					matches := regexp.MustCompile(`TYPEDEF (.*) ([^ ]*);`).FindStringSubmatch(line)
					value := matches[1]
					name := matches[2]
					line = fmt.Sprintf("#if defined(__NEED_%s) && !defined(__DEFINED_%s)\ntypedef %s %s;\n#define __DEFINED_%s\n#endif\n", name, name, value, name, name)
				}
				if strings.HasPrefix(line, "STRUCT ") {
					matches := regexp.MustCompile(`STRUCT * ([^ ]*) (.*);`).FindStringSubmatch(line)
					name := matches[1]
					value := matches[2]
					line = fmt.Sprintf("#if defined(__NEED_struct_%s) && !defined(__DEFINED_struct_%s)\nstruct %s %s;\n#define __DEFINED_struct_%s\n#endif\n", name, name, name, value, name)
				}
				f.WriteString(line + "\n")
			}
		}
		f.Close()

		// Create the file syscall.h.
		f, err = os.Create(filepath.Join(bits, "syscall.h"))
		if err != nil {
			return err
		}
		data, err := os.ReadFile(filepath.Join(muslDir, "arch", arch, "bits", "syscall.h.in"))
		if err != nil {
			return err
		}
		_, err = f.Write(bytes.ReplaceAll(data, []byte("__NR_"), []byte("SYS_")))
		if err != nil {
			return err
		}
		f.Close()

		return nil
	},
	cflags: func(target, headerPath string) []string {
		arch := compileopts.MuslArchitecture(target)
		muslDir := filepath.Join(goenv.Get("TINYGOROOT"), "lib/musl")
		cflags := []string{
			"-std=c99",            // same as in musl
			"-D_XOPEN_SOURCE=700", // same as in musl
			// Musl triggers some warnings and we don't want to show any
			// warnings while compiling (only errors or silence), so disable
			// specific warnings that are triggered in musl.
			"-Werror",
			"-Wno-logical-op-parentheses",
			"-Wno-bitwise-op-parentheses",
			"-Wno-shift-op-parentheses",
			"-Wno-ignored-attributes",
			"-Wno-string-plus-int",
			"-Wno-ignored-pragmas",
			"-Wno-tautological-constant-out-of-range-compare",
			"-Wno-deprecated-non-prototype",
			"-Wno-format",
			"-Wno-parentheses",
			"-Qunused-arguments",
			// Select include dirs. Don't include standard library includes
			// (that would introduce host dependencies and other complications),
			// but do include all the include directories expected by musl.
			"-nostdlibinc",
			"-I" + muslDir + "/arch/" + arch,
			"-I" + muslDir + "/arch/generic",
			"-I" + muslDir + "/src/include",
			"-I" + muslDir + "/src/internal",
			"-I" + headerPath,
			"-I" + muslDir + "/include",
			"-fno-stack-protector",
		}
		return cflags
	},
	sourceDir: func() string { return filepath.Join(goenv.Get("TINYGOROOT"), "lib/musl/src") },
	librarySources: func(target string) ([]string, error) {
		arch := compileopts.MuslArchitecture(target)
		globs := []string{
			"ctype/*.c",
			"env/*.c",
			"errno/*.c",
			"exit/*.c",
			"fcntl/*.c",
			"internal/defsysinfo.c",
			"internal/intscan.c",
			"internal/libc.c",
			"internal/shgetc.c",
			"internal/syscall_ret.c",
			"internal/vdso.c",
			"legacy/*.c",
			"locale/*.c",
			"linux/*.c",
			"locale/*.c",
			"malloc/*.c",
			"malloc/mallocng/*.c",
			"mman/*.c",
			"math/*.c",
			"misc/*.c",
			"multibyte/*.c",
			"sched/*.c",
			"signal/*.c",
			"stdio/*.c",
			"stdlib/*.c",
			"string/*.c",
			"thread/" + arch + "/*.s",
			"thread/*.c",
			"time/*.c",
			"unistd/*.c",
		}
		if arch == "arm" {
			// These files need to be added to the start for some reason.
			globs = append([]string{"thread/arm/*.c"}, globs...)
		}

		var sources []string
		seenSources := map[string]struct{}{}
		basepath := goenv.Get("TINYGOROOT") + "/lib/musl/src/"
		for _, pattern := range globs {
			matches, err := filepath.Glob(basepath + pattern)
			if err != nil {
				// From the documentation:
				// > Glob ignores file system errors such as I/O errors reading
				// > directories. The only possible returned error is
				// > ErrBadPattern, when pattern is malformed.
				// So the only possible error is when the (statically defined)
				// pattern is wrong. In other words, a programming bug.
				return nil, fmt.Errorf("musl: could not glob source dirs: %w", err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("musl: did not find any files for pattern %#v", pattern)
			}
			for _, match := range matches {
				relpath, err := filepath.Rel(basepath, match)
				if err != nil {
					// Not sure if this is even possible.
					return nil, err
				}
				// Make sure architecture specific files override generic files.
				id := strings.ReplaceAll(relpath, "/"+arch+"/", "/")
				if _, ok := seenSources[id]; ok {
					// Already seen this file, skipping this (generic) file.
					continue
				}
				seenSources[id] = struct{}{}
				sources = append(sources, relpath)
			}
		}
		return sources, nil
	},
	crt1Source: "../crt/crt1.c", // lib/musl/crt/crt1.c
}
