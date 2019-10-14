// +build !byollvm

package main

// This file provides a Link() function that always runs an external command. It
// is provided for when tinygo is built without linking to liblld.

import (
	"os"
	"os/exec"

	"github.com/tinygo-org/tinygo/goenv"
)

// Link invokes a linker with the given name and arguments.
//
// This version always runs the linker as an external command.
func Link(linker string, flags ...string) error {
	if cmdNames, ok := commands[linker]; ok {
		return execCommand(cmdNames, flags...)
	}
	cmd := exec.Command(linker, flags...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = goenv.Get("TINYGOROOT")
	return cmd.Run()
}
