// +build !byollvm

package main

// This file provides a Link() function that always runs an external command. It
// is provided for when tinygo is built without linking to liblld.

import (
	"os"
	"os/exec"
)

// Link invokes a linker with the given name and arguments.
//
// This version always runs the linker as an external command.
func Link(linker string, flags ...string) error {
	cmd := exec.Command(linker, flags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
