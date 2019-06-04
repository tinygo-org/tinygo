package main

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Commands used by the compilation process might have different file names
// across operating systems and distributions.
var commands = map[string][]string{
	"clang":   {"clang-8"},
	"ld.lld":  {"ld.lld-8", "ld.lld"},
	"wasm-ld": {"wasm-ld-8", "wasm-ld"},
}

func init() {
	// Add the path to a Homebrew-installed LLVM 8 for ease of use (no need to
	// manually set $PATH).
	if runtime.GOOS == "darwin" {
		commands["clang"] = append(commands["clang"], "/usr/local/opt/llvm/bin/clang-8")
		commands["ld.lld"] = append(commands["ld.lld"], "/usr/local/opt/llvm/bin/ld.lld")
		commands["wasm-ld"] = append(commands["wasm-ld"], "/usr/local/opt/llvm/bin/wasm-ld")
	}
}

func execCommand(cmdNames []string, args ...string) error {
	for _, cmdName := range cmdNames {
		cmd := exec.Command(cmdName, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			if err, ok := err.(*exec.Error); ok && err.Err == exec.ErrNotFound {
				// this command was not found, try the next
				continue
			}
			return err
		}
		return nil
	}
	return errors.New("none of these commands were found in your $PATH: " + strings.Join(cmdNames, " "))
}
