package builder

import (
	"errors"
	"os"
	"os/exec"

	"github.com/tinygo-org/tinygo/goenv"
)

// runCCompiler invokes a C compiler with the given arguments.
func runCCompiler(flags ...string) error {
	if hasBuiltinTools {
		// Compile this with the internal Clang compiler.
		headerPath := getClangHeaderPath(goenv.Get("TINYGOROOT"))
		if headerPath == "" {
			return errors.New("could not locate Clang headers")
		}
		flags = append(flags, "-I"+headerPath)
		cmd := exec.Command(os.Args[0], append([]string{"clang"}, flags...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Compile this with an external invocation of the Clang compiler.
	return execCommand("clang", flags...)
}

// link invokes a linker with the given name and flags.
func link(linker string, flags ...string) error {
	if hasBuiltinTools && (linker == "ld.lld" || linker == "wasm-ld") {
		// Run command with internal linker.
		cmd := exec.Command(os.Args[0], append([]string{linker}, flags...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Fall back to external command.
	if _, ok := commands[linker]; ok {
		return execCommand(linker, flags...)
	}

	cmd := exec.Command(linker, flags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = goenv.Get("TINYGOROOT")
	return cmd.Run()
}
