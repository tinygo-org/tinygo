package builder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

// Extra declarations for libSystem exports absent from macos-minimal-sdk.
// See https://github.com/tinygo-org/macos-minimal-sdk for the SDK sources.
var darwinExtraLibSystemSymbols = []string{
	// The BSD socket API, which the host netdev in src/net reaches through
	// the standard library syscall package.
	"accept",
	"bind",
	"connect",
	"getpeername",
	"getsockname",
	"getsockopt",
	"listen",
	"recvfrom",
	"recvmsg",
	"sendmsg",
	"sendto",
	"setsockopt",
	"shutdown",
	"socket",
	"socketpair",

	// addchdir_np requires macOS 10.15 or later.
	"posix_spawn",
	"posix_spawn_file_actions_addchdir_np",
	"posix_spawn_file_actions_addclose",
	"posix_spawn_file_actions_adddup2",
	"posix_spawn_file_actions_addopen",
	"posix_spawn_file_actions_destroy",
	"posix_spawn_file_actions_init",
	"posix_spawnattr_destroy",
	"posix_spawnattr_init",
	"posix_spawnattr_setflags",
	"posix_spawnattr_setpgroup",
	"posix_spawnattr_setsigmask",
}

// Create a job that builds a Darwin libSystem.dylib stub library. This library
// contains all the symbols needed so that we can link against it, but it
// doesn't contain any real symbol implementations.
func makeDarwinLibSystemJob(config *compileopts.Config, tmpdir string) *compileJob {
	return &compileJob{
		description: "compile Darwin libSystem.dylib",
		run: func(job *compileJob) (err error) {
			arch, _, _ := strings.Cut(config.Triple(), "-")
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

			// Compile the extra stubs into a second object file, so that the
			// generated one stays as it is.
			extrapath := filepath.Join(tmpdir, "libSystem-extra.s")
			extraobjpath := filepath.Join(tmpdir, "libSystem-extra.o")
			var extra strings.Builder
			extra.WriteString("// Stubs for symbols exported by libSystem but not declared in lib/macos-minimal-sdk.\n")
			for _, symbol := range darwinExtraLibSystemSymbols {
				extra.WriteString("\n.global _" + symbol + "\n_" + symbol + ":\n")
			}
			if err := os.WriteFile(extrapath, []byte(extra.String()), 0o666); err != nil {
				return err
			}
			flags = []string{
				"-nostdlib",
				"--target=" + config.Triple(),
				"-c",
				"-o", extraobjpath,
				extrapath,
			}
			if config.Options.PrintCommands != nil {
				config.Options.PrintCommands("clang", flags...)
			}
			err = runCCompiler(flags...)
			if err != nil {
				return err
			}

			// Link object files to dynamic library.
			platformVersion := strings.TrimPrefix(strings.Split(config.Triple(), "-")[2], "macosx")
			flags = []string{
				"-flavor", "darwin",
				"-demangle",
				"-dynamic",
				"-dylib",
				"-arch", arch,
				"-platform_version", "macos", platformVersion, platformVersion,
				"-install_name", "/usr/lib/libSystem.B.dylib",
				"-o", job.result,
				objpath,
				extraobjpath,
			}
			if config.Options.PrintCommands != nil {
				config.Options.PrintCommands("ld.lld", flags...)
			}
			return link("ld.lld", flags...)
		},
	}
}
