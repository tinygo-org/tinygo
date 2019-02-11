// +build !byollvm

package main

// This file provides a Clang() function that always runs an external command.
// It is provided for when tinygo is built without linking to libclang.

import (
	"os"
	"os/exec"
)

// CCompiler invokes a C compiler with the given arguments.
//
// This version always runs the compiler as an external command.
func CCompiler(dir, command, path string, flags ...string) error {
	cmd := exec.Command(command, append(flags, path)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	return cmd.Run()
}
