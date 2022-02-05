package builder

import (
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"tinygo.org/x/go-llvm"
)

// Create a job that builds a Darwin libSystem.dylib stub library. This library
// contains all the symbols needed so that we can link against it, but it
// doesn't contain any real symbol implementations.
func makeDarwinLibSystemJob(config *compileopts.Config, tmpdir string) *compileJob {
	return &compileJob{
		description: "compile Darwin libSystem.dylib",
		run: func(job *compileJob) (err error) {
			arch := strings.Split(config.Triple(), "-")[0]
			job.result = filepath.Join(tmpdir, "libSystem.dylib")
			objpath := filepath.Join(tmpdir, "libSystem.o")
			inpath := filepath.Join(goenv.Get("TINYGOROOT"), "lib/macos-minimal-sdk/src", arch, "libSystem.s")

			// Compile assembly file to object file.
			flags := []string{
				"-nostdlib",
				"--target=" + config.Triple(),
				"-c",
				"-o", objpath,
				inpath,
			}
			if config.Options.PrintCommands != nil {
				config.Options.PrintCommands("clang", flags...)
			}
			err = runCCompiler(flags...)
			if err != nil {
				return err
			}

			// Link object file to dynamic library.
			platformVersion := strings.TrimPrefix(strings.Split(config.Triple(), "-")[2], "macosx")
			flavor := "darwin"
			if strings.Split(llvm.Version, ".")[0] < "13" {
				flavor = "darwinnew" // needed on LLVM 12 and below
			}
			flags = []string{
				"-flavor", flavor,
				"-demangle",
				"-dynamic",
				"-dylib",
				"-arch", arch,
				"-platform_version", "macos", platformVersion, platformVersion,
				"-install_name", "/usr/lib/libSystem.B.dylib",
				"-o", job.result,
				objpath,
			}
			if config.Options.PrintCommands != nil {
				config.Options.PrintCommands("ld.lld", flags...)
			}
			return link("ld.lld", flags...)
		},
	}
}
