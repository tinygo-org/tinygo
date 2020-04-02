package builder

import (
	"errors"
	"os"
	"os/exec"

	"github.com/tinygo-org/tinygo/goenv"
)

// runCCompiler invokes a C compiler with the given arguments.
func runCCompiler(command string, flags ...string) error {
	if hasBuiltinTools && command == "clang" {
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

	// Running some other compiler. Maybe it has been defined in the
	// commands map (unlikely).
	if cmdNames, ok := commands[command]; ok {
		return execCommand(cmdNames, flags...)
	}

	// Alternatively, run the compiler directly.
	cmd := exec.Command(command, flags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
	if cmdNames, ok := commands[linker]; ok {
		return execCommand(cmdNames, flags...)
	}

	cmd := exec.Command(linker, flags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = goenv.Get("TINYGOROOT")
	return cmd.Run()
}
