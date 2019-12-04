// +build !byollvm

package builder

// This file provides a way to a C compiler as an external command. See also:
// clang-external.go

import (
	"os"
	"os/exec"
)

// runCCompiler invokes a C compiler with the given arguments.
//
// This version always runs the compiler as an external command.
func runCCompiler(command string, flags ...string) error {
	if cmdNames, ok := commands[command]; ok {
		return execCommand(cmdNames, flags...)
	}
	cmd := exec.Command(command, flags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
